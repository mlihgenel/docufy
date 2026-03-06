package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mlihgenel/docufy/v2/internal/converter"
)

const (
	aiProviderOpenAI           = "openai"
	aiProviderOpenAICompatible = "openai-compatible"
	aiProviderOllama           = "ollama"
)

func (g *Gateway) TranscribeMedia(req TranscribeMediaRequest) (ToolResult, error) {
	inputPath, err := g.policy.ValidatePath(req.InputPath)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTranscribeMedia, Error: err.Error()}, err
	}
	info, err := os.Stat(inputPath)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTranscribeMedia, InputPath: inputPath, Error: err.Error()}, err
	}
	if info.IsDir() {
		err = fmt.Errorf("input media dosyasi dizin olamaz")
		return ToolResult{Status: StatusFailed, Tool: ToolTranscribeMedia, InputPath: inputPath, Error: err.Error()}, err
	}

	outputPath, err := g.resolveTextArtifactPath(inputPath, req.OutputPath, req.OutputDir, req.Name, "transcript", "txt")
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTranscribeMedia, InputPath: inputPath, Error: err.Error()}, err
	}
	conflict := converter.NormalizeConflictPolicy(req.OnConflict)
	resolvedOutput, skip, err := converter.ResolveOutputPathConflict(outputPath, conflict)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTranscribeMedia, InputPath: inputPath, OutputPath: outputPath, Error: err.Error()}, err
	}
	if skip {
		return ToolResult{
			Status:         StatusSkipped,
			Tool:           ToolTranscribeMedia,
			InputPath:      inputPath,
			OutputPath:     outputPath,
			TranscriptPath: outputPath,
			Message:        "output exists (policy: skip)",
		}, nil
	}

	transcriptText, segments, err := runTranscriptionRequest(req, inputPath)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTranscribeMedia, InputPath: inputPath, OutputPath: resolvedOutput, Error: err.Error()}, err
	}
	transcriptText = strings.TrimSpace(transcriptText)
	if transcriptText == "" {
		err = fmt.Errorf("provider bos transcript dondu")
		return ToolResult{Status: StatusFailed, Tool: ToolTranscribeMedia, InputPath: inputPath, OutputPath: resolvedOutput, Error: err.Error()}, err
	}

	if err := os.MkdirAll(filepath.Dir(resolvedOutput), 0755); err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTranscribeMedia, InputPath: inputPath, OutputPath: resolvedOutput, Error: err.Error()}, err
	}
	if err := os.WriteFile(resolvedOutput, []byte(transcriptText+"\n"), 0644); err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTranscribeMedia, InputPath: inputPath, OutputPath: resolvedOutput, Error: err.Error()}, err
	}

	return ToolResult{
		Status:             StatusSuccess,
		Tool:               ToolTranscribeMedia,
		InputPath:          inputPath,
		OutputPath:         resolvedOutput,
		TranscriptPath:     resolvedOutput,
		TranscriptSegments: segments,
		Message:            "transcription completed",
	}, nil
}

