package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	aigateway "github.com/mlihgenel/docufy/v2/internal/ai"
	"github.com/mlihgenel/docufy/v2/internal/config"
	"github.com/mlihgenel/docufy/v2/internal/converter"
)

const (
	aiProviderOpenAI           = "openai"
	aiProviderOpenAICompatible = "openai-compatible"
	aiProviderOllama           = "ollama"
)

type aiToolDoneMsg struct {
	resultText  string
	currentFile string
	err         error
}

func (m interactiveModel) isAITextInputState() bool {
	return m.state == stateAICommandInput || m.state == stateAIAuthInput
}

func (m *interactiveModel) currentAIInputField() *string {
	switch m.state {
	case stateAICommandInput:
		return &m.aiPromptInput
	case stateAIAuthInput:
		return &m.aiAPIKeyInput
	default:
		return nil
	}
}

func (m *interactiveModel) appendAIInput(token string) bool {
	field := m.currentAIInputField()
	if field == nil {
		return false
	}

	if token == "space" {
		*field += " "
		return true
	}

	r := []rune(token)
	if len(r) != 1 {
		return false
	}
	if unicode.IsControl(r[0]) {
		return false
	}

	*field += string(r[0])
	return true
}

func (m *interactiveModel) popAIInput() {
	field := m.currentAIInputField()
	if field == nil || *field == "" {
		return
	}
	runes := []rune(*field)
	*field = string(runes[:len(runes)-1])
}

func (m interactiveModel) viewAICommandInput() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(menuTitleStyle.Render(" ◆ ⌨️ AI Komut Girişi "))
	b.WriteString("\n\n")

	if strings.TrimSpace(m.aiCurrentFile) != "" {
		b.WriteString(dimStyle.Render("  Aktif dosya: " + shortenPath(m.aiCurrentFile)))
		b.WriteString("\n")
	}

	b.WriteString(infoStyle.Render("  Komut örneği:"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(`  "Bu videonun 20 ile 30 saniyesini kırp ve mp4 olarak ver"`))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(`  "/dosya /tam/yol/video.mp4"`))
	b.WriteString("\n\n")

	input := m.aiPromptInput
	cursor := " "
	if m.showCursor {
		cursor = "▌"
	}
	b.WriteString(pathStyle.Render("  > " + input + cursor))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  Enter Çalıştır  •  Esc Geri  •  Backspace Sil"))
	b.WriteString("\n")
	return b.String()
}

func (m interactiveModel) viewAIAuthProviderSelect() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(menuTitleStyle.Render(" ◆ 🔐 AI Provider Seçimi "))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  Mevcut: " + aiProviderLabel(m.aiProvider) + " / " + m.aiModel))
	b.WriteString("\n\n")

	for i, choice := range m.choices {
		icon := ""
		if i < len(m.choiceIcons) {
			icon = m.choiceIcons[i]
		}
		desc := ""
		if i < len(m.choiceDescs) {
			desc = m.choiceDescs[i]
		}
		label := menuLine(icon, choice)
		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render("▸ " + label))
			b.WriteString("\n")
			if desc != "" {
				b.WriteString(lipgloss.NewStyle().PaddingLeft(6).Foreground(dimTextColor).Italic(true).Render(desc))
				b.WriteString("\n")
			}
		} else {
			b.WriteString(normalItemStyle.Render("  " + label))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  ↑↓ Gezin  •  Enter Seç  •  Esc Geri"))
	b.WriteString("\n")
	return b.String()
}

