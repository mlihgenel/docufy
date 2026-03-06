package cli

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/mlihgenel/docufy/v2/internal/secret"
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

type aiExecutionConfig struct {
	Provider string
	Model    string
	BaseURL  string
	APIKey   string
}

func (m interactiveModel) isAITextInputState() bool {
	return m.state == stateAICommandInput || m.state == stateAIAuthInput || m.state == stateAIRiskConfirmInput
}

func (m *interactiveModel) currentAIInputField() *string {
	switch m.state {
	case stateAICommandInput:
		return &m.aiPromptInput
	case stateAIAuthInput:
		return &m.aiAPIKeyInput
	case stateAIRiskConfirmInput:
		return &m.aiRiskAckInput
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

func (m interactiveModel) viewAIPlanConfirm() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(menuTitleStyle.Render(" ◆ 🧠 AI Plan Onayı "))
	b.WriteString("\n\n")

	if strings.TrimSpace(m.aiPendingPrompt) != "" {
		b.WriteString(infoStyle.Render("  Komut: " + m.aiPendingPrompt))
		b.WriteString("\n\n")
	}

	planText := strings.TrimSpace(m.aiPendingPlan)
	if planText == "" {
		planText = "Komut yorumlanacak ve uygun tool çağrısı yapılacak."
	}

	planCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(0, 1).
		MarginLeft(1)
	b.WriteString(planCard.Render("Plan:\n" + planText))
	b.WriteString("\n\n")

	if len(m.aiPendingRisks) > 0 {
		riskLines := make([]string, 0, len(m.aiPendingRisks)+1)
		riskLines = append(riskLines, "Risk Notları:")
		for _, risk := range m.aiPendingRisks {
			riskLines = append(riskLines, "- "+risk)
		}
		riskCard := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(warningColor).
			Foreground(warningColor).
			Padding(0, 1).
			MarginLeft(1)
		b.WriteString(riskCard.Render(strings.Join(riskLines, "\n")))
		b.WriteString("\n\n")
	}

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
	b.WriteString(dimStyle.Render("  ↑↓ Gezin  •  Enter Seç  •  Esc Komutu Düzenle"))
	b.WriteString("\n")
	return b.String()
}

func (m interactiveModel) viewAIDone() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(menuTitleStyle.Render(" ◆ ✅ AI Sonucu "))
	b.WriteString("\n\n")

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(0, 1).
		MarginLeft(1)

	body := strings.TrimSpace(m.aiLastResult)
	if strings.TrimSpace(m.aiError) != "" {
		body = "Hata: " + m.aiError
	}
	if body == "" {
		body = strings.TrimSpace(m.aiStatusMessage)
	}
	if body == "" {
		body = "AI komutu tamamlandı."
	}
	b.WriteString(card.Render(body))
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
	b.WriteString(dimStyle.Render("  ↑↓ Gezin  •  Enter Seç  •  Esc Sohbet"))
	b.WriteString("\n")
	return b.String()
}

func (m interactiveModel) viewAIRiskConfirmInput() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(menuTitleStyle.Render(" ◆ ⚠️ Risk Onayı "))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(warningColor).Bold(true).Render("  Riskli komut tespit edildi. Devam için ONAY yazın."))
	b.WriteString("\n")
	if len(m.aiPendingRisks) > 0 {
		for _, risk := range m.aiPendingRisks {
			b.WriteString(dimStyle.Render("  - " + risk))
			b.WriteString("\n")
		}
	}
	if strings.TrimSpace(m.aiStatusMessage) != "" {
		b.WriteString("\n")
		b.WriteString(infoStyle.Render("  " + m.aiStatusMessage))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	cursor := " "
	if m.showCursor {
		cursor = "▌"
	}
	b.WriteString(pathStyle.Render("  > " + m.aiRiskAckInput + cursor))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render(`  Enter Onayla  •  Esc Plan Ekranı  •  "ONAY" yaz`))
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

func normalizeAISidecarURL(url string) string {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return strings.TrimRight(trimmed, "/")
	}
	return "http://" + strings.TrimRight(trimmed, "/")
}

