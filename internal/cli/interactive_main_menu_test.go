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
	if next.state != stateAIAuthInput {
		t.Fatalf("expected stateAIAuthInput, got %v", next.state)
	}
}

func TestAIIntroTransitionToChatWithEnvKey(t *testing.T) {
	t.Setenv("DOCUFY_AI_API_KEY", "test-key")
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