func (m interactiveModel) viewAIAuthInput() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(menuTitleStyle.Render(" ◆ 🔑 API Key Girişi "))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  Provider: " + aiProviderLabel(m.aiProvider)))
	b.WriteString("\n")

	envKeys := envKeysForProvider(m.aiProvider)
	if len(envKeys) > 0 {
		b.WriteString(dimStyle.Render("  Ortam değişkeni alternatifi: " + strings.Join(envKeys, ", ")))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	masked := strings.Repeat("•", len([]rune(m.aiAPIKeyInput)))
	cursor := " "
	if m.showCursor {
		cursor = "▌"
	}
	b.WriteString(pathStyle.Render("  > " + masked + cursor))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  Enter Kaydet ve Devam Et  •  Esc Geri  •  Backspace Sil"))
	b.WriteString("\n")
	return b.String()
}

func (m interactiveModel) viewAIExecuting() string {
	var b strings.Builder

	b.WriteString("\n\n")
	frame := spinnerFrames[m.spinnerIdx]
	b.WriteString(infoStyle.Render(fmt.Sprintf("  %s AI komutu çalıştırılıyor...", frame)))
	b.WriteString("\n\n")
	if strings.TrimSpace(m.aiPromptInput) != "" {
		b.WriteString(dimStyle.Render("  Komut: " + m.aiPromptInput))
		b.WriteString("\n")
	}
	if strings.TrimSpace(m.aiCurrentFile) != "" {
		b.WriteString(dimStyle.Render("  Aktif dosya: " + shortenPath(m.aiCurrentFile)))
		b.WriteString("\n")
	}
	return b.String()
}

func normalizeAIProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case aiProviderOpenAI:
		return aiProviderOpenAI
	case aiProviderOllama:
		return aiProviderOllama
	case "openai_compatible", "openai-compatible", "compatible", "openai compatible":
		return aiProviderOpenAICompatible
	default:
		return ""
	}
}

func aiProviderLabel(provider string) string {
	switch normalizeAIProvider(provider) {
	case aiProviderOpenAI:
		return "OpenAI"
	case aiProviderOllama:
		return "Ollama (Yerel)"
	default:
		return "OpenAI-compatible"
	}
}

func defaultAIModel(provider string) string {
	switch normalizeAIProvider(provider) {
	case aiProviderOllama:
		return "qwen2.5:7b"
	default:
		return "gpt-4o-mini"
	}
}

func defaultAIBaseURL(provider string) string {
	switch normalizeAIProvider(provider) {
	case aiProviderOllama:
		return "http://localhost:11434/v1"
	default:
		return "https://api.openai.com/v1"
	}
}

func displayAIBaseURL(baseURL string) string {
	if strings.TrimSpace(baseURL) == "" {
		return "(varsayılan)"
	}
	return strings.TrimSpace(baseURL)
}

func providerNeedsAPIKey(provider string) bool {
	return normalizeAIProvider(provider) != aiProviderOllama
}

func envKeysForProvider(provider string) []string {
	switch normalizeAIProvider(provider) {
	case aiProviderOpenAI:
		return []string{"OPENAI_API_KEY", "DOCUFY_AI_API_KEY"}
	case aiProviderOllama:
		return []string{"OLLAMA_API_KEY", "DOCUFY_AI_API_KEY"}
	default:
		return []string{"DOCUFY_AI_API_KEY", "OPENAI_API_KEY"}
	}
}

func (m interactiveModel) hasAIKeyConfigured() bool {
	if strings.TrimSpace(m.aiAPIKey) != "" {
		return true
	}
	for _, key := range envKeysForProvider(m.aiProvider) {
		if strings.TrimSpace(os.Getenv(key)) != "" {
			return true
		}
	}
	return false
}

func (m interactiveModel) applyAIProviderChoice(provider string) interactiveModel {
	normalized := normalizeAIProvider(provider)
	if normalized == "" {
		m.aiStatusMessage = "Provider seçimi anlaşılamadı."
		return m.goToAIIntro()
	}

	m.aiProvider = normalized
	m.aiModel = defaultAIModel(normalized)
	m.aiBaseURL = defaultAIBaseURL(normalized)
	m.aiSessionReady = false
	m.aiSessionID = ""

	err := config.SetAISettings(config.AISettings{
		Provider: m.aiProvider,
		Model:    m.aiModel,
		BaseURL:  m.aiBaseURL,
	})
	if err != nil {
		m.aiStatusMessage = "AI ayarları kaydedilemedi: " + err.Error()
	} else {
		m.aiStatusMessage = fmt.Sprintf("AI ayarı güncellendi: %s / %s", aiProviderLabel(m.aiProvider), m.aiModel)
	}

	return m.goToAIIntro()
}

