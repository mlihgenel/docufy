package converter

import (
	"strings"
	"testing"
	"time"
)

func TestParseFFmpegProgressTime(t *testing.T) {
	got, ok := parseFFmpegProgressTime("00:01:30.500000")
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if got != 90*time.Second+500*time.Millisecond {
		t.Fatalf("unexpected duration: %s", got)
	}
}

func TestScanFFmpegProgress(t *testing.T) {
	input := strings.NewReader("out_time_ms=5000000\nprogress=continue\nout_time_ms=10000000\nprogress=end\n")
	var last ProgressInfo
	err := scanFFmpegProgress(input, 20*time.Second, func(info ProgressInfo) {
		last = info
	})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if last.Percent != 100 {
		t.Fatalf("expected final percent 100, got %.2f", last.Percent)
	}
}
