package profile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mlihgenel/docufy/v2/internal/batch"
	"github.com/mlihgenel/docufy/v2/internal/converter"
)

// Definition dönüşüm profili alanlarını tutar.
// nil pointer alanlar "profil bu alanı zorlamıyor" anlamına gelir.
type Definition struct {
	Name         string
	Description  string
	Source       string
	Path         string
	Quality      *int
	OnConflict   string
	Retry        *int
	RetryDelay   *time.Duration
	Report       string
	ResizePreset string
	ResizeMode   string
	Width        *float64
	Height       *float64
	Unit         string
	DPI          *float64
	MetadataMode string
}

var userHomeDir = os.UserHomeDir

var builtins = map[string]Definition{
	"social-story": {
		Name:         "social-story",
		Description:  "Dikey sosyal medya içerikleri için story/reel odaklı profil",
		Source:       "builtin",
		Quality:      intPtr(82),
		OnConflict:   converter.ConflictVersioned,
		Retry:        intPtr(1),
		RetryDelay:   durationPtr(500 * time.Millisecond),
		Report:       batch.ReportOff,
		ResizePreset: "story",
		ResizeMode:   string(converter.ResizeModePad),
		MetadataMode: converter.MetadataStrip,
	},
	"podcast-clean": {
		Name:         "podcast-clean",
		Description:  "Podcast ve konuşma sesleri için temiz varsayılanlar",
		Source:       "builtin",
		Quality:      intPtr(90),
		OnConflict:   converter.ConflictVersioned,
		Retry:        intPtr(2),
		RetryDelay:   durationPtr(1 * time.Second),
		Report:       batch.ReportTXT,
		MetadataMode: converter.MetadataPreserve,
	},
	"archive-lossless": {
		Name:         "archive-lossless",
		Description:  "Arşivleme odaklı, metadata koruyan güvenli profil",
		Source:       "builtin",
		Quality:      intPtr(100),
		OnConflict:   converter.ConflictVersioned,
		Retry:        intPtr(0),
		RetryDelay:   durationPtr(0),
		Report:       batch.ReportJSON,
		MetadataMode: converter.MetadataPreserve,
	},
}

// Resolve built-in ve kullanıcı profillerini birleştirerek isimden profile döner.
func Resolve(name string) (Definition, error) {
	key := normalizeProfileName(name)
	if key == "" {
		return Definition{}, fmt.Errorf("profil adi bos")
	}

	all, err := mergedProfiles()
	if err != nil {
		return Definition{}, err
	}
	p, ok := all[key]
	if !ok {
		return Definition{}, fmt.Errorf("profil bulunamadi: %s", name)
	}
	return p, nil
}

// List tüm görünür profilleri alfabetik olarak döner.
func List() ([]Definition, error) {
	all, err := mergedProfiles()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]Definition, 0, len(names))
	for _, name := range names {
		result = append(result, all[name])
	}
	return result, nil
}

// Names görünür profil isimlerini döner. Kullanıcı profilleri okunamazsa built-in'lere düşer.
func Names() []string {
	list, err := List()
	if err != nil {
		names := make([]string, 0, len(builtins))
		for name := range builtins {
			names = append(names, name)
		}
		sort.Strings(names)
		return names
	}

	names := make([]string, 0, len(list))
	for _, item := range list {
		names = append(names, item.Name)
	}
	return names
}

// UserProfileDir kullanıcı profil dizinini döner.
func UserProfileDir() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dizini bulunamadi: %w", err)
	}
	return filepath.Join(home, ".docufy", "profiles"), nil
}

func legacyUserProfileDir() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dizini bulunamadi: %w", err)
	}
	return filepath.Join(home, ".fileconverter", "profiles"), nil
}

// SaveUserProfile kullanıcı profilini TOML dosyası olarak kaydeder.
func SaveUserProfile(def Definition) (string, error) {
	name := normalizeProfileName(def.Name)
	if name == "" {
		return "", fmt.Errorf("profil adi bos")
	}
	def.Name = name
	if err := validateDefinition(def); err != nil {
		return "", err
	}

	dir, err := UserProfileDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("profil dizini olusturulamadi: %w", err)
	}
	def.Source = "user"
	path := filepath.Join(dir, name+".toml")
	def.Path = path
	content := encodeProfileTOML(def)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("profil kaydedilemedi: %w", err)
	}
	return path, nil
}