func (g *Gateway) SummarizeTranscript(req SummarizeTranscriptRequest) (ToolResult, error) {
	transcriptPath, err := g.policy.ValidatePath(req.TranscriptPath)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolSummarizeTranscript, Error: err.Error()}, err
	}
	info, err := os.Stat(transcriptPath)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolSummarizeTranscript, InputPath: transcriptPath, Error: err.Error()}, err
	}
	if info.IsDir() {
		err = fmt.Errorf("transcript dosyasi dizin olamaz")
		return ToolResult{Status: StatusFailed, Tool: ToolSummarizeTranscript, InputPath: transcriptPath, Error: err.Error()}, err
	}

	data, err := os.ReadFile(transcriptPath)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolSummarizeTranscript, InputPath: transcriptPath, Error: err.Error()}, err
	}
	transcriptText := strings.TrimSpace(string(data))
	if transcriptText == "" {
		err = fmt.Errorf("transcript bos")
		return ToolResult{Status: StatusFailed, Tool: ToolSummarizeTranscript, InputPath: transcriptPath, Error: err.Error()}, err
	}

	outputPath, err := g.resolveTextArtifactPath(transcriptPath, req.OutputPath, req.OutputDir, req.Name, "summary", "md")
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolSummarizeTranscript, InputPath: transcriptPath, Error: err.Error()}, err
	}
	conflict := converter.NormalizeConflictPolicy(req.OnConflict)
	resolvedOutput, skip, err := converter.ResolveOutputPathConflict(outputPath, conflict)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolSummarizeTranscript, InputPath: transcriptPath, OutputPath: outputPath, Error: err.Error()}, err
	}
	if skip {
		return ToolResult{
			Status:      StatusSkipped,
			Tool:        ToolSummarizeTranscript,
			InputPath:   transcriptPath,
			OutputPath:  outputPath,
			SummaryPath: outputPath,
			Message:     "output exists (policy: skip)",
		}, nil
	}

	summaryText, err := runSummaryRequest(req, transcriptText)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolSummarizeTranscript, InputPath: transcriptPath, OutputPath: resolvedOutput, Error: err.Error()}, err
	}
	summaryText = strings.TrimSpace(summaryText)
	if summaryText == "" {
		err = fmt.Errorf("provider bos ozet dondu")
		return ToolResult{Status: StatusFailed, Tool: ToolSummarizeTranscript, InputPath: transcriptPath, OutputPath: resolvedOutput, Error: err.Error()}, err
	}

	content := renderSummaryMarkdown(summaryText, req.Style, req.TargetLanguage)
	if err := os.MkdirAll(filepath.Dir(resolvedOutput), 0755); err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolSummarizeTranscript, InputPath: transcriptPath, OutputPath: resolvedOutput, Error: err.Error()}, err
	}
	if err := os.WriteFile(resolvedOutput, []byte(content), 0644); err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolSummarizeTranscript, InputPath: transcriptPath, OutputPath: resolvedOutput, Error: err.Error()}, err
	}

	return ToolResult{
		Status:      StatusSuccess,
		Tool:        ToolSummarizeTranscript,
		InputPath:   transcriptPath,
		OutputPath:  resolvedOutput,
		SummaryPath: resolvedOutput,
		SummaryText: summaryText,
		Message:     "summary completed",
	}, nil
}

func (g *Gateway) resolveTextArtifactPath(inputPath string, outputPath string, outputDir string, name string, defaultSuffix string, ext string) (string, error) {
	if strings.TrimSpace(outputPath) != "" {
		return g.policy.ValidatePath(outputPath)
	}

	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	if strings.TrimSpace(name) != "" {
		baseName = strings.TrimSpace(name)
	} else if strings.TrimSpace(defaultSuffix) != "" {
		baseName = baseName + "_" + strings.TrimSpace(defaultSuffix)
	}

	if filepath.Ext(baseName) == "" {
		baseName += "." + strings.TrimPrefix(strings.TrimSpace(ext), ".")
	}

	if strings.TrimSpace(outputDir) != "" {
		return g.policy.ValidatePath(filepath.Join(outputDir, baseName))
	}
	return g.policy.ValidatePath(filepath.Join(filepath.Dir(inputPath), baseName))
}

