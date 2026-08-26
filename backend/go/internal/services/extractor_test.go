package services

import "testing"

func TestExtractorParseHelpers(t *testing.T) {
	e := NewExtractor("mediainfo", 1)

	t.Run("video codec normalization", func(t *testing.T) {
		tests := []struct {
			name string
			in   string
			want string
		}{
			{name: "avc", in: "AVC", want: "H.264"},
			{name: "hevc", in: "HEVC", want: "H.265"},
			{name: "av1", in: "AV1", want: "AV1"},
			{name: "passthrough", in: "ProRes", want: "ProRes"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := e.parseVideoCodec(MediainfoTrack{Format: tt.in})
				if got != tt.want {
					t.Fatalf("parseVideoCodec() = %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("audio codec normalization", func(t *testing.T) {
		tests := []struct {
			name  string
			track MediainfoTrack
			want  string
		}{
			{name: "truehd atmos", track: MediainfoTrack{Format: "MLP FBA", FormatCommercial: "Dolby TrueHD with Dolby Atmos"}, want: "TrueHD Atmos"},
			{name: "dtsx", track: MediainfoTrack{Format: "DTS", FormatCommercial: "DTS:X"}, want: "DTS:X"},
			{name: "eac3 atmos", track: MediainfoTrack{Format: "E-AC-3", FormatAdditionalFeatures: "JOC"}, want: "E-AC-3 Atmos"},
			{name: "aac", track: MediainfoTrack{Format: "AAC LC"}, want: "AAC"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := e.parseAudioCodec(tt.track)
				if got != tt.want {
					t.Fatalf("parseAudioCodec() = %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("resolution and channels", func(t *testing.T) {
		if got := e.parseResolution(MediainfoTrack{Width: "1920 pixels", Height: "1080 pixels"}); got != "1920x1080" {
			t.Fatalf("parseResolution() = %q, want %q", got, "1920x1080")
		}
		if got := e.parseChannels(MediainfoTrack{Channels: "6"}); got != "5.1" {
			t.Fatalf("parseChannels() = %q, want %q", got, "5.1")
		}
	})

	t.Run("hdr and colorspace", func(t *testing.T) {
		track := MediainfoTrack{HDRFormat: "Dolby Vision", HDRFormatComp: "HDR10", BitDepth: "10", ColorPrimaries: "BT.2020"}
		if got := e.parseHDR(track); got != "Dolby Vision / HDR10" {
			t.Fatalf("parseHDR() = %q, want %q", got, "Dolby Vision / HDR10")
		}
		if got := e.parseColorSpace(MediainfoTrack{ColorPrimaries: "BT.709"}); got != "BT.709" {
			t.Fatalf("parseColorSpace() = %q, want %q", got, "BT.709")
		}
	})

	t.Run("formatters", func(t *testing.T) {
		if got := e.formatBitrate("2500000"); got != "2.5 Mbps" {
			t.Fatalf("formatBitrate() = %q, want %q", got, "2.5 Mbps")
		}
		if got := e.formatSampleRate("48000"); got != "48000 Hz" {
			t.Fatalf("formatSampleRate() = %q, want %q", got, "48000 Hz")
		}
		if got := e.formatLanguage("EN"); got != "en" {
			t.Fatalf("formatLanguage() = %q, want %q", got, "en")
		}
	})
}

func TestParseMediaInfo_MapsTracks(t *testing.T) {
	e := NewExtractor("mediainfo", 1)

	mi := &MediainfoOutput{}
	mi.Media.Track = []MediainfoTrack{
		{Type: "General", FileSize: "123456789", Duration: "7200"},
		{
			Type:           "Video",
			Format:         "HEVC",
			Width:          "3840",
			Height:         "2160",
			FrameRate:      "23.976",
			BitRate:        "8000000",
			HDRFormat:      "SMPTE ST 2086",
			ColorPrimaries: "BT.2020",
			Extra:          MediaInfoTrackExtra{Source: "Blu-ray"},
		},
		{
			Type:                     "Audio",
			Format:                   "E-AC-3",
			FormatAdditionalFeatures: "JOC",
			Channels:                 "6",
			Language:                 "fr",
			SamplingRate:             "48000",
			BitRate:                  "768000",
			Default:                  "Yes",
		},
		{
			Type:     "Text",
			Language: "en",
			Format:   "SRT",
			Forced:   "No",
			Default:  "No",
		},
	}

	info, fileSize, duration, err := e.parseMediaInfo(mi)
	if err != nil {
		t.Fatalf("parseMediaInfo() returned error: %v", err)
	}
	if fileSize != 123456789 {
		t.Fatalf("fileSize = %d, want %d", fileSize, 123456789)
	}
	if duration != 7200 {
		t.Fatalf("duration = %d, want %d", duration, 7200)
	}
	if len(info.VideoTracks) != 1 || len(info.AudioTracks) != 1 || len(info.SubtitleTracks) != 1 {
		t.Fatalf("unexpected track counts: video=%d audio=%d subtitles=%d", len(info.VideoTracks), len(info.AudioTracks), len(info.SubtitleTracks))
	}
	if info.VideoTracks[0].Codec != "H.265" || info.VideoTracks[0].HDR != "HDR10" {
		t.Fatalf("unexpected video track mapping: %+v", info.VideoTracks[0])
	}
	if info.AudioTracks[0].Codec != "E-AC-3 Atmos" || info.AudioTracks[0].Channels != "5.1" {
		t.Fatalf("unexpected audio track mapping: %+v", info.AudioTracks[0])
	}
}

func TestParsePrimitiveHelpers(t *testing.T) {
	if got := parseFloat("23.976 fps"); got != 23.976 {
		t.Fatalf("parseFloat() = %v, want 23.976", got)
	}
	if got := parseFileSize("987654 bytes"); got != 987654 {
		t.Fatalf("parseFileSize() = %d, want 987654", got)
	}
	if got := parseDuration("5432.9"); got != 5432 {
		t.Fatalf("parseDuration() = %d, want 5432", got)
	}
}
