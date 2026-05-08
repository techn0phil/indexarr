package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"indexarr/internal/models"
)

// MediainfoOutput represents the JSON output from mediainfo CLI
type MediainfoOutput struct {
	Media struct {
		Ref   string           `json:"@ref"`
		Track []MediainfoTrack `json:"track"`
	} `json:"media"`
}

type MediainfoTrack struct {
	// General track fields
	FileSize string `json:"FileSize"`

	// Common fields for all track types
	Type      string `json:"@type"`
	Format    string `json:"Format"`
	CodecID   string `json:"CodecID"`
	Duration  string `json:"Duration"`
	FrameRate string `json:"FrameRate"`
	BitRate   string `json:"BitRate"`
	Default   string `json:"Default"`
	Forced    string `json:"Forced"`

	// Video-specific fields
	FormatProfile     string `json:"Format_profile"`
	HDRFormat         string `json:"HDR_Format"`
	HDRFormatComp     string `json:"HDR_Format_Compatibility"`
	Width             string `json:"Width"`
	Height            string `json:"Height"`
	FrameRateMode     string `json:"FrameRate_Mode"`
	ColorSpace        string `json:"ColorSpace"`
	ChromaSubsampling string `json:"ChromaSubsampling"`
	BitDepth          string `json:"BitDepth"`
	ColorPrimaries    string `json:"colour_primaries"`

	// Audio-specific fields
	FormatCommercial         string `json:"Format_Commercial_IfAny"`
	FormatAdditionalFeatures string `json:"Format_AdditionalFeatures"`
	BitRateMode              string `json:"BitRate_Mode"`
	Channels                 string `json:"Channels"`
	ChannelLayout            string `json:"ChannelLayout"`
	SamplingRate             string `json:"SamplingRate"`
	CompressionMode          string `json:"Compression_Mode"`

	// Audio and Subtitle-specific fields
	Language string `json:"Language"`
	Title    string `json:"Title"`
}

// Extractor handles mediainfo extraction from files
type Extractor struct {
	mediainfoPath string
	timeout       time.Duration
}

// NewExtractor creates a new mediainfo extractor
func NewExtractor(mediainfoPath string, timeoutSeconds int) *Extractor {
	return &Extractor{
		mediainfoPath: mediainfoPath,
		timeout:       time.Duration(timeoutSeconds) * time.Second,
	}
}

// Extract runs mediainfo on a file and returns parsed MediaInfo
func (e *Extractor) Extract(filePath string) (*models.MediaInfo, int64, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.mediainfoPath, "--Output=JSON", filePath)
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, 0, 0, fmt.Errorf("mediainfo timed out after %v", e.timeout)
		}
		return nil, 0, 0, fmt.Errorf("mediainfo failed: %w", err)
	}

	var mi MediainfoOutput
	if err := json.Unmarshal(output, &mi); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to parse mediainfo JSON: %w", err)
	}

	return e.parseMediaInfo(&mi)
}

// parseMediaInfo converts MediainfoOutput to models.MediaInfo
func (e *Extractor) parseMediaInfo(mi *MediainfoOutput) (*models.MediaInfo, int64, int, error) {
	info := &models.MediaInfo{
		VideoTracks:    []models.VideoTrack{},
		AudioTracks:    []models.AudioTrack{},
		SubtitleTracks: []models.SubtitleTrack{},
	}

	var fileSize int64
	var duration int // seconds

	for _, track := range mi.Media.Track {
		switch track.Type {
		case "General":
			fileSize = parseFileSize(track.FileSize)
			duration = parseDuration(track.Duration)

		case "Video":
			info.VideoTracks = append(info.VideoTracks, models.VideoTrack{
				Codec:      e.parseVideoCodec(track),
				Resolution: e.parseResolution(track),
				FPS:        parseFloat(track.FrameRate),
				Bitrate:    e.formatBitrate(track.BitRate),
				HDR:        e.parseHDR(track),
				ColorSpace: e.parseColorSpace(track),
			})

		case "Audio":
			info.AudioTracks = append(info.AudioTracks, models.AudioTrack{
				Codec:      e.parseAudioCodec(track),
				Channels:   e.parseChannels(track),
				Language:   e.formatLanguage(track.Language),
				SampleRate: e.formatSampleRate(track.SamplingRate),
				Bitrate:    e.formatBitrate(track.BitRate),
				Default:    track.Default,
			})

		case "Text":
			info.SubtitleTracks = append(info.SubtitleTracks, models.SubtitleTrack{
				Language: e.formatLanguage(track.Language),
				Format:   track.Format,
				Forced:   track.Forced,
				Default:  track.Default,
			})
		}
	}

	return info, fileSize, duration, nil
}

