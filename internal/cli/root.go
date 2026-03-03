package cli

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mlihgenel/docufy/v2/internal/config"
	"github.com/spf13/cobra"

	// Converter modüllerini kaydet
	_ "github.com/mlihgenel/docufy/v2/internal/converter"
)

var (
	verbose      bool
	outputDir    string
	workers      int
	outputFormat string

	appVersion = "dev"
	appDate    = ""

	activeProjectConfig     *config.ProjectConfig
	activeProjectConfigPath string
)

// SetVersionInfo build-time version bilgisini ayarlar
func SetVersionInfo(version, date string) {
	if strings.TrimSpace(version) != "" {
		appVersion = version
	}
	appDate = strings.TrimSpace(date)
	if appDate == "" || appDate == "unknown" {
		appDate = time.Now().Format("2006-01-02 15:04:05")
	}
	rootCmd.Version = appVersion
	rootCmd.SetVersionTemplate(versionTemplate())
}

func versionTemplate() string {
	return fmt.Sprintf(
		"Docufy v%s\nTarih:  %s\nGo:     %s\nOS:     %s/%s\n",
		appVersion, appDate, runtime.Version(), runtime.GOOS, runtime.GOARCH,
	)
}

var rootCmd = &cobra.Command{
	Use:   "docufy",
	Short: "Docufy - yerel dosya format donusturucu",
	Long: `Docufy — Dosyalarınızı yerel ortamda güvenli bir şekilde dönüştürün.

Belge, ses, görsel ve video dosyalarını internet'e yüklemeden, tamamen yerel
olarak farklı formatlara dönüştürmenizi sağlar.

Kullanim:
  docufy                  # Interaktif TUI (bolum bazli menu)
  docufy --help           # Tum komutlari goster
  docufy help <komut>     # Belirli komut yardimi

Desteklenen kategoriler:
  Belgeler:  MD, HTML, PDF, DOCX, TXT, ODT, RTF, CSV (+ CSV -> XLSX)
  Ses:       MP3, WAV, OGG, FLAC, AAC, M4A, WMA, OPUS, WEBM  (FFmpeg gerekir)
  Gorseller: PNG, JPEG, WEBP, BMP, GIF, TIFF, ICO  (WEBP yalnizca kaynak)
  Videolar:  MP4, MOV, MKV, AVI, WEBM, M4V, WMV, FLV, GIF  (FFmpeg gerekir)

Örnekler:
  docufy convert dosya.md --to pdf
  docufy convert muzik.mp3 --to wav
  docufy convert resim.png --to jpg --quality 90
  docufy convert klip.mp4 --to mp4 --preset story --resize-mode pad
  docufy convert klip.mp4 --to mp4 --profile social-story --strip-metadata
  docufy convert klip.mp4 --to gif --quality 80
  docufy batch ./belgeler --from md --to pdf
  docufy batch ./resimler --from webp --to jpg --profile archive-lossless --preserve-metadata
  docufy batch ./resimler --from jpg --to png --on-conflict versioned --retry 2 --report json
  docufy watch ./incoming --from webp --to jpg
  docufy pipeline run ./pipeline.json --profile social-story
  docufy video trim input.mp4 --start 00:00:05 --duration 10
  docufy video trim input.mp4 --mode remove --start 00:00:23 --duration 2
  docufy video trim input.mp4 --mode remove --ranges "00:00:05-00:00:08,00:00:20-00:00:25"
  docufy video trim input.mp4 --mode remove --ranges "5-8,20-25" --dry-run
  docufy help video
  docufy help formats
  docufy resize-presets
  docufy formats`,
	Version: appVersion,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		cfg, cfgPath, err := config.LoadProjectConfig(wd)
		if err != nil {
			return err
		}
		activeProjectConfig = cfg
		activeProjectConfigPath = cfgPath

		return applyRootDefaults(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Argümansız çalıştırıldığında interaktif mod başlat
		return RunInteractive()
	},
}

// Execute CLI'ı çalıştırır
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Detaylı çıktı modu")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "Çıktı dizini (varsayılan: kaynak dizin)")
	rootCmd.PersistentFlags().IntVarP(&workers, "workers", "w", runtime.NumCPU(), "Paralel worker sayısı (batch modunda)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output-format", OutputFormatText, "CLI çıktı formatı: text veya json")

	SetVersionInfo(appVersion, appDate)

	// Hata mesajlarını özelleştir
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		fmt.Fprintf(os.Stderr, "Hata: %s\n\n", err.Error())
		cmd.Usage()
		return err
	})
}