func (m interactiveModel) startAISession() interactiveModel {
	if normalizeAIProvider(m.aiProvider) == "" {
		m.aiProvider = aiProviderOpenAICompatible
	}
	if strings.TrimSpace(m.aiModel) == "" {
		m.aiModel = defaultAIModel(m.aiProvider)
	}
	if strings.TrimSpace(m.aiBaseURL) == "" {
		m.aiBaseURL = defaultAIBaseURL(m.aiProvider)
	}

	if providerNeedsAPIKey(m.aiProvider) && !m.hasAIKeyConfigured() {
		m.aiStatusMessage = "Bu provider için API key gerekli. Ortam değişkeni veya giriş ekranını kullanın."
		return m.goToAIAuthInput()
	}

	if strings.TrimSpace(m.aiSessionID) == "" {
		m.aiSessionID = fmt.Sprintf("ai-%d", time.Now().Unix())
	}
	m.aiSessionReady = true
	m.aiStatusMessage = fmt.Sprintf("AI oturumu hazır: %s / %s", aiProviderLabel(m.aiProvider), m.aiModel)
	return m.goToAIChat()
}

func (m interactiveModel) doAICommand(prompt string) tea.Cmd {
	return func() tea.Msg {
		gw := aigateway.NewGateway(m.aiPolicy())
		resultText, currentFile, err := executeAICommand(gw, prompt, m.aiCurrentFile)
		return aiToolDoneMsg{
			resultText:  resultText,
			currentFile: currentFile,
			err:         err,
		}
	}
}

func (m interactiveModel) aiPolicy() aigateway.Policy {
	roots := make([]string, 0, 4)
	if wd, err := os.Getwd(); err == nil {
		roots = append(roots, wd)
	}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, home)
	}
	if strings.TrimSpace(m.defaultOutput) != "" {
		roots = append(roots, m.defaultOutput)
	}
	if strings.TrimSpace(m.aiCurrentFile) != "" {
		roots = append(roots, filepath.Dir(m.aiCurrentFile))
	}
	return aigateway.Policy{AllowedRoots: uniqueStrings(roots)}
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		v := strings.TrimSpace(value)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