func aiRuntimeLabel(sidecarURL string) string {
	sidecarURL = normalizeAISidecarURL(sidecarURL)
	if sidecarURL == "" {
		return "Yerel Gateway"
	}
	return "Strands Sidecar (" + sidecarURL + ")"
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

func aiNeedsHighRiskAck(prompt string) bool {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	if lower == "" {
		return false
	}
	if parseAIOnConflictPolicy(prompt) == converter.ConflictOverwrite {
		return true
	}
	if isAITrimCommand(lower) && parseAITrimMode(prompt) == "remove" {
		return true
	}
	return false
}

func loadStoredAIAPIKey(provider string) string {
	key, err := secret.LoadAIAPIKey(provider)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(key)
}

func persistAIAPIKey(provider string, key string) string {
	err := secret.SaveAIAPIKey(provider, key)
	if err == nil {
		return ""
	}
	if errors.Is(err, secret.ErrDisabled) {
		return ""
	}
	if errors.Is(err, secret.ErrNotSupported) {
		return "Keychain bu sistemde desteklenmiyor; anahtar sadece bu oturumda kullanılacak."
	}
	return "API key keychain'e kaydedilemedi: " + err.Error()
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
	m.aiAPIKey = loadStoredAIAPIKey(normalized)
	m.aiSessionReady = false
	m.aiSessionID = ""

	err := config.SetAISettings(config.AISettings{
		Provider:   m.aiProvider,
		Model:      m.aiModel,
		BaseURL:    m.aiBaseURL,
		SidecarURL: m.aiSidecarURL,
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
	m.aiSidecarURL = normalizeAISidecarURL(m.aiSidecarURL)
	if providerNeedsAPIKey(m.aiProvider) && strings.TrimSpace(m.aiAPIKey) == "" {
		m.aiAPIKey = loadStoredAIAPIKey(m.aiProvider)
	}

	if providerNeedsAPIKey(m.aiProvider) && !m.hasAIKeyConfigured() {
		m.aiStatusMessage = "Bu provider için API key gerekli. Ortam değişkeni veya giriş ekranını kullanın."
		return m.goToAIAuthInput()
	}

	if strings.TrimSpace(m.aiSessionID) == "" {
		m.aiSessionID = fmt.Sprintf("ai-%d", time.Now().Unix())
	}
	m.aiSessionReady = true

	if sidecar := m.sidecarClient(); sidecar != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if _, err := sidecar.Health(ctx); err != nil {
			m.aiStatusMessage = fmt.Sprintf("AI oturumu hazır (local fallback). Sidecar erişilemedi: %v", err)
			return m.goToAIChat()
		}
	}

	m.aiStatusMessage = fmt.Sprintf("AI oturumu hazır: %s / %s • %s", aiProviderLabel(m.aiProvider), m.aiModel, aiRuntimeLabel(m.aiSidecarURL))
	return m.goToAIChat()
}

func (m interactiveModel) doAICommand(prompt string) tea.Cmd {
	runtime := aiExecutionConfig{
		Provider: m.aiProvider,
		Model:    m.aiModel,
		BaseURL:  m.aiBaseURL,
		APIKey:   m.aiAPIKey,
	}

	return func() tea.Msg {
		if sidecar := m.sidecarClient(); sidecar != nil {
			ctx, cancel := context.WithTimeout(context.Background(), aigateway.DefaultSidecarTimeout)
			defer cancel()
			res, err := sidecar.RunCommand(ctx, aigateway.SidecarCommandRequest{
				Prompt:        prompt,
				CurrentFile:   m.aiCurrentFile,
				AllowedRoots:  m.aiPolicy().AllowedRoots,
				DefaultOutput: m.defaultOutput,
				Provider:      m.aiProvider,
				Model:         m.aiModel,
			})
			if err == nil {
				currentFile := m.aiCurrentFile
				if strings.TrimSpace(res.CurrentFile) != "" {
					currentFile = res.CurrentFile
				}
				resultText := strings.TrimSpace(res.ResultText)
				if resultText == "" {
					resultText = "AI komutu tamamlandı."
				}
				m.writeAIAuditLog("sidecar", prompt, currentFile, "success", resultText, "")
				return aiToolDoneMsg{
					resultText:  resultText,
					currentFile: currentFile,
					err:         nil,
				}
			}

			gw := aigateway.NewGateway(m.aiPolicy())
			resultText, currentFile, localErr := executeAICommand(gw, prompt, m.aiCurrentFile, runtime)
			if localErr != nil {
				m.writeAIAuditLog("sidecar+local", prompt, m.aiCurrentFile, "failed", "", fmt.Sprintf("sidecar: %v; local: %v", err, localErr))
				return aiToolDoneMsg{
					err: fmt.Errorf("sidecar hatası: %v; local hata: %w", err, localErr),
				}
			}
			m.writeAIAuditLog("sidecar+local", prompt, currentFile, "success", resultText, fmt.Sprintf("sidecar: %v", err))
			return aiToolDoneMsg{
				resultText:  fmt.Sprintf("Sidecar erişilemedi (%v). Yerel gateway ile devam edildi. %s", err, resultText),
				currentFile: currentFile,
				err:         nil,
			}
		}

		gw := aigateway.NewGateway(m.aiPolicy())
		resultText, currentFile, err := executeAICommand(gw, prompt, m.aiCurrentFile, runtime)
		if err != nil {
			m.writeAIAuditLog("local", prompt, m.aiCurrentFile, "failed", "", err.Error())
		} else {
			m.writeAIAuditLog("local", prompt, currentFile, "success", resultText, "")
		}
		return aiToolDoneMsg{
			resultText:  resultText,
			currentFile: currentFile,
			err:         err,
		}
	}
}

func (m interactiveModel) sidecarClient() *aigateway.SidecarClient {
	if strings.TrimSpace(m.aiSidecarURL) == "" {
		return nil
	}
	return aigateway.NewSidecarClient(m.aiSidecarURL, 0)
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

func (m interactiveModel) buildAIPlan(prompt string) (string, []string) {
	trimmed := strings.TrimSpace(prompt)
	lower := strings.ToLower(trimmed)
	steps := make([]string, 0, 4)
	risks := make([]string, 0, 3)
	onConflict := parseAIOnConflictPolicy(trimmed)
	trimMode := parseAITrimMode(trimmed)

	if strings.HasPrefix(lower, "/dosya ") || strings.HasPrefix(lower, "/file ") {
		steps = append(steps, "Aktif dosya güncellenecek ve doğrulama için get_file_info çağrılacak.")
	} else {
		switch {
		case isAIWatchCommand(lower):
			dir, err := resolveAIDirectoryPath(trimmed, m.aiCurrentFile)
			from, to := parseAIFormatPair(trimmed)
			if err == nil {
				steps = append(steps, fmt.Sprintf("Watch görevi üretilecek: klasör=%s from=%s to=%s.", dir, fromOrUnknown(from), fromOrUnknown(to)))
			} else {
				steps = append(steps, "Watch görevi üretilecek (klasör yolu komuttan doğrulanacak).")
			}
			steps = append(steps, "Komut otomatik çalıştırılmayacak; güvenli görev dosyası oluşturulacak.")

		case isAIBatchCommand(lower):
			dir, err := resolveAIDirectoryPath(trimmed, m.aiCurrentFile)
			from, to := parseAIFormatPair(trimmed)
			if err == nil {
				steps = append(steps, fmt.Sprintf("Batch görevi üretilecek: klasör=%s from=%s to=%s.", dir, fromOrUnknown(from), fromOrUnknown(to)))
			} else {
				steps = append(steps, "Batch görevi üretilecek (klasör yolu komuttan doğrulanacak).")
			}
			steps = append(steps, "Komut otomatik çalıştırılmayacak; güvenli görev dosyası oluşturulacak.")

		case isAITranscribeCommand(lower) && isAISummarizeCommand(lower):
			steps = append(steps, "transcribe_media aracı ile medya metne dökülecek.")
			steps = append(steps, "summarize_transcript aracı ile transcript özeti üretilecek.")
		case isAITranscribeCommand(lower):
			steps = append(steps, "transcribe_media aracı ile medya metne dökülecek (transcript dosyası üretilecek).")
		case isAISummarizeCommand(lower):
			steps = append(steps, "summarize_transcript aracı ile transcript dosyası özetlenecek (markdown çıktı üretilecek).")

		case isAITrimCommand(lower):
			startSec, endSec, err := parseTrimSecondRange(lower)
			if err == nil {
				steps = append(steps, fmt.Sprintf("trim_video aracı %s modunda %d-%d saniye aralığı ile çalıştırılacak.", trimMode, startSec, endSec))
			} else {
				steps = append(steps, "trim_video aracı ile kırpma denenecek (zaman aralığı doğrulanacak).")
			}
		case isAIExtractAudioCommand(lower):
			target := parseAudioTargetFormat(trimmed)
			if target == "" {
				target = "mp3"
			}
			steps = append(steps, fmt.Sprintf("extract_audio aracı %s formatında çalıştırılacak (on_conflict: %s).", target, onConflict))
		case isAIInfoCommand(lower):
			steps = append(steps, "get_file_info aracı ile medya bilgileri alınacak.")
		case isAIConvertCommand(lower):
			target := parseTargetFormat(trimmed)
			if target == "" {
				steps = append(steps, "convert_file aracı ile dönüştürme denenecek (hedef format komuttan çözümlenecek).")
			} else {
				steps = append(steps, fmt.Sprintf("convert_file aracı %s formatına dönüştürmek için çalıştırılacak (on_conflict: %s).", target, onConflict))
			}
		default:
			steps = append(steps, "Komut yorumlanacak; desteklenen bir işlemse ilgili tool çağrılacak.")
		}
	}

	if strings.TrimSpace(m.aiCurrentFile) == "" && parseQuotedPath(trimmed) == "" && parseExistingPathToken(trimmed) == "" && !strings.HasPrefix(lower, "/dosya ") && !strings.HasPrefix(lower, "/file ") {
		risks = append(risks, "Aktif dosya belirtilmemiş görünüyor; önce /dosya <yol> vermen gerekebilir.")
	}
	if onConflict == converter.ConflictOverwrite {
		risks = append(risks, "on_conflict=overwrite: mevcut çıktı dosyası üzerine yazılabilir.")
	}
	if trimMode == "remove" && isAITrimCommand(lower) {
		risks = append(risks, "trim mode=remove: videodan aralık silme talebi var.")
	}
	if isAIBatchCommand(lower) || isAIWatchCommand(lower) {
		risks = append(risks, "Toplu/watch komutları otomatik icra edilmez; görev dosyası üretilir.")
	}

	return strings.Join(steps, "\n"), uniqueStrings(risks)
}

func executeAICommand(gw *aigateway.Gateway, prompt string, currentFile string, runtime aiExecutionConfig) (string, string, error) {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return "", currentFile, fmt.Errorf("komut boş olamaz")
	}
	lower := strings.ToLower(trimmed)
	onConflict := parseAIOnConflictPolicy(trimmed)
	trimMode := parseAITrimMode(trimmed)
	metadataMode := parseAIMetadataMode(trimmed)

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

	if isAIWatchCommand(lower) {
		dir, err := resolveAIDirectoryPath(trimmed, currentFile)
		if err != nil {
			return "", currentFile, err
		}
		validatedDir, err := gw.ValidatePath(dir)
		if err != nil {
			return "", currentFile, err
		}
		from, to := parseAIFormatPair(trimmed)
		if from == "" || to == "" {
			return "", currentFile, fmt.Errorf("watch icin kaynak/hedef format gerekli (örn: from mp4 to gif)")
		}
		onConflict := parseAIOnConflictPolicy(trimmed)
		metadataMode := parseAIMetadataMode(trimmed)
		recursive := parseAIRecursive(trimmed)
		command := buildAIWatchCommand(validatedDir, from, to, onConflict, metadataMode, recursive)
		taskPath, err := writeAITaskFile("watch", trimmed, command)
		if err != nil {
			return "", currentFile, err
		}
		return fmt.Sprintf("Watch görevi üretildi (otomatik çalıştırılmadı).\nKomut: %s\nGörev dosyası: %s", command, taskPath), currentFile, nil
	}

	if isAIBatchCommand(lower) {
		dir, err := resolveAIDirectoryPath(trimmed, currentFile)
		if err != nil {
			return "", currentFile, err
		}
		validatedDir, err := gw.ValidatePath(dir)
		if err != nil {
			return "", currentFile, err
		}
		from, to := parseAIFormatPair(trimmed)
		if from == "" || to == "" {
			return "", currentFile, fmt.Errorf("batch icin kaynak/hedef format gerekli (örn: from png to webp)")
		}
		onConflict := parseAIOnConflictPolicy(trimmed)
		metadataMode := parseAIMetadataMode(trimmed)
		recursive := parseAIRecursive(trimmed)
		command := buildAIBatchCommand(validatedDir, from, to, onConflict, metadataMode, recursive)
		taskPath, err := writeAITaskFile("batch", trimmed, command)
		if err != nil {
			return "", currentFile, err
		}
		return fmt.Sprintf("Batch görevi üretildi (otomatik çalıştırılmadı).\nKomut: %s\nGörev dosyası: %s", command, taskPath), currentFile, nil
	}

	if isAITranscribeCommand(lower) && isAISummarizeCommand(lower) {
		inputPath, err := resolveAIInputPath(trimmed, currentFile)
		if err != nil {
			return "", currentFile, err
		}

		transcribeReq := aigateway.TranscribeMediaRequest{
			InputPath:   inputPath,
			Provider:    runtime.Provider,
			Model:       resolveAITranscribeModel(trimmed, runtime.Model),
			BaseURL:     runtime.BaseURL,
			APIKey:      runtime.APIKey,
			Language:    parseAITranscribeLanguage(trimmed),
			Diarization: parseAIDiarization(trimmed),
			OnConflict:  onConflict,
		}
		transcribeResult, err := gw.TranscribeMedia(transcribeReq)
		if err != nil {
			return "", inputPath, err
		}

		transcriptPath := strings.TrimSpace(transcribeResult.TranscriptPath)
		if transcriptPath == "" {
			transcriptPath = strings.TrimSpace(transcribeResult.OutputPath)
		}
		if transcriptPath == "" {
			return "", inputPath, fmt.Errorf("transcript dosya yolu alinamadi")
		}

		summaryReq := aigateway.SummarizeTranscriptRequest{
			TranscriptPath: transcriptPath,
			Provider:       runtime.Provider,
			Model:          runtime.Model,
			BaseURL:        runtime.BaseURL,
			APIKey:         runtime.APIKey,
			Style:          parseAISummaryStyle(trimmed),
			TargetLanguage: parseAISummaryTargetLanguage(trimmed),
			OnConflict:     onConflict,
		}
		summaryResult, err := gw.SummarizeTranscript(summaryReq)
		if err != nil {
			return "", transcriptPath, err
		}

		nextCurrentFile := transcriptPath
		if strings.TrimSpace(summaryResult.SummaryPath) != "" {
			nextCurrentFile = strings.TrimSpace(summaryResult.SummaryPath)
		}
		resultText := summarizeToolResult(transcribeResult) + "\n" + summarizeToolResult(summaryResult)
		return resultText, nextCurrentFile, nil
	}

	if isAISummarizeCommand(lower) {
		transcriptPath, err := resolveAITranscriptPath(trimmed, currentFile)
		if err != nil {
			return "", currentFile, err
		}
		res, err := gw.SummarizeTranscript(aigateway.SummarizeTranscriptRequest{
			TranscriptPath: transcriptPath,
			Provider:       runtime.Provider,
			Model:          runtime.Model,
			BaseURL:        runtime.BaseURL,
			APIKey:         runtime.APIKey,
			Style:          parseAISummaryStyle(trimmed),
			TargetLanguage: parseAISummaryTargetLanguage(trimmed),
			OnConflict:     onConflict,
		})
		if err != nil {
			return "", transcriptPath, err
		}
		nextCurrentFile := transcriptPath
		if strings.TrimSpace(res.SummaryPath) != "" {
			nextCurrentFile = strings.TrimSpace(res.SummaryPath)
		}
		return summarizeToolResult(res), nextCurrentFile, nil
	}

	if isAITranscribeCommand(lower) {
		inputPath, err := resolveAIInputPath(trimmed, currentFile)
		if err != nil {
			return "", currentFile, err
		}
		res, err := gw.TranscribeMedia(aigateway.TranscribeMediaRequest{
			InputPath:   inputPath,
			Provider:    runtime.Provider,
			Model:       resolveAITranscribeModel(trimmed, runtime.Model),
			BaseURL:     runtime.BaseURL,
			APIKey:      runtime.APIKey,
			Language:    parseAITranscribeLanguage(trimmed),
			Diarization: parseAIDiarization(trimmed),
			OnConflict:  onConflict,
		})
		if err != nil {
			return "", inputPath, err
		}
		nextCurrentFile := inputPath
		if strings.TrimSpace(res.TranscriptPath) != "" {
			nextCurrentFile = strings.TrimSpace(res.TranscriptPath)
		}
		return summarizeToolResult(res), nextCurrentFile, nil
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
			Mode:         trimMode,
			Start:        strconv.Itoa(startSec),
			End:          strconv.Itoa(endSec),
			To:           target,
			Codec:        "auto",
			Quality:      0,
			MetadataMode: metadataMode,
			OnConflict:   onConflict,
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
			MetadataMode: metadataMode,
			OnConflict:   onConflict,
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
			MetadataMode: metadataMode,
			OnConflict:   onConflict,
		})
		if err != nil {
			return "", inputPath, err
		}
		return summarizeToolResult(res), inputPath, nil

	default:
		return "Komut anlaşılamadı. Örn: dönüştür, kırp, ses çıkar, transkript al, özetle, bilgi al. Gerekirse önce /dosya <yol> kullan.", inputPath, nil
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

func resolveAITranscriptPath(prompt string, currentFile string) (string, error) {
	if path := parseQuotedPath(prompt); path != "" {
		return path, nil
	}
	if path := parseExistingPathToken(prompt); path != "" {
		return path, nil
	}
	if strings.TrimSpace(currentFile) == "" {
		return "", fmt.Errorf("transcript dosyasi bulunamadi. Komutta transcript yolu verin veya once transcribe calistirin")
	}

	ext := converter.NormalizeFormat(filepath.Ext(currentFile))
	switch ext {
	case "txt", "md", "srt", "vtt":
		return currentFile, nil
	default:
		// transcript uzantısı olmayabilir; kullanıcı akışını kesmemek için yine de kullan.
		return currentFile, nil
	}
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

func isAIBatchCommand(lower string) bool {
	return strings.Contains(lower, "batch") || strings.Contains(lower, "toplu")
}

func isAIWatchCommand(lower string) bool {
	return strings.Contains(lower, "watch") || strings.Contains(lower, "izle")
}

func isAITranscribeCommand(lower string) bool {
	return strings.Contains(lower, "transcribe") ||
		strings.Contains(lower, "transkript") ||
		strings.Contains(lower, "metne dök") ||
		strings.Contains(lower, "metne dok") ||
		strings.Contains(lower, "metne donustur") ||
		strings.Contains(lower, "metne dönüştür") ||
		strings.Contains(lower, "sesi metne") ||
		strings.Contains(lower, "speech to text")
}

func isAISummarizeCommand(lower string) bool {
	return strings.Contains(lower, "özet") ||
		strings.Contains(lower, "ozet") ||
		strings.Contains(lower, "summarize") ||
		strings.Contains(lower, "summary")
}

func parseAITrimMode(prompt string) string {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	if lower == "" {
		return "clip"
	}
	if strings.Contains(lower, "remove") || strings.Contains(lower, "sil") {
		return "remove"
	}
	return "clip"
}

func parseAIOnConflictPolicy(prompt string) string {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	if lower == "" {
		return converter.ConflictVersioned
	}
	if strings.Contains(lower, "overwrite") || strings.Contains(lower, "üzerine yaz") || strings.Contains(lower, "uzerine yaz") {
		return converter.ConflictOverwrite
	}
	if strings.Contains(lower, "skip") || strings.Contains(lower, "atla") {
		return converter.ConflictSkip
	}
	return converter.ConflictVersioned
}

func parseAIMetadataMode(prompt string) string {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	if lower == "" {
		return converter.MetadataAuto
	}
	if strings.Contains(lower, "strip metadata") || strings.Contains(lower, "metadata temizle") || strings.Contains(lower, "metadata sil") {
		return converter.MetadataStrip
	}
	if strings.Contains(lower, "preserve metadata") || strings.Contains(lower, "metadata koru") {
		return converter.MetadataPreserve
	}
	return converter.MetadataAuto
}

func parseAITranscribeLanguage(prompt string) string {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	switch {
	case strings.Contains(lower, "turkce"), strings.Contains(lower, "türkçe"), strings.Contains(lower, "turkish"):
		return "tr"
	case strings.Contains(lower, "ingilizce"), strings.Contains(lower, "english"):
		return "en"
	default:
		return ""
	}
}

func resolveAITranscribeModel(prompt string, runtimeModel string) string {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	if strings.Contains(lower, "whisper-1") {
		return "whisper-1"
	}
	if strings.Contains(lower, "gpt-4o-mini-transcribe") {
		return "gpt-4o-mini-transcribe"
	}
	rt := strings.TrimSpace(runtimeModel)
	if strings.Contains(strings.ToLower(rt), "transcribe") || strings.Contains(strings.ToLower(rt), "whisper") {
		return rt
	}
	return ""
}

func parseAIDiarization(prompt string) bool {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	return strings.Contains(lower, "diarization") || strings.Contains(lower, "konusmaci ayrimi") || strings.Contains(lower, "konuşmacı ayrımı")
}

func parseAISummaryStyle(prompt string) string {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	switch {
	case strings.Contains(lower, "madde"), strings.Contains(lower, "bullet"):
		return "bullet_points"
	case strings.Contains(lower, "detay"), strings.Contains(lower, "ayrintili"), strings.Contains(lower, "ayrıntılı"):
		return "detailed"
	case strings.Contains(lower, "kisa"), strings.Contains(lower, "kısa"), strings.Contains(lower, "short"):
		return "short"
	default:
		return ""
	}
}

func parseAISummaryTargetLanguage(prompt string) string {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	switch {
	case strings.Contains(lower, "turkce"), strings.Contains(lower, "türkçe"), strings.Contains(lower, "turkish"):
		return "tr"
	case strings.Contains(lower, "ingilizce"), strings.Contains(lower, "english"):
		return "en"
	default:
		return ""
	}
}

func parseAIRecursive(prompt string) bool {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	return strings.Contains(lower, "recursive") || strings.Contains(lower, "alt klas") || strings.Contains(lower, "alt dizin")
}

func parseAIFormatPair(prompt string) (string, string) {
	lower := strings.ToLower(prompt)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\bfrom\s+([a-z0-9]+)\s+to\s+([a-z0-9]+)\b`),
		regexp.MustCompile(`\b([a-z0-9]+)'?den\s+([a-z0-9]+)'?ye\b`),
	}

	for _, re := range patterns {
		m := re.FindStringSubmatch(lower)
		if len(m) > 2 {
			from := converter.NormalizeFormat(m[1])
			to := converter.NormalizeFormat(m[2])
			if from != "" && to != "" {
				return from, to
			}
		}
	}

	from := parseAISourceFormat(prompt)
	to := parseTargetFormat(prompt)
	return from, to
}

func parseAISourceFormat(prompt string) string {
	lower := strings.ToLower(prompt)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\bfrom\s+([a-z0-9]+)\b`),
		regexp.MustCompile(`\b([a-z0-9]+)'?den\b`),
	}

	for _, re := range patterns {
		m := re.FindStringSubmatch(lower)
		if len(m) > 1 {
			if normalized := converter.NormalizeFormat(m[1]); normalized != "" {
				return normalized
			}
		}
	}
	return ""
}

func resolveAIDirectoryPath(prompt string, currentFile string) (string, error) {
	if path := parseQuotedDirectory(prompt); path != "" {
		return path, nil
	}
	if path := parseExistingDirectoryToken(prompt); path != "" {
		return path, nil
	}
	if strings.TrimSpace(currentFile) != "" {
		return filepath.Dir(currentFile), nil
	}
	return "", fmt.Errorf("islenecek klasor bulunamadi. Komutta klasor yolu verin veya once /dosya <yol> kullanin")
}

func parseQuotedDirectory(prompt string) string {
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
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}

func parseExistingDirectoryToken(prompt string) string {
	fields := strings.Fields(prompt)
	for _, field := range fields {
		candidate := strings.Trim(field, `"'.,;:!?()[]{}<>`)
		if candidate == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}

func buildAIBatchCommand(dir string, from string, to string, onConflict string, metadataMode string, recursive bool) string {
	parts := []string{
		"docufy", "batch", quoteShellArg(dir),
		"--from", from,
		"--to", to,
		"--on-conflict", onConflict,
	}
	if recursive {
		parts = append(parts, "--recursive")
	}
	switch metadataMode {
	case converter.MetadataStrip:
		parts = append(parts, "--strip-metadata")
	case converter.MetadataPreserve:
		parts = append(parts, "--preserve-metadata")
	}
	return strings.Join(parts, " ")
}

func buildAIWatchCommand(dir string, from string, to string, onConflict string, metadataMode string, recursive bool) string {
	parts := []string{
		"docufy", "watch", quoteShellArg(dir),
		"--from", from,
		"--to", to,
		"--on-conflict", onConflict,
	}
	if recursive {
		parts = append(parts, "--recursive")
	}
	switch metadataMode {
	case converter.MetadataStrip:
		parts = append(parts, "--strip-metadata")
	case converter.MetadataPreserve:
		parts = append(parts, "--preserve-metadata")
	}
	return strings.Join(parts, " ")
}

func writeAITaskFile(kind string, prompt string, command string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	taskDir := filepath.Join(wd, ".docufy", "tasks")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return "", err
	}
	timestamp := time.Now().UTC().Format("20060102-150405.000000000")
	file := filepath.Join(taskDir, fmt.Sprintf("%s-%s.md", kind, timestamp))
	content := fmt.Sprintf(
		"# Docufy AI Task (%s)\n\nPrompt:\n%s\n\nCommand:\n%s\n",
		kind,
		prompt,
		command,
	)
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		return "", err
	}
	return file, nil
}

