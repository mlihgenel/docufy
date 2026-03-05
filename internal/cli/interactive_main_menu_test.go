package cli

import "testing"

func TestNewInteractiveModelTopLevelSections(t *testing.T) {
	m := newInteractiveModel(nil, false)
	if m.state != stateMainMenu {
		t.Fatalf("expected initial stateMainMenu, got %v", m.state)
	}
	if len(m.choices) != len(topLevelSections)+1 {
		t.Fatalf("expected %d top-level entries, got %d", len(topLevelSections)+1, len(m.choices))
	}
	if m.choices[0] != "Dönüştürme" {
		t.Fatalf("unexpected first top-level entry: %s", m.choices[0])
	}
}

func TestMainMenuSectionTransition(t *testing.T) {
	m := newInteractiveModel(nil, false)
	m.state = stateMainMenu
	m.cursor = 0

	nextModel, cmd := m.handleEnter()
	if cmd != nil {
		t.Fatalf("expected no command for section transition")
	}
	next, ok := nextModel.(interactiveModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if next.state != stateMainSectionMenu {
		t.Fatalf("expected stateMainSectionMenu, got %v", next.state)
	}
	if next.mainSection != "conversion" {
		t.Fatalf("expected conversion section, got %s", next.mainSection)
	}
	if len(next.choices) == 0 || next.choices[0] != "Tek Dosya Dönüştür" {
		t.Fatalf("unexpected section choices: %+v", next.choices)
	}
}

func TestMainSectionActionVideoTrim(t *testing.T) {
	m := newInteractiveModel(nil, false)
	m = m.goToMainSection("video")
	m.cursor = 0

	nextModel, cmd := m.handleEnter()
	if cmd != nil {
		t.Fatalf("expected no command for video trim menu action")
	}
	next, ok := nextModel.(interactiveModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if !next.flowVideoTrim {
		t.Fatalf("expected flowVideoTrim true")
	}
	if next.state != stateFileBrowser {
		t.Fatalf("expected stateFileBrowser, got %v", next.state)
	}
}

func TestMainSectionActionAIAssistant(t *testing.T) {
	m := newInteractiveModel(nil, false)
	m = m.goToMainSection("ai")
	m.cursor = 0

	nextModel, cmd := m.handleEnter()
	if cmd != nil {
		t.Fatalf("expected no command for AI assistant menu action")
	}
	next, ok := nextModel.(interactiveModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if next.state != stateAIIntro {
		t.Fatalf("expected stateAIIntro, got %v", next.state)
	}
	if len(next.choices) == 0 || next.choices[0] != "AI Oturumunu Başlat" {
		t.Fatalf("unexpected AI intro choices: %+v", next.choices)
	}
}

func TestAIIntroTransitionRequiresAuthByDefault(t *testing.T) {
	t.Setenv("DOCUFY_AI_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("DOCUFY_AI_DISABLE_KEYCHAIN", "1")
	m := newInteractiveModel(nil, false).goToAIIntro()
	m.aiProvider = aiProviderOpenAI
	m.aiAPIKey = ""
	m.cursor = 0

	nextModel, cmd := m.handleEnter()
	if cmd != nil {
		t.Fatalf("expected no command for AI intro -> chat transition")
	}
	next, ok := nextModel.(interactiveModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if next.state != stateAIAuthInput {
		t.Fatalf("expected stateAIAuthInput, got %v", next.state)
	}
}

func TestAIIntroTransitionToChatWithEnvKey(t *testing.T) {
	t.Setenv("DOCUFY_AI_API_KEY", "test-key")
	t.Setenv("DOCUFY_AI_DISABLE_KEYCHAIN", "1")
	m := newInteractiveModel(nil, false).goToAIIntro()
	m.cursor = 0

	nextModel, cmd := m.handleEnter()
	if cmd != nil {
		t.Fatalf("expected no command for AI intro -> chat transition")
	}
	next, ok := nextModel.(interactiveModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if next.state != stateAIChat {
		t.Fatalf("expected stateAIChat, got %v", next.state)
	}
	if len(next.choices) == 0 || next.choices[0] != "Komut Yaz" {
		t.Fatalf("unexpected AI chat choices: %+v", next.choices)
	}
}

func TestAIIntroTransitionToProviderSettings(t *testing.T) {
	m := newInteractiveModel(nil, false).goToAIIntro()
	m.cursor = 1

	nextModel, cmd := m.handleEnter()
	if cmd != nil {
		t.Fatalf("expected no command for AI intro -> provider settings transition")
	}
	next, ok := nextModel.(interactiveModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if next.state != stateAIAuthProviderSelect {
		t.Fatalf("expected stateAIAuthProviderSelect, got %v", next.state)
	}
	if len(next.choices) == 0 || next.choices[0] != "OpenAI" {
		t.Fatalf("unexpected provider choices: %+v", next.choices)
	}
}

func TestAICommandInputTransitionsToPlanConfirm(t *testing.T) {
	m := newInteractiveModel(nil, false)
	m = m.goToAICommandInput()
	m.aiCurrentFile = "/tmp/sample.mp4"
	m.aiPromptInput = "Bu videoyu mp4 olarak dönüştür"

	nextModel, cmd := m.handleEnter()
	if cmd != nil {
		t.Fatalf("expected no command before plan confirmation")
	}
	next, ok := nextModel.(interactiveModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if next.state != stateAIPlanConfirm {
		t.Fatalf("expected stateAIPlanConfirm, got %v", next.state)
	}
	if next.aiPendingPrompt == "" {
		t.Fatalf("expected pending prompt to be set")
	}
	if len(next.choices) == 0 || next.choices[0] != "Onayla ve Çalıştır" {
		t.Fatalf("unexpected plan confirm choices: %+v", next.choices)
	}
}

func TestAIPlanConfirmApproveStartsExecution(t *testing.T) {
	m := newInteractiveModel(nil, false)
	m = m.goToAIPlanConfirm()
	m.aiPendingPrompt = "bilgi ver"
	m.aiCurrentFile = "/tmp/sample.txt"
	m.cursor = 0

	nextModel, cmd := m.handleEnter()
	if cmd == nil {
		t.Fatalf("expected command execution cmd after confirmation")
	}
	next, ok := nextModel.(interactiveModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if next.state != stateAIExecuting {
		t.Fatalf("expected stateAIExecuting, got %v", next.state)
	}
}

func TestAIPlanConfirmRiskRequiresSecondConfirmation(t *testing.T) {
	m := newInteractiveModel(nil, false)
	m = m.goToAIPlanConfirm()
	m.aiPendingPrompt = "dosyayı üzerine yaz ve overwrite et"
	m.aiNeedsRiskAck = true
	m.cursor = 0

	nextModel, cmd := m.handleEnter()
	if cmd != nil {
		t.Fatalf("expected no command before risk ack")
	}
	next, ok := nextModel.(interactiveModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if next.state != stateAIRiskConfirmInput {
		t.Fatalf("expected stateAIRiskConfirmInput, got %v", next.state)
	}
}

func TestAIRiskConfirmInputApprovesExecution(t *testing.T) {
	m := newInteractiveModel(nil, false)
	m.state = stateAIRiskConfirmInput
	m.aiPendingPrompt = "dosya bilgisi ver"
	m.aiCurrentFile = "/tmp/a.txt"
	m.aiRiskAckInput = "ONAY"

	nextModel, cmd := m.handleEnter()
	if cmd == nil {
		t.Fatalf("expected command after valid risk confirmation")
	}
	next, ok := nextModel.(interactiveModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if next.state != stateAIExecuting {
		t.Fatalf("expected stateAIExecuting, got %v", next.state)
	}
}