func executeAICommand(gw *aigateway.Gateway, prompt string, currentFile string) (string, string, error) {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return "", currentFile, fmt.Errorf("komut boş olamaz")
	}
	lower := strings.ToLower(trimmed)

	if strings.HasPrefix(lower, "/dosya ") || strings.HasPrefix(lower, "/file ") {
		path := strings.TrimSpace(trimmed[strings.Index(trimmed, " "):])
		path = strings.Trim(path, "\"'")
		if path == "" {
			return "", currentFile, fmt.Errorf("dosya yolu belirtilmedi")
		}
		res, err := gw.GetFileInfo(aigateway.FileInfoRequest{Path: path})
		if err != nil {
			return "", currentFile, err
		}
		return fmt.Sprintf("Aktif dosya ayarlandı: %s", res.InputPath), res.InputPath, nil
	}

	inputPath, err := resolveAIInputPath(trimmed, currentFile)
	if err != nil {
		return "", currentFile, err
	}

	switch {
	case isAITrimCommand(lower):
		startSec, endSec, err := parseTrimSecondRange(lower)
		if err != nil {
			return "", inputPath, err
		}
		target := parseTargetFormat(trimmed)
		res, err := gw.TrimVideo(aigateway.TrimVideoRequest{
			InputPath:    inputPath,
			Mode:         "clip",
			Start:        strconv.Itoa(startSec),
			End:          strconv.Itoa(endSec),
			To:           target,
			Codec:        "auto",
			Quality:      0,
			MetadataMode: converter.MetadataAuto,
			OnConflict:   converter.ConflictVersioned,
		})
		if err != nil {
			return "", inputPath, err
		}
		return summarizeToolResult(res), inputPath, nil

	case isAIExtractAudioCommand(lower):
		target := parseAudioTargetFormat(trimmed)
		if target == "" {
			target = "mp3"
		}
		copyMode := strings.Contains(lower, "copy")
		res, err := gw.ExtractAudio(aigateway.ExtractAudioRequest{
			InputPath:    inputPath,
			To:           target,
			Copy:         copyMode,
			MetadataMode: converter.MetadataAuto,
			OnConflict:   converter.ConflictVersioned,
		})
		if err != nil {
			return "", inputPath, err
		}
		return summarizeToolResult(res), inputPath, nil

	case isAIInfoCommand(lower):
		res, err := gw.GetFileInfo(aigateway.FileInfoRequest{Path: inputPath})
		if err != nil {
			return "", inputPath, err
		}
		if res.FileInfo == nil {
			return "Dosya bilgisi alınamadı.", inputPath, nil
		}
		return summarizeFileInfo(*res.FileInfo), inputPath, nil

	case isAIConvertCommand(lower):
		target := parseTargetFormat(trimmed)
		if target == "" {
			return "", inputPath, fmt.Errorf("hedef format anlasilmadi (örn: mp4 olarak, to png)")
		}
		res, err := gw.ConvertFile(aigateway.ConvertFileRequest{
			InputPath:    inputPath,
			To:           target,
			Quality:      0,
			MetadataMode: converter.MetadataAuto,
			OnConflict:   converter.ConflictVersioned,
		})
		if err != nil {
			return "", inputPath, err
		}
		return summarizeToolResult(res), inputPath, nil

	default:
		return "Komut anlaşılamadı. Örn: dönüştür, kırp, ses çıkar, bilgi al. Gerekirse önce /dosya <yol> kullan.", inputPath, nil
	}
}

func summarizeFileInfo(info converter.FileInfo) string {
	parts := []string{
		fmt.Sprintf("Dosya: %s", info.FileName),
		fmt.Sprintf("Format: %s", info.Format),
		fmt.Sprintf("Boyut: %s", info.SizeText),
	}
	if info.Resolution != "" {
		parts = append(parts, "Çözünürlük: "+info.Resolution)
	}
	if info.Duration != "" {
		parts = append(parts, "Süre: "+info.Duration)
	}
	return strings.Join(parts, "  •  ")
}

func summarizeToolResult(res aigateway.ToolResult) string {
	switch res.Status {
	case aigateway.StatusSuccess:
		if strings.TrimSpace(res.OutputPath) != "" {
			return fmt.Sprintf("%s tamamlandı: %s", res.Tool, res.OutputPath)
		}
		return fmt.Sprintf("%s tamamlandı.", res.Tool)
	case aigateway.StatusSkipped:
		return fmt.Sprintf("%s atlandı: %s", res.Tool, res.Message)
	case aigateway.StatusNotImplemented:
		return fmt.Sprintf("%s henüz hazır değil: %s", res.Tool, res.Message)
	default:
		if res.Error != "" {
			return fmt.Sprintf("%s hata: %s", res.Tool, res.Error)
		}
		return fmt.Sprintf("%s başarısız.", res.Tool)
	}
}

func resolveAIInputPath(prompt string, currentFile string) (string, error) {
	if path := parseQuotedPath(prompt); path != "" {
		return path, nil
	}
	if path := parseExistingPathToken(prompt); path != "" {
		return path, nil
	}
	if strings.TrimSpace(currentFile) != "" {
		return currentFile, nil
	}
	return "", fmt.Errorf("işlenecek dosya bulunamadı. Önce /dosya <yol> yazın veya komutta dosya yolu verin")
}

