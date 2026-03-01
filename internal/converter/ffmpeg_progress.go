package converter

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func RunFFmpegWithProgress(ffmpegPath string, input string, args []string, opts Options, prefix string) error {
	parsedArgs := append([]string{}, args...)
	useProgress := opts.Progress != nil
	if useProgress {
		parsedArgs = append([]string{"-progress", "pipe:1", "-nostats"}, parsedArgs...)
	}

	cmd := exec.Command(ffmpegPath, parsedArgs...)

	if !useProgress {
		outputBytes, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s\n%s", prefix, err.Error(), string(outputBytes))
		}
		return nil
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("%s: stdout pipe oluşturulamadı: %w", prefix, err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("%s: stderr pipe oluşturulamadı: %w", prefix, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%s: ffmpeg başlatılamadı: %w", prefix, err)
	}

	totalDuration := probeMediaDuration(input)
	errCh := make(chan error, 1)
	go func() {
		errCh <- scanFFmpegProgress(stdout, totalDuration, opts.Progress)
	}()

	stderrBytes, readErr := io.ReadAll(stderr)
	waitErr := cmd.Wait()
	progressErr := <-errCh

	if progressErr != nil {
		return fmt.Errorf("%s: progress parse hatası: %w", prefix, progressErr)
	}
	if readErr != nil {
		return fmt.Errorf("%s: stderr okunamadı: %w", prefix, readErr)
	}
	if waitErr != nil {
		msg := strings.TrimSpace(string(stderrBytes))
		if msg == "" {
			return fmt.Errorf("%s: %s", prefix, waitErr.Error())
		}
		return fmt.Errorf("%s: %s\n%s", prefix, waitErr.Error(), msg)
	}

	if opts.Progress != nil {
		opts.Progress(ProgressInfo{
			Percent:      100,
			Current:      totalDuration,
			Total:        totalDuration,
			ETA:          0,
			CurrentLabel: "Tamamlandı",
		})
	}

	return nil
}

func scanFFmpegProgress(r io.Reader, total time.Duration, cb func(ProgressInfo)) error {
	if cb == nil {
		return nil
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var current time.Duration
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		switch key {
		case "out_time_ms", "out_time_us":
			if micros, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64); err == nil && micros >= 0 {
				current = time.Duration(micros) * time.Microsecond
				cb(buildFFmpegProgressInfo(current, total, "İşleniyor"))
			}
		case "out_time":
			if parsed, ok := parseFFmpegProgressTime(value); ok {
				current = parsed
				cb(buildFFmpegProgressInfo(current, total, "İşleniyor"))
			}
		case "progress":
			switch strings.TrimSpace(value) {
			case "end":
				cb(ProgressInfo{Percent: 100, Current: total, Total: total, ETA: 0, CurrentLabel: "Tamamlandı"})
			case "continue":
				if current > 0 {
					cb(buildFFmpegProgressInfo(current, total, "İşleniyor"))
				}
			}
		}
	}

	return scanner.Err()
}

func buildFFmpegProgressInfo(current time.Duration, total time.Duration, label string) ProgressInfo {
	info := ProgressInfo{
		Current:      current,
		Total:        total,
		CurrentLabel: label,
	}
	if total > 0 {
		percent := float64(current) / float64(total) * 100
		if percent < 0 {
			percent = 0
		}
		if percent > 99.5 {
			percent = 99.5
		}
		info.Percent = percent
		if current > 0 && total > current {
			remaining := total - current
			info.ETA = remaining
		}
	}
	return info
}

func parseFFmpegProgressTime(raw string) (time.Duration, bool) {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) != 3 {
		return 0, false
	}

	hours, err1 := strconv.ParseFloat(parts[0], 64)
	minutes, err2 := strconv.ParseFloat(parts[1], 64)
	seconds, err3 := strconv.ParseFloat(parts[2], 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, false
	}

	totalSeconds := hours*3600 + minutes*60 + seconds
	if totalSeconds < 0 {
		return 0, false
	}
	return time.Duration(totalSeconds * float64(time.Second)), true
}

func probeMediaDuration(input string) time.Duration {
	ffprobePath := findFFprobe()
	if ffprobePath == "" || strings.TrimSpace(input) == "" {
		return 0
	}

	out, err := exec.Command(ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	).Output()
	if err != nil {
		return 0
	}

	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds <= 0 {
		return 0
	}

	return time.Duration(seconds * float64(time.Second))
}