func quoteShellArg(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func fromOrUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "?"
	}
	return value
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

type aiAuditLogRecord struct {
	Timestamp   string `json:"timestamp"`
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	Runtime     string `json:"runtime"`
	Mode        string `json:"mode"`
	Prompt      string `json:"prompt"`
	CurrentFile string `json:"current_file,omitempty"`
	Status      string `json:"status"`
	Result      string `json:"result,omitempty"`
	Error       string `json:"error,omitempty"`
}

func (m interactiveModel) writeAIAuditLog(mode string, prompt string, currentFile string, status string, result string, errText string) {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return
	}
	dir := filepath.Join(home, ".docufy")
	if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
		return
	}

	record := aiAuditLogRecord{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Provider:    m.aiProvider,
		Model:       m.aiModel,
		Runtime:     aiRuntimeLabel(m.aiSidecarURL),
		Mode:        mode,
		Prompt:      strings.TrimSpace(prompt),
		CurrentFile: strings.TrimSpace(currentFile),
		Status:      status,
		Result:      strings.TrimSpace(result),
		Error:       strings.TrimSpace(errText),
	}

	payload, marshalErr := json.Marshal(record)
	if marshalErr != nil {
		return
	}

	file := filepath.Join(dir, "ai_audit.log")
	f, openErr := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if openErr != nil {
		return
	}
	defer f.Close()

	_, _ = f.Write(append(payload, '\n'))
}
