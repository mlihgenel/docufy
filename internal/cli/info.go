package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/mlihgenel/fileconverter-cli/internal/converter"
	"github.com/mlihgenel/fileconverter-cli/internal/ui"
)

var infoCmd = &cobra.Command{
	Use:   "info <dosya>",
	Short: "Dosya hakkında detaylı bilgi göster",
	Long: `Bir dosyanın format, boyut, çözünürlük, codec ve metadata bilgilerini gösterir.

Örnekler:
  fileconverter-cli info foto.jpg
  fileconverter-cli info video.mp4
  fileconverter-cli info ses.mp3
  fileconverter-cli info belge.pdf
  fileconverter-cli info foto.jpg --output-format json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		info, err := converter.GetFileInfo(filePath)
		if err != nil {
			ui.PrintError(err.Error())
			return err
		}

		if isJSONOutput() {
			data, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		printFileInfo(info)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func printFileInfo(info converter.FileInfo) {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#10B981"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E2E8F0")).
		Width(16)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#334155")).
		Padding(1, 2).
		MarginTop(1)

	var lines []string

	// Başlık
	icon := categoryIcon(info.Category)
	lines = append(lines, headerStyle.Render(fmt.Sprintf("%s  %s", icon, info.FileName)))
	lines = append(lines, dimStyle.Render(strings.Repeat("─", 40)))

	// Temel bilgiler
	lines = append(lines, formatInfoLine(labelStyle, valueStyle, "Format", info.Format))
	lines = append(lines, formatInfoLine(labelStyle, valueStyle, "Kategori", categoryLabel(info.Category)))
	lines = append(lines, formatInfoLine(labelStyle, valueStyle, "Boyut", info.SizeText))

	// Görsel bilgileri
	if info.Resolution != "" {
		lines = append(lines, formatInfoLine(labelStyle, valueStyle, "Çözünürlük", info.Resolution))
	}

	// Video/Ses bilgileri
	if info.Duration != "" {
		lines = append(lines, formatInfoLine(labelStyle, valueStyle, "Süre", info.Duration))
	}
	if info.VideoCodec != "" {
		lines = append(lines, formatInfoLine(labelStyle, valueStyle, "Video Codec", info.VideoCodec))
	}
	if info.AudioCodec != "" {
		lines = append(lines, formatInfoLine(labelStyle, valueStyle, "Ses Codec", info.AudioCodec))
	}
	if info.Bitrate != "" {
		lines = append(lines, formatInfoLine(labelStyle, valueStyle, "Bitrate", info.Bitrate))
	}
	if info.FPS > 0 {
		lines = append(lines, formatInfoLine(labelStyle, valueStyle, "FPS", fmt.Sprintf("%.2f", info.FPS)))
	}
	if info.Channels > 0 {
		chLabel := fmt.Sprintf("%d", info.Channels)
		if info.Channels == 1 {
			chLabel = "Mono"
		} else if info.Channels == 2 {
			chLabel = "Stereo"
		}
		lines = append(lines, formatInfoLine(labelStyle, valueStyle, "Kanal", chLabel))
	}
	if info.SampleRate > 0 {
		lines = append(lines, formatInfoLine(labelStyle, valueStyle, "Örnekleme", fmt.Sprintf("%d Hz", info.SampleRate)))
	}

	fmt.Println(boxStyle.Render(strings.Join(lines, "\n")))
}

func formatInfoLine(labelStyle, valueStyle lipgloss.Style, label, value string) string {
	return labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}

func categoryIcon(category string) string {
	switch category {
	case "image":
		return "🖼️"
	case "video":
		return "🎬"
	case "audio":
		return "🎵"
	case "document":
		return "📄"
	default:
		return "📁"
	}
}

func categoryLabel(category string) string {
	switch category {
	case "image":
		return "Görsel"
	case "video":
		return "Video"
	case "audio":
		return "Ses"
	case "document":
		return "Belge"
	default:
		return "Diğer"
	}
}