func (e *Extractor) parseVideoCodec(track MediainfoTrack) string {
	format := track.Format

	if format == "" {
		return "Unknown"
	}

	// Normalize common codec names
	switch strings.ToUpper(format) {
	case "AVC", "H264":
		return "H.264"
	case "HEVC", "H265":
		return "H.265"
	case "VP9":
		return "VP9"
	case "AV1":
		return "AV1"
	case "MPEG-4 VISUAL":
		return "MPEG-4"
	default:
		return format
	}
}

func (e *Extractor) parseResolution(track MediainfoTrack) string {
	width := strings.TrimSpace(track.Width)
	height := strings.TrimSpace(track.Height)

	if width == "" || height == "" {
		return "Unknown"
	}

	// Remove any non-numeric suffixes
	width = strings.Split(width, " ")[0]
	height = strings.Split(height, " ")[0]
	return fmt.Sprintf("%sx%s", width, height)
}

func (e *Extractor) parseHDR(track MediainfoTrack) string {
	hdrFormat := track.HDRFormat
	hdrCompat := track.HDRFormatComp

	// Check for Dolby Vision
	if strings.Contains(strings.ToLower(hdrFormat), "dolby vision") {
		if (strings.Contains(strings.ToLower(hdrCompat), "hdr10+")) &&
			track.BitDepth == "10" {

			return "Dolby Vision / HDR10+"
		}
		if (strings.Contains(strings.ToLower(hdrCompat), "hdr10")) &&
			track.BitDepth == "10" {

			return "Dolby Vision / HDR10"
		}

		return "Dolby Vision"
	}

	// Check for HDR10+
	if strings.Contains(strings.ToLower(hdrFormat), "hdr10+") ||
		strings.Contains(strings.ToLower(hdrCompat), "hdr10+") {
		return "HDR10+"
	}

	// Check for HDR10
	if strings.Contains(strings.ToLower(hdrFormat), "smpte st 2086") ||
		strings.Contains(strings.ToLower(hdrCompat), "hdr10") ||
		track.BitDepth == "10" {
		// Only mark as HDR10 if color primaries is BT.2020
		if e.parseColorSpace(track) == "BT.2020" {
			return "HDR10"
		}
	}

	return ""
}

func (e *Extractor) parseColorSpace(track MediainfoTrack) string {
	cp := track.ColorPrimaries

	if cp == "" {
		return "BT.709" // Default SDR color space
	}

	if strings.Contains(strings.ToLower(cp), "2020") {
		return "BT.2020"
	}
	if strings.Contains(strings.ToLower(cp), "709") {
		return "BT.709"
	}

	return cp
}