func mergedProfiles() (map[string]Definition, error) {
	merged := make(map[string]Definition, len(builtins))
	for name, def := range builtins {
		copied := def
		copied.Name = name
		copied.Source = "builtin"
		merged[name] = copied
	}

	userProfiles, err := loadUserProfiles()
	if err != nil {
		return nil, err
	}

	for name, userDef := range userProfiles {
		if base, ok := merged[name]; ok {
			merged[name] = mergeDefinition(base, userDef)
			continue
		}
		merged[name] = userDef
	}

	return merged, nil
}

func loadUserProfiles() (map[string]Definition, error) {
	result := make(map[string]Definition)
	dirs := make([]string, 0, 2)
	if legacyDir, err := legacyUserProfileDir(); err == nil {
		dirs = append(dirs, legacyDir)
	}
	if dir, err := UserProfileDir(); err == nil {
		dirs = append(dirs, dir)
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("profil dizini okunamadi: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".toml" {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			def, err := parseProfileFile(path)
			if err != nil {
				return nil, err
			}
			result[def.Name] = def
		}
	}
	return result, nil
}

func parseProfileFile(path string) (Definition, error) {
	f, err := os.Open(path)
	if err != nil {
		return Definition{}, err
	}
	defer f.Close()

	def := Definition{
		Name:   normalizeProfileName(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))),
		Source: "user",
		Path:   path,
	}

	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(stripInlineComment(scanner.Text()))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Definition{}, fmt.Errorf("%s:%d gecersiz satir", path, lineNo)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if err := assignProfileValue(&def, key, value); err != nil {
			return Definition{}, fmt.Errorf("%s:%d %w", path, lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return Definition{}, err
	}

	if def.Name == "" {
		return Definition{}, fmt.Errorf("%s profil adi bos", path)
	}
	return def, validateDefinition(def)
}

func assignProfileValue(def *Definition, key, raw string) error {
	switch key {
	case "name":
		v, err := parseTomlString(raw)
		if err != nil {
			return err
		}
		def.Name = normalizeProfileName(v)
	case "description":
		v, err := parseTomlString(raw)
		if err != nil {
			return err
		}
		def.Description = strings.TrimSpace(v)
	case "quality":
		v, err := parseTomlInt(raw)
		if err != nil {
			return err
		}
		def.Quality = intPtr(v)
	case "on_conflict":
		v, err := parseTomlString(raw)
		if err != nil {
			return err
		}
		def.OnConflict = strings.ToLower(strings.TrimSpace(v))
	case "retry":
		v, err := parseTomlInt(raw)
		if err != nil {
			return err
		}
		def.Retry = intPtr(v)
	case "retry_delay":
		v, err := parseTomlDuration(raw)
		if err != nil {
			return err
		}
		def.RetryDelay = durationPtr(v)
	case "report":
		v, err := parseTomlString(raw)
		if err != nil {
			return err
		}
		def.Report = strings.ToLower(strings.TrimSpace(v))
	case "resize_preset":
		v, err := parseTomlString(raw)
		if err != nil {
			return err
		}
		def.ResizePreset = strings.TrimSpace(v)
	case "resize_mode":
		v, err := parseTomlString(raw)
		if err != nil {
			return err
		}
		def.ResizeMode = strings.ToLower(strings.TrimSpace(v))
	case "width":
		v, err := parseTomlFloat(raw)
		if err != nil {
			return err
		}
		def.Width = floatPtr(v)
	case "height":
		v, err := parseTomlFloat(raw)
		if err != nil {
			return err
		}
		def.Height = floatPtr(v)
	case "unit":
		v, err := parseTomlString(raw)
		if err != nil {
			return err
		}
		def.Unit = strings.ToLower(strings.TrimSpace(v))
	case "dpi":
		v, err := parseTomlFloat(raw)
		if err != nil {
			return err
		}
		def.DPI = floatPtr(v)
	case "metadata_mode":
		v, err := parseTomlString(raw)
		if err != nil {
			return err
		}
		def.MetadataMode = strings.ToLower(strings.TrimSpace(v))
	default:
		// Bilinmeyen alanları görmezden gel.
	}
	return nil
}

func mergeDefinition(base, override Definition) Definition {
	merged := base
	merged.Name = override.Name
	merged.Path = override.Path
	merged.Source = "builtin+user"

	if strings.TrimSpace(override.Description) != "" {
		merged.Description = override.Description
	}
	if override.Quality != nil {
		merged.Quality = override.Quality
	}
	if override.OnConflict != "" {
		merged.OnConflict = override.OnConflict
	}
	if override.Retry != nil {
		merged.Retry = override.Retry
	}
	if override.RetryDelay != nil {
		merged.RetryDelay = override.RetryDelay
	}
	if override.Report != "" {
		merged.Report = override.Report
	}
	if override.ResizePreset != "" {
		merged.ResizePreset = override.ResizePreset
	}
	if override.ResizeMode != "" {
		merged.ResizeMode = override.ResizeMode
	}
	if override.Width != nil {
		merged.Width = override.Width
	}
	if override.Height != nil {
		merged.Height = override.Height
	}
	if override.Unit != "" {
		merged.Unit = override.Unit
	}
	if override.DPI != nil {
		merged.DPI = override.DPI
	}
	if override.MetadataMode != "" {
		merged.MetadataMode = override.MetadataMode
	}
	return merged
}

