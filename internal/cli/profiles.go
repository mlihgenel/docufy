package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mlihgenel/docufy/v2/internal/profile"
	"github.com/mlihgenel/docufy/v2/internal/ui"
)

var (
	profileCreateDescription  string
	profileCreateScope        string
	profileCreateQuality      int
	profileCreateOnConflict   string
	profileCreateRetry        int
	profileCreateRetryDelay   string
	profileCreateReport       string
	profileCreatePreset       string
	profileCreateResizeMode   string
	profileCreateWidth        float64
	profileCreateHeight       float64
	profileCreateUnit         string
	profileCreateDPI          float64
	profileCreateMetadataMode string
)

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Hazır ve kullanıcı tanımlı profilleri yönet",
	Long: `Profil, tekrar kullanabileceginiz varsayilan ayar setidir.
Profil kaynak/hedef format secmez; quality/retry/metadata/resize gibi flag degerlerini topluca uygular.`,
}

var profilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Kullanılabilir profilleri listeler",
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := profile.List()
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(items)
		}

		fmt.Println()
		fmt.Printf("  %s %sProfiller%s\n\n", "🧩", ui.Bold, ui.Reset)
		ui.PrintInfo("Profil: donusum komutlarinda varsayilan ayarlari hizli uygulamak icin kullanilir.")
		ui.PrintInfo("Not: Komutta verdiginiz flag degeri, profildeki ayni degerin ustune yazar.")
		fmt.Println()

		groups := groupProfilesByScope(items)
		for _, group := range groups {
			fmt.Printf("  • %sKapsam: %s%s\n", ui.Bold, group.Label, ui.Reset)
			headers := []string{"Ad", "Kaynak", "Kapsam", "Ne Yapar"}
			rows := make([][]string, 0, len(group.Items))
			for _, item := range group.Items {
				desc := item.Description
				if strings.TrimSpace(desc) == "" {
					desc = summarizeProfile(item)
				}
				rows = append(rows, []string{
					item.Name,
					item.Source,
					profile.ScopeLabel(item.Scope),
					desc,
				})
			}
			ui.PrintTable(headers, rows)
			fmt.Println()
		}
		return nil
	},
}

var profilesCreateCmd = &cobra.Command{
	Use:   "create [profil-adi]",
	Short: "Yeni bir kullanıcı profili oluşturur",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		def, err := buildProfileDefinitionFromInput(cmd, args)
		if err != nil {
			return err
		}

		path, err := profile.SaveUserProfile(def)
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(map[string]string{
				"status": "created",
				"name":   def.Name,
				"path":   path,
			})
		}

		ui.PrintSuccess(fmt.Sprintf("Profil olusturuldu: %s", def.Name))
		ui.PrintInfo(fmt.Sprintf("Dosya: %s", path))
		return nil
	},
}

