package services

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ParsedFilename contains extracted metadata from a filename
type ParsedFilename struct {
	Title      string
	Year       int
	Season     int // 0 if not a series
	Episode    int // 0 if not a series
	IsSeries   bool
	Resolution string // e.g., "4K", "1080p", "720p"
	Source     string // e.g., "BluRay", "WEB-DL", "HDTV"
}

var (
	// Series patterns: S01E05, 1x05, Season 1 Episode 5
	seriesPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)[.\s_-]?S(\d{1,2})[E.](\d{1,3})[.\s_-]`),                  // S01E05, S01.05
		regexp.MustCompile(`(?i)[.\s_-]?(\d{1,2})x(\d{1,3})[.\s_-]`),                      // 1x05
		regexp.MustCompile(`(?i)Season[.\s_-]?(\d{1,2})[.\s_-]?Episode[.\s_-]?(\d{1,3})`), // Season 1 Episode 5
	}

	// Year patterns: (2024), 2024, .2024.
	yearPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\((\d{4})\)`),       // (2024)
		regexp.MustCompile(`[._-](\d{4})[._-]`), // .2024., -2024-, _2024_
		regexp.MustCompile(`\b(\d{4})\b`),       // 2024
	}

	// Resolution patterns
	resolutionPatterns = map[string]*regexp.Regexp{
		"4K":    regexp.MustCompile(`(?i)(2160p|4K|UHD)`),
		"1080p": regexp.MustCompile(`(?i)1080p`),
		"720p":  regexp.MustCompile(`(?i)720p`),
		"480p":  regexp.MustCompile(`(?i)480p`),
	}

	// Source patterns
	sourcePatterns = map[string]*regexp.Regexp{
		"BluRay": regexp.MustCompile(`(?i)(BluRay|Blu-Ray|BDRip|BRRip)`),
		"WEB-DL": regexp.MustCompile(`(?i)(WEB-DL|WEBDL|WEB\.DL)`),
		"WEBRip": regexp.MustCompile(`(?i)(WEBRip|WEB-Rip)`),
		"HDTV":   regexp.MustCompile(`(?i)HDTV`),
		"DVDRip": regexp.MustCompile(`(?i)(DVDRip|DVD-Rip)`),
		"Remux":  regexp.MustCompile(`(?i)Remux`),
	}

	// Cleanup patterns - things to remove from title
	cleanupPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\.(mkv|mp4|avi|mov|m4v|webm|flv|wmv)$`),                              // file extensions
		regexp.MustCompile(`(?i)(2160p|1080p|720p|480p)`),                                            // resolutions
		regexp.MustCompile(`(?i)(BluRay|Blu-Ray|BDRip|BRRip|WEB-DL|WEBDL|WEBRip|HDTV|DVDRip|Remux)`), // sources
		regexp.MustCompile(`(?i)(x264|x265|H\.?264|H\.?265|HEVC|AVC|AAC|AC3|DTS|TrueHD|Atmos)`),      // codecs
		regexp.MustCompile(`(?i)(HDR10\+?|Dolby\.?Vision|DV|DoVi)`),                                  // HDR formats
		regexp.MustCompile(`(?i)(PROPER|REPACK|EXTENDED|UNRATED|DIRECTORS\.?CUT)`),                   // release tags
		regexp.MustCompile(`(?i)(\[.*?\]|\(.*?\))`),                                                  // bracketed content
		regexp.MustCompile(`-[A-Za-z0-9]+$`),                                                         // release group
	}
)

// ParseFilename extracts metadata from a media filename
func ParseFilename(filename string) *ParsedFilename {
	result := &ParsedFilename{}

	// Get base name without path
	basename := filepath.Base(filename)

	// Try to detect if it's a series
	for _, pattern := range seriesPatterns {
		matches := pattern.FindStringSubmatch(basename)
		if len(matches) >= 3 {
			result.IsSeries = true
			result.Season, _ = strconv.Atoi(matches[1])
			result.Episode, _ = strconv.Atoi(matches[2])
			// Remove the series identifier from filename for title extraction
			basename = pattern.ReplaceAllString(basename, ".")
			break
		}
	}

	// Extract year
	for _, yearPattern := range yearPatterns {
		matches := yearPattern.FindStringSubmatch(filename)
		if len(matches) >= 2 {
			year, _ := strconv.Atoi(matches[1])
			if year >= 1900 && year <= 2100 {
				result.Year = year
				break
			}
		}
	}

	// Extract resolution
	for res, pattern := range resolutionPatterns {
		if pattern.MatchString(basename) {
			result.Resolution = res
			break
		}
	}

	// Extract source
	for src, pattern := range sourcePatterns {
		if pattern.MatchString(basename) {
			result.Source = src
			break
		}
	}

	// Extract title
	result.Title = extractTitle(filename, result.Year)

	return result
}

// extractTitle cleans up the filename to extract just the title
func extractTitle(filename string, year int) string {
	title := filename

	// Remove year and everything after if we found a year
	if year > 0 {
		yearStr := strconv.Itoa(year)
		idx := strings.Index(title, yearStr)
		if idx > 0 {
			title = title[:idx-1]
		}
	}

	// Apply cleanup patterns
	for _, pattern := range cleanupPatterns {
		title = pattern.ReplaceAllString(title, "")
	}

	// Replace dots and underscores with spaces
	title = strings.ReplaceAll(title, ".", " ")
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.ReplaceAll(title, "-", " ")

	// Clean up multiple spaces
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")

	// Trim and clean
	title = strings.TrimSpace(title)

	// Remove folder path
	title = filepath.Base(title)

	return title
}

// IsVideoFile checks if a file extension is a supported video format
func IsVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	videoExtensions := map[string]bool{
		".mkv":  true,
		".mp4":  true,
		".avi":  true,
		".mov":  true,
		".m4v":  true,
		".webm": true,
		".flv":  true,
		".wmv":  true,
		".ts":   true,
		".m2ts": true,
	}
	return videoExtensions[ext]
}

// GetContainer returns the container format from filename
func GetContainer(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) > 0 {
		return ext[1:] // Remove the leading dot
	}
	return ""
}