func runTranscriptionRequest(req TranscribeMediaRequest, inputPath string) (string, []TranscribeSegment, error) {
	provider := normalizeAIProviderName(req.Provider)
	baseURL := providerDefaultBaseURL(provider, req.BaseURL)
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = defaultTranscribeModel(provider)
	}
	apiKey := resolveProviderAPIKey(provider, req.APIKey)
	if providerNeedsAPIKey(provider, baseURL) && apiKey == "" {
		return "", nil, fmt.Errorf("api key gerekli")
	}

	file, err := os.Open(inputPath)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filepath.Base(inputPath))
	if err != nil {
		return "", nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", nil, err
	}
	if err := writer.WriteField("model", model); err != nil {
		return "", nil, err
	}
	if language := strings.TrimSpace(req.Language); language != "" {
		if err := writer.WriteField("language", language); err != nil {
			return "", nil, err
		}
	}
	if req.Diarization {
		_ = writer.WriteField("diarization", "true")
	}
	if err := writer.Close(); err != nil {
		return "", nil, err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/audio/transcriptions"
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, &body)
	if err != nil {
		return "", nil, err
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	if strings.TrimSpace(apiKey) != "" {
		httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	payload, readErr := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if readErr != nil {
		return "", nil, readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := strings.TrimSpace(string(payload))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return "", nil, fmt.Errorf("provider http %d: %s", resp.StatusCode, msg)
	}

	text, segments := parseTranscriptionResponse(payload)
	if strings.TrimSpace(text) == "" {
		return "", nil, fmt.Errorf("transcription response bos")
	}
	return text, segments, nil
}

func parseTranscriptionResponse(payload []byte) (string, []TranscribeSegment) {
	trimmed := strings.TrimSpace(string(payload))
	if trimmed == "" {
		return "", nil
	}

	var response struct {
		Text       string `json:"text"`
		Transcript string `json:"transcript"`
		Segments   []struct {
			Start float64 `json:"start"`
			End   float64 `json:"end"`
			Text  string  `json:"text"`
		} `json:"segments"`
	}
	if json.Unmarshal(payload, &response) == nil {
		segments := make([]TranscribeSegment, 0, len(response.Segments))
		for _, segment := range response.Segments {
			if strings.TrimSpace(segment.Text) == "" {
				continue
			}
			segments = append(segments, TranscribeSegment{
				Start: segment.Start,
				End:   segment.End,
				Text:  strings.TrimSpace(segment.Text),
			})
		}
		text := strings.TrimSpace(response.Text)
		if text == "" {
			text = strings.TrimSpace(response.Transcript)
		}
		if text == "" && len(segments) > 0 {
			lines := make([]string, 0, len(segments))
			for _, segment := range segments {
				lines = append(lines, segment.Text)
			}
			text = strings.Join(lines, "\n")
		}
		if text != "" {
			return text, segments
		}
	}

	return trimmed, nil
}

func runSummaryRequest(req SummarizeTranscriptRequest, transcriptText string) (string, error) {
	provider := normalizeAIProviderName(req.Provider)
	baseURL := providerDefaultBaseURL(provider, req.BaseURL)
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = defaultSummaryModel(provider)
	}
	apiKey := resolveProviderAPIKey(provider, req.APIKey)
	if providerNeedsAPIKey(provider, baseURL) && apiKey == "" {
		return "", fmt.Errorf("api key gerekli")
	}

	systemPrompt := buildSummarySystemPrompt(req.Style, req.TargetLanguage)
	requestBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": transcriptText},
		},
		"temperature": 0.2,
	}
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(apiKey) != "" {
		httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if readErr != nil {
		return "", readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return "", fmt.Errorf("provider http %d: %s", resp.StatusCode, msg)
	}

	summary := parseSummaryResponse(body)
	if strings.TrimSpace(summary) == "" {
		return "", fmt.Errorf("summary response bos")
	}
	return summary, nil
}