func buildProfileDefinitionFromInput(cmd *cobra.Command, args []string) (profile.Definition, error) {
	prompter := newProfilePrompter(cmd)

	name := ""
	if len(args) > 0 {
		name = strings.TrimSpace(args[0])
	}
	var err error
	if name == "" {
		name, err = prompter.String("Profil adi", "")
		if err != nil {
			return profile.Definition{}, err
		}
	}
	if strings.TrimSpace(name) == "" {
		return profile.Definition{}, fmt.Errorf("profil adi gerekli")
	}

	description, err := firstString(prompter, profileCreateDescription, "Aciklama")
	if err != nil {
		return profile.Definition{}, err
	}
	scope, err := firstString(prompter, profileCreateScope, "scope (all/document/image/video/audio veya image,video)")
	if err != nil {
		return profile.Definition{}, err
	}
	onConflict, err := firstString(prompter, profileCreateOnConflict, "on-conflict")
	if err != nil {
		return profile.Definition{}, err
	}
	metadataMode, err := firstString(prompter, profileCreateMetadataMode, "metadata_mode")
	if err != nil {
		return profile.Definition{}, err
	}
	resizePreset, err := firstString(prompter, profileCreatePreset, "resize_preset")
	if err != nil {
		return profile.Definition{}, err
	}
	resizeMode, err := firstString(prompter, profileCreateResizeMode, "resize_mode")
	if err != nil {
		return profile.Definition{}, err
	}
	unit, err := firstString(prompter, profileCreateUnit, "unit")
	if err != nil {
		return profile.Definition{}, err
	}
	report, err := firstString(prompter, profileCreateReport, "report")
	if err != nil {
		return profile.Definition{}, err
	}
	retryDelayText, err := firstString(prompter, profileCreateRetryDelay, "retry_delay")
	if err != nil {
		return profile.Definition{}, err
	}

	def := profile.Definition{
		Name:         name,
		Description:  strings.TrimSpace(description),
		Scope:        strings.TrimSpace(scope),
		OnConflict:   strings.TrimSpace(onConflict),
		Report:       strings.TrimSpace(report),
		ResizePreset: strings.TrimSpace(resizePreset),
		ResizeMode:   strings.TrimSpace(resizeMode),
		Unit:         strings.TrimSpace(unit),
		MetadataMode: strings.TrimSpace(metadataMode),
	}

	if cmd.Flags().Changed("quality") {
		def.Quality = profile.IntPtr(profileCreateQuality)
	} else if value, ok, err := prompter.Int("quality", 0); err != nil {
		return profile.Definition{}, err
	} else if ok {
		def.Quality = profile.IntPtr(value)
	}

	if cmd.Flags().Changed("retry") {
		def.Retry = profile.IntPtr(profileCreateRetry)
	} else if value, ok, err := prompter.Int("retry", 0); err != nil {
		return profile.Definition{}, err
	} else if ok {
		def.Retry = profile.IntPtr(value)
	}

	if cmd.Flags().Changed("width") {
		def.Width = profile.FloatPtr(profileCreateWidth)
	} else if value, ok, err := prompter.Float("width", 0); err != nil {
		return profile.Definition{}, err
	} else if ok {
		def.Width = profile.FloatPtr(value)
	}

	if cmd.Flags().Changed("height") {
		def.Height = profile.FloatPtr(profileCreateHeight)
	} else if value, ok, err := prompter.Float("height", 0); err != nil {
		return profile.Definition{}, err
	} else if ok {
		def.Height = profile.FloatPtr(value)
	}

	if cmd.Flags().Changed("dpi") {
		def.DPI = profile.FloatPtr(profileCreateDPI)
	} else if value, ok, err := prompter.Float("dpi", 0); err != nil {
		return profile.Definition{}, err
	} else if ok {
		def.DPI = profile.FloatPtr(value)
	}

	if strings.TrimSpace(retryDelayText) != "" {
		parsed, err := time.ParseDuration(strings.TrimSpace(retryDelayText))
		if err != nil {
			return profile.Definition{}, fmt.Errorf("gecersiz retry_delay: %w", err)
		}
		def.RetryDelay = profile.DurationPtr(parsed)
	}

	return def, nil
}

type profilePrompter struct {
	reader  *bufio.Reader
	writer  io.Writer
	enabled bool
}

func newProfilePrompter(cmd *cobra.Command) profilePrompter {
	return profilePrompter{
		reader:  bufio.NewReader(cmd.InOrStdin()),
		writer:  cmd.OutOrStdout(),
		enabled: isTerminalPromptAllowed(cmd),
	}
}

func (p profilePrompter) String(label, defaultValue string) (string, error) {
	if !p.enabled {
		return defaultValue, nil
	}
	if _, err := fmt.Fprint(p.writer, buildPrompt(label, defaultValue)); err != nil {
		return "", err
	}
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue, nil
	}
	return line, nil
}

func (p profilePrompter) Int(label string, defaultValue int) (int, bool, error) {
	raw, err := p.String(label, "")
	if err != nil {
		return 0, false, err
	}
	if strings.TrimSpace(raw) == "" {
		return defaultValue, false, nil
	}
	value, err := parseInt(raw)
	if err != nil {
		return 0, false, fmt.Errorf("%s icin gecersiz sayi", label)
	}
	return value, true, nil
}

func (p profilePrompter) Float(label string, defaultValue float64) (float64, bool, error) {
	raw, err := p.String(label, "")
	if err != nil {
		return 0, false, err
	}
	if strings.TrimSpace(raw) == "" {
		return defaultValue, false, nil
	}
	value, err := parseFloat(raw)
	if err != nil {
		return 0, false, fmt.Errorf("%s icin gecersiz sayi", label)
	}
	return value, true, nil
}

func firstString(prompter profilePrompter, current, label string) (string, error) {
	if strings.TrimSpace(current) != "" {
		return current, nil
	}
	return prompter.String(label, "")
}

func buildPrompt(label, defaultValue string) string {
	if strings.TrimSpace(defaultValue) == "" {
		return fmt.Sprintf("%s: ", label)
	}
	return fmt.Sprintf("%s [%s]: ", label, defaultValue)
}

