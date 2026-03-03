package cli

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mlihgenel/docufy/v2/internal/converter"
	"github.com/spf13/cobra"
)

const (
	envOutput     = "DOCUFY_OUTPUT"
	envWorkers    = "DOCUFY_WORKERS"
	envQuality    = "DOCUFY_QUALITY"
	envProfile    = "DOCUFY_PROFILE"
	envConflict   = "DOCUFY_ON_CONFLICT"
	envMetadata   = "DOCUFY_METADATA"
	envRetry      = "DOCUFY_RETRY"
	envRetryDelay = "DOCUFY_RETRY_DELAY"
	envReport     = "DOCUFY_REPORT"

	legacyEnvOutput     = "FILECONVERTER_OUTPUT"
	legacyEnvWorkers    = "FILECONVERTER_WORKERS"
	legacyEnvQuality    = "FILECONVERTER_QUALITY"
	legacyEnvProfile    = "FILECONVERTER_PROFILE"
	legacyEnvConflict   = "FILECONVERTER_ON_CONFLICT"
	legacyEnvMetadata   = "FILECONVERTER_METADATA"
	legacyEnvRetry      = "FILECONVERTER_RETRY"
	legacyEnvRetryDelay = "FILECONVERTER_RETRY_DELAY"
	legacyEnvReport     = "FILECONVERTER_REPORT"
)

func applyRootDefaults(cmd *cobra.Command) error {
	if !cmd.Flags().Changed("output") {
		if v := readEnvStringCompat(envOutput, legacyEnvOutput); v != "" {
			outputDir = v
		} else if activeProjectConfig != nil && strings.TrimSpace(activeProjectConfig.DefaultOutput) != "" {
			outputDir = strings.TrimSpace(activeProjectConfig.DefaultOutput)
		}
	}

	if !cmd.Flags().Changed("workers") {
		if v, ok := readEnvIntCompat(envWorkers, legacyEnvWorkers); ok && v > 0 {
			workers = v
		} else if activeProjectConfig != nil && activeProjectConfig.Workers > 0 {
			workers = activeProjectConfig.Workers
		}
	}

	normalizedOutputFormat := NormalizeOutputFormat(outputFormat)
	if normalizedOutputFormat == "" {
		return outputFormatError(outputFormat)
	}
	outputFormat = normalizedOutputFormat

	return nil
}

func applyQualityDefault(cmd *cobra.Command, flagName string, value *int) {
	if cmd.Flags().Changed(flagName) {
		return
	}
	if v, ok := readEnvIntCompat(envQuality, legacyEnvQuality); ok && v >= 0 {
		*value = v
		return
	}
	if activeProjectConfig != nil && activeProjectConfig.Quality > 0 {
		*value = activeProjectConfig.Quality
	}
}

func applyOnConflictDefault(cmd *cobra.Command, flagName string, value *string) {
	if cmd.Flags().Changed(flagName) {
		return
	}
	if v := readEnvStringCompat(envConflict, legacyEnvConflict); v != "" {
		*value = strings.ToLower(v)
		return
	}
	if activeProjectConfig != nil && strings.TrimSpace(activeProjectConfig.OnConflict) != "" {
		*value = strings.ToLower(strings.TrimSpace(activeProjectConfig.OnConflict))
	}
}

func applyProfileDefault(cmd *cobra.Command, flagName string, value *string) {
	if cmd.Flags().Changed(flagName) {
		return
	}
	if v := readEnvStringCompat(envProfile, legacyEnvProfile); v != "" {
		*value = v
		return
	}
	if activeProjectConfig != nil && strings.TrimSpace(activeProjectConfig.Profile) != "" {
		*value = strings.TrimSpace(activeProjectConfig.Profile)
	}
}

func applyMetadataDefault(cmd *cobra.Command, preserveFlag string, preserveValue *bool, stripFlag string, stripValue *bool) {
	if cmd.Flags().Changed(preserveFlag) || cmd.Flags().Changed(stripFlag) {
		return
	}

	mode := readEnvStringCompat(envMetadata, legacyEnvMetadata)
	if mode == "" && activeProjectConfig != nil {
		mode = strings.TrimSpace(activeProjectConfig.MetadataMode)
	}
	mode = converter.NormalizeMetadataMode(mode)
	switch mode {
	case converter.MetadataPreserve:
		*preserveValue = true
		*stripValue = false
	case converter.MetadataStrip:
		*stripValue = true
		*preserveValue = false
	}
}

func applyRetryDefaults(cmd *cobra.Command, retryFlag string, retryValue *int, delayFlag string, delayValue *time.Duration) {
	if !cmd.Flags().Changed(retryFlag) {
		if v, ok := readEnvIntCompat(envRetry, legacyEnvRetry); ok && v >= 0 {
			*retryValue = v
		} else if activeProjectConfig != nil && activeProjectConfig.Retry > 0 {
			*retryValue = activeProjectConfig.Retry
		}
	}

	if !cmd.Flags().Changed(delayFlag) {
		if v, ok := readEnvDurationCompat(envRetryDelay, legacyEnvRetryDelay); ok {
			*delayValue = v
		} else if activeProjectConfig != nil && activeProjectConfig.RetryDelay > 0 {
			*delayValue = activeProjectConfig.RetryDelay
		}
	}
}

func applyReportDefault(cmd *cobra.Command, flagName string, value *string) {
	if cmd.Flags().Changed(flagName) {
		return
	}
	if v := readEnvStringCompat(envReport, legacyEnvReport); v != "" {
		*value = strings.ToLower(v)
		return
	}
	if activeProjectConfig != nil && strings.TrimSpace(activeProjectConfig.ReportFormat) != "" {
		*value = strings.ToLower(strings.TrimSpace(activeProjectConfig.ReportFormat))
	}
}

func readEnvStringCompat(name string, legacyName string) string {
	if raw := strings.TrimSpace(os.Getenv(name)); raw != "" {
		return raw
	}
	if legacyName == "" {
		return ""
	}
	return strings.TrimSpace(os.Getenv(legacyName))
}

func readEnvIntCompat(name string, legacyName string) (int, bool) {
	if v, ok := readEnvInt(name); ok {
		return v, true
	}
	if legacyName == "" {
		return 0, false
	}
	return readEnvInt(legacyName)
}

func readEnvDurationCompat(name string, legacyName string) (time.Duration, bool) {
	if v, ok := readEnvDuration(name); ok {
		return v, true
	}
	if legacyName == "" {
		return 0, false
	}
	return readEnvDuration(legacyName)
}

func readEnvInt(name string) (int, bool) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return 0, false
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return v, true
}

func readEnvDuration(name string) (time.Duration, bool) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return 0, false
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, false
	}
	return v, true
}