func (e *Extractor) parseAudioCodec(track MediainfoTrack) string {
	format := track.Format
	commercialName := track.FormatCommercial
	additionalFeatures := track.FormatAdditionalFeatures
	codec := track.CodecID
	title := track.Title

	if format == "" {
		return "Unknown"
	}

	// Handle TrueHD with Atmos
	if strings.Contains(strings.ToLower(commercialName), "truehd") ||
		strings.Contains(strings.ToLower(codec), "truehd") ||
		strings.Contains(strings.ToLower(title), "truehd") {

		if strings.Contains(strings.ToLower(commercialName), "atmos") ||
			strings.Contains(strings.ToLower(title), "atmos") {

			return "TrueHD Atmos"
		}

		return "TrueHD"
	}

	// Handle DTS variants
	if strings.Contains(strings.ToUpper(format), "DTS") ||
		strings.Contains(strings.ToUpper(codec), "DTS") {

		if strings.Contains(strings.ToLower(commercialName), "master audio") ||
			strings.Contains(strings.ToLower(commercialName), "ma") ||
			strings.Contains(strings.ToLower(title), "master audio") ||
			strings.Contains(strings.ToLower(title), "ma") {

			return "DTS-HD MA"
		}
		if strings.Contains(strings.ToUpper(commercialName), "X") ||
			strings.Contains(strings.ToLower(title), "X") {

			return "DTS:X"
		}

		return "DTS"
	}

	// Handle AC-3 / E-AC-3
	if strings.Contains(strings.ToUpper(format), "E-AC-3") ||
		strings.Contains(strings.ToUpper(format), "EAC3") {

		if strings.Contains(strings.ToUpper(additionalFeatures), "JOC") {
			return "E-AC-3 Atmos"
		}

		return "E-AC-3"
	}

	if strings.Contains(strings.ToUpper(format), "AC-3") ||
		strings.Contains(strings.ToUpper(format), "AC3") {

		return "AC-3"
	}

	// Handle AAC
	if strings.Contains(strings.ToUpper(format), "AAC") {
		return "AAC"
	}

	return format
}

func (e *Extractor) parseChannels(track MediainfoTrack) string {
	channels := track.Channels
	if channels == "" {
		return "2.0"
	}

	// Parse channel count
	ch, err := strconv.Atoi(strings.Split(channels, " ")[0])
	if err != nil {
		return channels
	}

	// Convert to surround format
	switch ch {
	case 1:
		return "1.0"
	case 2:
		return "2.0"
	case 6:
		return "5.1"
	case 8:
		return "7.1"
	default:
		return fmt.Sprintf("%d.0", ch)
	}
}

func (e *Extractor) formatLanguage(code string) string {
	if code == "" {
		return "Unknown"
	}

	switch strings.ToLower(code) {
	case "en":
		return "English"
	case "fr":
		return "French"
	case "es":
		return "Spanish"
	case "de":
		return "German"
	case "it":
		return "Italian"
	case "ja":
		return "Japanese"
	case "ko":
		return "Korean"
	case "zh":
		return "Chinese"
	default:
		return code
	}
}

func (e *Extractor) formatBitrate(bitrate string) string {
	if bitrate == "" {
		return "Unknown"
	}

	// Parse as integer (bits per second)
	bps, err := strconv.ParseFloat(strings.Split(bitrate, " ")[0], 64)
	if err != nil {
		return bitrate
	}

	// Convert to Mbps or kbps
	if bps >= 1000000 {
		return fmt.Sprintf("%.1f Mbps", bps/1000000)
	}

	if bps >= 1000 {
		return fmt.Sprintf("%.0f kbps", bps/1000)
	}

	return fmt.Sprintf("%.0f bps", bps)
}

func (e *Extractor) formatSampleRate(rate string) string {
	if rate == "" {
		return "Unknown"
	}

	return fmt.Sprintf("%s Hz", strings.Split(rate, " ")[0])
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(strings.Split(s, " ")[0], 64)
	return f
}

func parseFileSize(s string) int64 {
	if s == "" {
		return 0
	}
	size, _ := strconv.ParseInt(strings.Split(s, " ")[0], 10, 64)
	return size
}

func parseDuration(s string) int {
	if s == "" {
		return 0
	}
	// Duration in milliseconds
	ms, _ := strconv.ParseFloat(strings.Split(s, " ")[0], 64)
	return int(ms / 1000) // Convert to seconds
}
