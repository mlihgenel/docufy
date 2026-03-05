package secret

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	ErrNotSupported = errors.New("keychain backend is not supported on this platform")
	ErrDisabled     = errors.New("keychain integration is disabled")
)

const aiKeychainService = "com.docufy.ai"

func SaveAIAPIKey(provider string, key string) error {
	if keychainDisabled() {
		return ErrDisabled
	}

	normalizedProvider := normalizeProvider(provider)
	if normalizedProvider == "" {
		return fmt.Errorf("provider zorunlu")
	}

	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return fmt.Errorf("api key bos olamaz")
	}

	switch runtime.GOOS {
	case "darwin":
		return saveDarwin(aiKeychainAccount(normalizedProvider), trimmedKey)
	case "linux":
		return saveLinux(aiKeychainAccount(normalizedProvider), trimmedKey)
	case "windows":
		return saveWindows(aiKeychainAccount(normalizedProvider), trimmedKey)
	default:
		return ErrNotSupported
	}
}

func LoadAIAPIKey(provider string) (string, error) {
	if keychainDisabled() {
		return "", ErrDisabled
	}

	normalizedProvider := normalizeProvider(provider)
	if normalizedProvider == "" {
		return "", fmt.Errorf("provider zorunlu")
	}

	switch runtime.GOOS {
	case "darwin":
		return loadDarwin(aiKeychainAccount(normalizedProvider))
	case "linux":
		return loadLinux(aiKeychainAccount(normalizedProvider))
	case "windows":
		return loadWindows(aiKeychainAccount(normalizedProvider))
	default:
		return "", ErrNotSupported
	}
}

func keychainDisabled() bool {
	v := strings.TrimSpace(os.Getenv("DOCUFY_AI_DISABLE_KEYCHAIN"))
	if v == "" {
		return false
	}
	v = strings.ToLower(v)
	return v == "1" || v == "true" || v == "yes"
}

func normalizeProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai":
		return "openai"
	case "ollama":
		return "ollama"
	case "openai-compatible", "openai_compatible", "compatible", "openai compatible":
		return "openai-compatible"
	default:
		return ""
	}
}

func aiKeychainAccount(provider string) string {
	return "ai:" + provider
}

func saveDarwin(account string, key string) error {
	if _, err := exec.LookPath("security"); err != nil {
		return ErrNotSupported
	}

	cmd := exec.Command(
		"security",
		"add-generic-password",
		"-a", account,
		"-s", aiKeychainService,
		"-w", key,
		"-U",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("macOS keychain kaydetme hatasi: %s", msg)
	}
	return nil
}

func loadDarwin(account string) (string, error) {
	if _, err := exec.LookPath("security"); err != nil {
		return "", ErrNotSupported
	}

	cmd := exec.Command(
		"security",
		"find-generic-password",
		"-a", account,
		"-s", aiKeychainService,
		"-w",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if strings.Contains(strings.ToLower(msg), "could not be found") {
			return "", nil
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("macOS keychain okuma hatasi: %s", msg)
	}
	return strings.TrimSpace(string(out)), nil
}

func saveLinux(account string, key string) error {
	if _, err := exec.LookPath("secret-tool"); err != nil {
		return ErrNotSupported
	}

	cmd := exec.Command(
		"secret-tool",
		"store",
		"--label=Docufy AI API Key",
		"service", aiKeychainService,
		"account", account,
	)
	cmd.Stdin = strings.NewReader(key)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("linux secret-service kaydetme hatasi: %s", msg)
	}
	return nil
}

func loadLinux(account string) (string, error) {
	if _, err := exec.LookPath("secret-tool"); err != nil {
		return "", ErrNotSupported
	}

	cmd := exec.Command(
		"secret-tool",
		"lookup",
		"service", aiKeychainService,
		"account", account,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitCode(err) == 1 {
			return "", nil
		}
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("linux secret-service okuma hatasi: %s", msg)
	}
	return strings.TrimSpace(string(out)), nil
}

func saveWindows(account string, key string) error {
	powershell, err := resolvePowerShellPath()
	if err != nil {
		return ErrNotSupported
	}

	path, err := windowsSecretFilePath(account)
	if err != nil {
		return err
	}
	if mkErr := os.MkdirAll(filepath.Dir(path), 0700); mkErr != nil {
		return mkErr
	}

	script := fmt.Sprintf(
		"$secure = ConvertTo-SecureString '%s' -AsPlainText -Force; $enc = ConvertFrom-SecureString $secure; Set-Content -LiteralPath '%s' -Value $enc -NoNewline",
		psSingleQuote(key),
		psSingleQuote(path),
	)
	out, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("windows dpapi kaydetme hatasi: %s", msg)
	}
	return nil
}

func loadWindows(account string) (string, error) {
	powershell, err := resolvePowerShellPath()
	if err != nil {
		return "", ErrNotSupported
	}

	path, err := windowsSecretFilePath(account)
	if err != nil {
		return "", err
	}

	script := fmt.Sprintf(
		"if (!(Test-Path -LiteralPath '%s')) { exit 4 }; $enc = Get-Content -LiteralPath '%s' -Raw; if ([string]::IsNullOrWhiteSpace($enc)) { exit 4 }; $secure = ConvertTo-SecureString $enc; $bstr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure); try { [Runtime.InteropServices.Marshal]::PtrToStringBSTR($bstr) } finally { [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr) }",
		psSingleQuote(path),
		psSingleQuote(path),
	)
	out, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).CombinedOutput()
	if err != nil {
		if exitCode(err) == 4 {
			return "", nil
		}
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("windows dpapi okuma hatasi: %s", msg)
	}
	return strings.TrimSpace(string(out)), nil
}

func resolvePowerShellPath() (string, error) {
	if p, err := exec.LookPath("pwsh"); err == nil {
		return p, nil
	}
	if p, err := exec.LookPath("powershell"); err == nil {
		return p, nil
	}
	return "", ErrNotSupported
}

func windowsSecretFilePath(account string) (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	safeName := strings.ReplaceAll(strings.TrimSpace(account), ":", "_")
	if safeName == "" {
		return "", fmt.Errorf("gecersiz account")
	}
	return filepath.Join(cfgDir, "docufy", "secrets", safeName+".dpapi"), nil
}

func psSingleQuote(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}
