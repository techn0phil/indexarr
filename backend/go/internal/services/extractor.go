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
	Type          string `json:"@type"`
	Format        string `json:"Format"`
	FormatProfile string `json:"Format_profile"`
	CodecID       string `json:"CodecID"`
	Duration      string `json:"Duration"`
	Width         string `json:"Width"`
	Height        string `json:"Height"`
	FrameRate     string `json:"FrameRate"`
	FrameRateMode string `json:"FrameRate_Mode"`
	BitRate       string `json:"BitRate"`
	BitRateMode   string `json:"BitRate_Mode"`
	BitDepth      string `json:"BitDepth"`
	ColorSpace    string `json:"colour_primaries"`
	HDRFormat     string `json:"HDR_Format"`
	HDRFormatComp string `json:"HDR_Format_Compatibility"`
	Channels      string `json:"Channels"`
	ChannelLayout string `json:"ChannelLayout"`
	SamplingRate  string `json:"SamplingRate"`
	Language      string `json:"Language"`
	Title         string `json:"Title"`
	FileSize      string `json:"FileSize"`
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
				Language:   track.Language,
				SampleRate: e.formatSampleRate(track.SamplingRate),
				Bitrate:    e.formatBitrate(track.BitRate),
			})

		case "Text":
			info.SubtitleTracks = append(info.SubtitleTracks, models.SubtitleTrack{
				Language: track.Language,
				Format:   track.Format,
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
		// Only mark as HDR10 if color space is BT.2020
		if strings.Contains(strings.ToLower(track.ColorSpace), "bt.2020") ||
			strings.Contains(strings.ToLower(track.ColorSpace), "2020") {
			return "HDR10"
		}
	}

	return ""
}

func (e *Extractor) parseColorSpace(track MediainfoTrack) string {
	cs := track.ColorSpace
	if cs == "" {
		return "BT.709" // Default SDR color space
	}

	if strings.Contains(strings.ToLower(cs), "2020") {
		return "BT.2020"
	}
	if strings.Contains(strings.ToLower(cs), "709") {
		return "BT.709"
	}
	return cs
}

func (e *Extractor) parseAudioCodec(track MediainfoTrack) string {
	format := track.Format
	profile := track.FormatProfile

	if format == "" {
		return "Unknown"
	}

	// Handle TrueHD with Atmos
	if strings.Contains(strings.ToLower(format), "truehd") {
		if strings.Contains(strings.ToLower(profile), "atmos") ||
			strings.Contains(strings.ToLower(track.Title), "atmos") {
			return "TrueHD Atmos"
		}
		return "TrueHD"
	}

	// Handle DTS variants
	if strings.Contains(strings.ToUpper(format), "DTS") {
		if strings.Contains(strings.ToUpper(profile), "MA") {
			return "DTS-HD MA"
		}
		if strings.Contains(strings.ToUpper(profile), "X") {
			return "DTS:X"
		}
		return "DTS"
	}

	// Handle AC-3 / E-AC-3
	if strings.Contains(strings.ToUpper(format), "E-AC-3") ||
		strings.Contains(strings.ToUpper(format), "EAC3") {
		if strings.Contains(strings.ToLower(track.Title), "atmos") {
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

func (e *Extractor) formatBitrate(bitrate string) string {
	if bitrate == "" {
		return ""
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
	return fmt.Sprintf("%.0f kbps", bps/1000)
}

func (e *Extractor) formatSampleRate(rate string) string {
	if rate == "" {
		return ""
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
