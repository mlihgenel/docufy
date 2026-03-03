package cli

import (
	"strings"

	"github.com/mlihgenel/docufy/internal/converter"
	"github.com/mlihgenel/docufy/internal/ui"
)

func newCLIFFmpegProgress(label string) func(converter.ProgressInfo) {
	pb := ui.NewProgressBar(100, label)
	return func(info converter.ProgressInfo) {
		percent := int(info.Percent)
		if percent < 0 {
			percent = 0
		}
		if percent > 100 {
			percent = 100
		}
		pb.Update(percent)
	}
}

func progressLabelForConverter(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.Contains(lower, "video"):
		return "Video işleniyor"
	case strings.Contains(lower, "audio"):
		return "Ses işleniyor"
	default:
		return "Dönüştürülüyor"
	}
}
