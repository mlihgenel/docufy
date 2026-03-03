package entrypoint

import (
	"strings"

	"github.com/mlihgenel/docufy/v2/internal/cli"
)

// Run uygulama giriş akışını ortaklaştırır; farklı main paketleri bunu çağırabilir.
func Run(version, buildDate string) int {
	cli.SetVersionInfo(resolveVersion(version), resolveBuildDate(buildDate))
	if err := cli.Execute(); err != nil {
		return 1
	}
	return 0
}

func resolveVersion(version string) string {
	v := normalizeVersion(version)
	if v != "" && v != "dev" {
		return v
	}

	info, ok := readBuildInfo()
	if !ok {
		return "dev"
	}
	if moduleVersion := normalizeVersion(info.Main.Version); moduleVersion != "" && moduleVersion != "(devel)" {
		return moduleVersion
	}

	revision := strings.TrimSpace(buildSetting(info, "vcs.revision"))
	if revision == "" {
		return "dev"
	}
	if len(revision) > 7 {
		revision = revision[:7]
	}
	if strings.EqualFold(strings.TrimSpace(buildSetting(info, "vcs.modified")), "true") {
		return "dev-" + revision + "-dirty"
	}
	return "dev-" + revision
}

func resolveBuildDate(buildDate string) string {
	if v := strings.TrimSpace(buildDate); v != "" {
		return v
	}
	info, ok := readBuildInfo()
	if !ok {
		return ""
	}
	return strings.TrimSpace(buildSetting(info, "vcs.time"))
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, "v") {
		return strings.TrimPrefix(v, "v")
	}
	return v
}
