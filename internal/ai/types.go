package ai

import (
	"errors"

	"github.com/mlihgenel/docufy/v2/internal/converter"
)

const (
	StatusSuccess        = "success"
	StatusSkipped        = "skipped"
	StatusFailed         = "failed"
	StatusNotImplemented = "not_implemented"
)

var ErrNotImplemented = errors.New("tool behavior is not implemented yet")

type ConvertFileRequest struct {
	InputPath    string `json:"input_path"`
	To           string `json:"to"`
	OutputPath   string `json:"output_path,omitempty"`
	OutputDir    string `json:"output_dir,omitempty"`
	Name         string `json:"name,omitempty"`
	Quality      int    `json:"quality,omitempty"`
	MetadataMode string `json:"metadata_mode,omitempty"`
	OnConflict   string `json:"on_conflict,omitempty"`
}

type TrimVideoRequest struct {
	InputPath    string `json:"input_path"`
	Mode         string `json:"mode,omitempty"` // clip|remove
	Start        string `json:"start,omitempty"`
	End          string `json:"end,omitempty"`
	Duration     string `json:"duration,omitempty"`
	To           string `json:"to,omitempty"`
	OutputPath   string `json:"output_path,omitempty"`
	OutputDir    string `json:"output_dir,omitempty"`
	Name         string `json:"name,omitempty"`
	Codec        string `json:"codec,omitempty"` // auto|copy|reencode
	Quality      int    `json:"quality,omitempty"`
	MetadataMode string `json:"metadata_mode,omitempty"`
	OnConflict   string `json:"on_conflict,omitempty"`
}

type ExtractAudioRequest struct {
	InputPath    string `json:"input_path"`
	To           string `json:"to,omitempty"`
	OutputPath   string `json:"output_path,omitempty"`
	OutputDir    string `json:"output_dir,omitempty"`
	Name         string `json:"name,omitempty"`
	Quality      int    `json:"quality,omitempty"`
	Copy         bool   `json:"copy,omitempty"`
	MetadataMode string `json:"metadata_mode,omitempty"`
	OnConflict   string `json:"on_conflict,omitempty"`
}

type FileInfoRequest struct {
	Path string `json:"path"`
}

type TranscribeSegment struct {
	Start float64 `json:"start,omitempty"`
	End   float64 `json:"end,omitempty"`
	Text  string  `json:"text,omitempty"`
}

type TranscribeMediaRequest struct {
	InputPath   string `json:"input_path"`
	Provider    string `json:"provider,omitempty"`
	Model       string `json:"model,omitempty"`
	BaseURL     string `json:"base_url,omitempty"`
	APIKey      string `json:"api_key,omitempty"`
	Language    string `json:"language,omitempty"`
	Diarization bool   `json:"diarization,omitempty"`
	OutputPath  string `json:"output_path,omitempty"`
	OutputDir   string `json:"output_dir,omitempty"`
	Name        string `json:"name,omitempty"`
	OnConflict  string `json:"on_conflict,omitempty"`
}

type SummarizeTranscriptRequest struct {
	TranscriptPath string `json:"transcript_path"`
	Provider       string `json:"provider,omitempty"`
	Model          string `json:"model,omitempty"`
	BaseURL        string `json:"base_url,omitempty"`
	APIKey         string `json:"api_key,omitempty"`
	Style          string `json:"style,omitempty"`
	TargetLanguage string `json:"target_language,omitempty"`
	OutputPath     string `json:"output_path,omitempty"`
	OutputDir      string `json:"output_dir,omitempty"`
	Name           string `json:"name,omitempty"`
	OnConflict     string `json:"on_conflict,omitempty"`
}

type ToolResult struct {
	Status             string              `json:"status"`
	Tool               string              `json:"tool"`
	InputPath          string              `json:"input_path,omitempty"`
	OutputPath         string              `json:"output_path,omitempty"`
	TranscriptPath     string              `json:"transcript_path,omitempty"`
	TranscriptSegments []TranscribeSegment `json:"segments,omitempty"`
	SummaryPath        string              `json:"summary_path,omitempty"`
	SummaryText        string              `json:"summary_text,omitempty"`
	Message            string              `json:"message,omitempty"`
	Error              string              `json:"error,omitempty"`
	FileInfo           *converter.FileInfo `json:"file_info,omitempty"`
}