func parseQuotedPath(prompt string) string {
	re := regexp.MustCompile(`"([^"]+)"|'([^']+)'`)
	matches := re.FindAllStringSubmatch(prompt, -1)
	for _, match := range matches {
		candidate := strings.TrimSpace(match[1])
		if candidate == "" {
			candidate = strings.TrimSpace(match[2])
		}
		if candidate == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

func parseExistingPathToken(prompt string) string {
	fields := strings.Fields(prompt)
	for _, field := range fields {
		candidate := strings.Trim(field, `"'.,;:!?()[]{}<>`)
		if candidate == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

func isAITrimCommand(lower string) bool {
	return strings.Contains(lower, "kırp") || strings.Contains(lower, "kirp") || strings.Contains(lower, "trim")
}

func isAIExtractAudioCommand(lower string) bool {
	return strings.Contains(lower, "ses çıkar") || strings.Contains(lower, "ses cikar") || strings.Contains(lower, "sesi al") || strings.Contains(lower, "extract audio")
}

func isAIInfoCommand(lower string) bool {
	return strings.Contains(lower, "bilgi") || strings.Contains(lower, "info")
}

func isAIConvertCommand(lower string) bool {
	return strings.Contains(lower, "dönüştür") || strings.Contains(lower, "donustur") || strings.Contains(lower, "çevir") || strings.Contains(lower, "cevir") || strings.Contains(lower, "convert")
}

func parseTrimSecondRange(promptLower string) (int, int, error) {
	numRe := regexp.MustCompile(`\d+`)
	nums := numRe.FindAllString(promptLower, -1)
	if len(nums) < 2 {
		return 0, 0, fmt.Errorf("kırpma için iki saniye değeri gerekli (örn: 20 ile 30 saniye)")
	}

	start, _ := strconv.Atoi(nums[0])
	end, _ := strconv.Atoi(nums[1])
	if end <= start {
		return 0, 0, fmt.Errorf("bitiş saniyesi başlangıçtan büyük olmalı")
	}
	return start, end, nil
}

func parseAudioTargetFormat(prompt string) string {
	target := parseTargetFormat(prompt)
	switch target {
	case "mp3", "wav", "ogg", "flac", "aac", "m4a", "opus":
		return target
	default:
		return ""
	}
}

func parseTargetFormat(prompt string) string {
	known := map[string]bool{
		"md": true, "html": true, "pdf": true, "docx": true, "txt": true, "odt": true, "rtf": true, "csv": true, "xlsx": true,
		"mp3": true, "wav": true, "ogg": true, "flac": true, "aac": true, "m4a": true, "wma": true, "opus": true, "webm": true,
		"png": true, "jpg": true, "webp": true, "bmp": true, "gif": true, "tif": true, "ico": true, "svg": true, "heic": true, "heif": true,
		"mp4": true, "mov": true, "mkv": true, "avi": true, "m4v": true, "wmv": true, "flv": true,
	}

	lower := strings.ToLower(prompt)

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\bto\s+([a-z0-9]+)\b`),
		regexp.MustCompile(`\b([a-z0-9]+)\s+olarak\b`),
		regexp.MustCompile(`\b([a-z0-9]+)'?e\b`),
	}
	for _, re := range patterns {
		m := re.FindStringSubmatch(lower)
		if len(m) > 1 {
			if normalized := converter.NormalizeFormat(m[1]); known[normalized] {
				return normalized
			}
		}
	}

	wordRe := regexp.MustCompile(`[a-z0-9]+`)
	words := wordRe.FindAllString(lower, -1)
	for i := len(words) - 1; i >= 0; i-- {
		normalized := converter.NormalizeFormat(words[i])
		if known[normalized] {
			return normalized
		}
	}
	return ""
}