func parseSummaryResponse(payload []byte) string {
	var response struct {
		Choices []struct {
			Text    string `json:"text"`
			Message struct {
				Content any `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if json.Unmarshal(payload, &response) != nil || len(response.Choices) == 0 {
		return strings.TrimSpace(string(payload))
	}

	choice := response.Choices[0]
	if text := strings.TrimSpace(choice.Text); text != "" {
		return text
	}
	return strings.TrimSpace(extractMessageContent(choice.Message.Content))
}

func extractMessageContent(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			obj, ok := item.(map[string]any)
			if !ok {
				continue
			}
			text, _ := obj["text"].(string)
			text = strings.TrimSpace(text)
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

func renderSummaryMarkdown(summaryText string, style string, language string) string {
	var b strings.Builder
	b.WriteString("# Transcript Summary\n\n")

	if strings.TrimSpace(style) != "" || strings.TrimSpace(language) != "" {
		meta := make([]string, 0, 2)
		if strings.TrimSpace(style) != "" {
			meta = append(meta, "style="+strings.TrimSpace(style))
		}
		if strings.TrimSpace(language) != "" {
			meta = append(meta, "language="+strings.TrimSpace(language))
		}
		if len(meta) > 0 {
			b.WriteString("_")
			b.WriteString(strings.Join(meta, " | "))
			b.WriteString("_\n\n")
		}
	}

	b.WriteString(strings.TrimSpace(summaryText))
	b.WriteString("\n")
	return b.String()
}

func buildSummarySystemPrompt(style string, targetLanguage string) string {
	instructions := []string{
		"Transcript metnini ozetle.",
		"Ciktiyi yalnizca duz metin olarak ver.",
		"Varsayim ekleme; sadece transcriptteki bilgiye dayan.",
	}
	if strings.TrimSpace(style) != "" {
		instructions = append(instructions, "Stil: "+strings.TrimSpace(style))
	}
	if strings.TrimSpace(targetLanguage) != "" {
		instructions = append(instructions, "Dil: "+strings.TrimSpace(targetLanguage))
	}
	return strings.Join(instructions, " ")
}

func normalizeAIProviderName(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case aiProviderOpenAI:
		return aiProviderOpenAI
	case aiProviderOllama:
		return aiProviderOllama
	case "openai_compatible", "openai-compatible", "compatible", "openai compatible":
		return aiProviderOpenAICompatible
	default:
		return aiProviderOpenAICompatible
	}
}

func providerDefaultBaseURL(provider string, provided string) string {
	if v := strings.TrimSpace(provided); v != "" {
		return strings.TrimRight(v, "/")
	}
	switch normalizeAIProviderName(provider) {
	case aiProviderOllama:
		return "http://localhost:11434/v1"
	case aiProviderOpenAI:
		return "https://api.openai.com/v1"
	default:
		return "https://api.openai.com/v1"
	}
}

func defaultTranscribeModel(provider string) string {
	switch normalizeAIProviderName(provider) {
	case aiProviderOllama:
		return "whisper"
	default:
		return "whisper-1"
	}
}

func defaultSummaryModel(provider string) string {
	switch normalizeAIProviderName(provider) {
	case aiProviderOllama:
		return "qwen2.5:7b"
	default:
		return "gpt-4o-mini"
	}
}

func providerNeedsAPIKey(provider string, baseURL string) bool {
	p := normalizeAIProviderName(provider)
	if p == aiProviderOllama {
		return false
	}
	if p == aiProviderOpenAI {
		return true
	}
	base := strings.ToLower(strings.TrimSpace(baseURL))
	if strings.HasPrefix(base, "http://localhost") || strings.HasPrefix(base, "http://127.0.0.1") {
		return false
	}
	return true
}

func resolveProviderAPIKey(provider string, explicit string) string {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit)
	}
	for _, envKey := range envKeysForAIProvider(provider) {
		if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
			return value
		}
	}
	return ""
}

func envKeysForAIProvider(provider string) []string {
	switch normalizeAIProviderName(provider) {
	case aiProviderOpenAI:
		return []string{"OPENAI_API_KEY", "DOCUFY_AI_API_KEY"}
	case aiProviderOllama:
		return []string{"OLLAMA_API_KEY", "DOCUFY_AI_API_KEY"}
	default:
		return []string{"DOCUFY_AI_API_KEY", "OPENAI_API_KEY"}
	}
}