func validateDefinition(def Definition) error {
	if def.Quality != nil && (*def.Quality < 0 || *def.Quality > 100) {
		return fmt.Errorf("quality 0-100 araliginda olmali")
	}
	if def.Retry != nil && *def.Retry < 0 {
		return fmt.Errorf("retry 0 veya daha buyuk olmali")
	}
	if def.RetryDelay != nil && *def.RetryDelay < 0 {
		return fmt.Errorf("retry_delay negatif olamaz")
	}
	if def.DPI != nil && *def.DPI <= 0 {
		return fmt.Errorf("dpi 0'dan buyuk olmali")
	}
	if def.Width != nil && *def.Width <= 0 {
		return fmt.Errorf("width 0'dan buyuk olmali")
	}
	if def.Height != nil && *def.Height <= 0 {
		return fmt.Errorf("height 0'dan buyuk olmali")
	}
	return nil
}

func encodeProfileTOML(def Definition) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("name = %q", normalizeProfileName(def.Name)))
	if strings.TrimSpace(def.Description) != "" {
		lines = append(lines, fmt.Sprintf("description = %q", def.Description))
	}
	if def.Quality != nil {
		lines = append(lines, fmt.Sprintf("quality = %d", *def.Quality))
	}
	if def.OnConflict != "" {
		lines = append(lines, fmt.Sprintf("on_conflict = %q", def.OnConflict))
	}
	if def.Retry != nil {
		lines = append(lines, fmt.Sprintf("retry = %d", *def.Retry))
	}
	if def.RetryDelay != nil {
		lines = append(lines, fmt.Sprintf("retry_delay = %q", def.RetryDelay.String()))
	}
	if def.Report != "" {
		lines = append(lines, fmt.Sprintf("report = %q", def.Report))
	}
	if def.ResizePreset != "" {
		lines = append(lines, fmt.Sprintf("resize_preset = %q", def.ResizePreset))
	}
	if def.ResizeMode != "" {
		lines = append(lines, fmt.Sprintf("resize_mode = %q", def.ResizeMode))
	}
	if def.Width != nil {
		lines = append(lines, fmt.Sprintf("width = %g", *def.Width))
	}
	if def.Height != nil {
		lines = append(lines, fmt.Sprintf("height = %g", *def.Height))
	}
	if def.Unit != "" {
		lines = append(lines, fmt.Sprintf("unit = %q", def.Unit))
	}
	if def.DPI != nil {
		lines = append(lines, fmt.Sprintf("dpi = %g", *def.DPI))
	}
	if def.MetadataMode != "" {
		lines = append(lines, fmt.Sprintf("metadata_mode = %q", def.MetadataMode))
	}
	return strings.Join(lines, "\n") + "\n"
}

func normalizeProfileName(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func stripInlineComment(line string) string {
	var b strings.Builder
	inSingle := false
	inDouble := false
	for _, r := range line {
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				return b.String()
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

func parseTomlString(v string) (string, error) {
	v = strings.TrimSpace(v)
	if len(v) < 2 {
		return "", fmt.Errorf("gecersiz string deger")
	}
	if (strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"")) ||
		(strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'")) {
		return v[1 : len(v)-1], nil
	}
	return v, nil
}

func parseTomlInt(v string) (int, error) {
	parsed, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return 0, fmt.Errorf("gecersiz sayi degeri")
	}
	return parsed, nil
}

func parseTomlFloat(v string) (float64, error) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return 0, fmt.Errorf("gecersiz ondalik deger")
	}
	return parsed, nil
}

func parseTomlDuration(v string) (time.Duration, error) {
	str, err := parseTomlString(v)
	if err != nil {
		return 0, err
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(str))
	if err != nil {
		return 0, fmt.Errorf("gecersiz sure degeri")
	}
	return parsed, nil
}

func intPtr(v int) *int { return &v }

func floatPtr(v float64) *float64 { return &v }

func durationPtr(v time.Duration) *time.Duration { return &v }

// Helper exportları test/ileriki genişleme için tutuldu.
var (
	IntPtr      = intPtr
	FloatPtr    = floatPtr
	DurationPtr = durationPtr
)
