package entrypoint

import "runtime/debug"

func readBuildInfo() (*debug.BuildInfo, bool) {
	return debug.ReadBuildInfo()
}

func buildSetting(info *debug.BuildInfo, key string) string {
	for _, setting := range info.Settings {
		if setting.Key == key {
			return setting.Value
		}
	}
	return ""
}