func isTerminalPromptAllowed(cmd *cobra.Command) bool {
	file, ok := cmd.InOrStdin().(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func summarizeProfile(def profile.Definition) string {
	parts := make([]string, 0, 6)
	parts = append(parts, "kapsam="+strings.ToLower(profile.ScopeLabel(def.Scope)))
	if def.ResizePreset != "" {
		parts = append(parts, "preset="+def.ResizePreset)
	}
	if def.Quality != nil {
		parts = append(parts, fmt.Sprintf("quality=%d", *def.Quality))
	}
	if def.OnConflict != "" {
		parts = append(parts, "conflict="+def.OnConflict)
	}
	if def.MetadataMode != "" {
		parts = append(parts, "metadata="+def.MetadataMode)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, ", ")
}

type profileScopeGroup struct {
	Key   string
	Label string
	Items []profile.Definition
}

func groupProfilesByScope(items []profile.Definition) []profileScopeGroup {
	grouped := make(map[string][]profile.Definition)
	keys := make([]string, 0)

	for _, item := range items {
		key := profile.NormalizeScope(item.Scope)
		if key == "" {
			key = "all"
		}
		if _, ok := grouped[key]; !ok {
			keys = append(keys, key)
		}
		grouped[key] = append(grouped[key], item)
	}

	sort.Slice(keys, func(i, j int) bool {
		ri := scopeGroupRank(keys[i])
		rj := scopeGroupRank(keys[j])
		if ri != rj {
			return ri < rj
		}
		return keys[i] < keys[j]
	})

	result := make([]profileScopeGroup, 0, len(keys))
	for _, key := range keys {
		scopeItems := grouped[key]
		sort.Slice(scopeItems, func(i, j int) bool {
			return scopeItems[i].Name < scopeItems[j].Name
		})
		result = append(result, profileScopeGroup{
			Key:   key,
			Label: profile.ScopeLabel(key),
			Items: scopeItems,
		})
	}
	return result
}

func scopeGroupRank(key string) int {
	switch key {
	case "image,video":
		return 0
	case "image":
		return 1
	case "video":
		return 2
	case "audio":
		return 3
	case "document":
		return 4
	case "all":
		return 9
	default:
		return 8
	}
}

func parseInt(raw string) (int, error) {
	var value int
	_, err := fmt.Sscanf(strings.TrimSpace(raw), "%d", &value)
	return value, err
}

func parseFloat(raw string) (float64, error) {
	var value float64
	_, err := fmt.Sscanf(strings.TrimSpace(raw), "%f", &value)
	return value, err
}

func init() {
	profilesCreateCmd.Flags().StringVar(&profileCreateDescription, "description", "", "Profil aciklamasi")
	profilesCreateCmd.Flags().StringVar(&profileCreateScope, "scope", "", "Profil kapsami: all, document, image, video, audio (veya image,video)")
	profilesCreateCmd.Flags().IntVar(&profileCreateQuality, "quality", 0, "Varsayilan kalite")
	profilesCreateCmd.Flags().StringVar(&profileCreateOnConflict, "on-conflict", "", "Varsayilan cakisma politikasi")
	profilesCreateCmd.Flags().IntVar(&profileCreateRetry, "retry", 0, "Varsayilan retry sayisi")
	profilesCreateCmd.Flags().StringVar(&profileCreateRetryDelay, "retry-delay", "", "Varsayilan retry suresi (orn: 1s)")
	profilesCreateCmd.Flags().StringVar(&profileCreateReport, "report", "", "Varsayilan rapor formati")
	profilesCreateCmd.Flags().StringVar(&profileCreatePreset, "preset", "", "Varsayilan resize preset")
	profilesCreateCmd.Flags().StringVar(&profileCreateResizeMode, "resize-mode", "", "Varsayilan resize modu")
	profilesCreateCmd.Flags().Float64Var(&profileCreateWidth, "width", 0, "Varsayilan genislik")
	profilesCreateCmd.Flags().Float64Var(&profileCreateHeight, "height", 0, "Varsayilan yukseklik")
	profilesCreateCmd.Flags().StringVar(&profileCreateUnit, "unit", "", "Varsayilan olcu birimi")
	profilesCreateCmd.Flags().Float64Var(&profileCreateDPI, "dpi", 0, "Varsayilan DPI")
	profilesCreateCmd.Flags().StringVar(&profileCreateMetadataMode, "metadata-mode", "", "Varsayilan metadata modu")

	profilesCmd.AddCommand(profilesListCmd)
	profilesCmd.AddCommand(profilesCreateCmd)
	rootCmd.AddCommand(profilesCmd)
}
