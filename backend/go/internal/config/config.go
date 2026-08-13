package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// GlobalLogger is the package-level logger instance accessible by all services
var GlobalLogger *zerolog.Logger

type Config struct {
	ServerPort         string
	DBPath             string
	TMDBAPIKey         string
	TVDBAPIKey         string
	RadarrURL          string
	SonarrURL          string
	RadarrAPIKey       string
	SonarrAPIKey       string
	RadarrPathMapping  string         // path mapping for Radarr (format: "from:to")
	SonarrPathMapping  string         // path mapping for Sonarr (format: "from:to")
	ScanInterval       int            // hours between scans (0 = disabled)
	MediaLibraryPaths  []string       // directories to scan for media files
	MoviesLibraryPaths []string       // directories to scan for movies (optional)
	SeriesLibraryPaths []string       // directories to scan for series (optional)
	SkipFolders        []string       // folder names to skip during scanning
	IgnoreFilePattern  *regexp.Regexp // regex pattern to ignore files
	MediainfoPath      string         // path to mediainfo binary
	ScanTimeout        int            // timeout in seconds per file
	DetectionLanguage  string         // language code for media detection (e.g., "en", "fr")
	MetadataLanguage   string         // language code for metadata fetching (e.g., "en", "fr")
}

func Load() *Config {
	return &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		DBPath:             getEnv("DB_PATH", "./indexarr.db"),
		TMDBAPIKey:         getEnv("TMDB_API_KEY", ""),
		TVDBAPIKey:         getEnv("TVDB_API_KEY", ""),
		RadarrURL:          getEnv("RADARR_URL", ""),
		SonarrURL:          getEnv("SONARR_URL", ""),
		RadarrAPIKey:       getEnv("RADARR_API_KEY", ""),
		SonarrAPIKey:       getEnv("SONARR_API_KEY", ""),
		RadarrPathMapping:  getEnv("RADARR_PATH_MAPPING", ""),
		SonarrPathMapping:  getEnv("SONARR_PATH_MAPPING", ""),
		ScanInterval:       getEnvInt("SCAN_INTERVAL", 24),
		MediaLibraryPaths:  getEnvList("MEDIA_LIBRARY_PATHS", []string{}),
		MoviesLibraryPaths: getEnvList("MOVIES_LIBRARY_PATHS", []string{}),
		SeriesLibraryPaths: getEnvList("SERIES_LIBRARY_PATHS", []string{}),
		SkipFolders:        getEnvList("SKIP_FOLDERS", []string{}),
		IgnoreFilePattern:  getEnvRegex("IGNORE_FILE_PATTERN", nil),
		MediainfoPath:      getEnv("MEDIAINFO_PATH", "mediainfo"),
		ScanTimeout:        getEnvInt("SCAN_TIMEOUT", 30),
		DetectionLanguage:  getEnv("DETECTION_LANGUAGE", "en"),
		MetadataLanguage:   getEnv("METADATA_LANGUAGE", "en"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvList(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		paths := strings.Split(value, ",")
		var result []string
		for _, p := range paths {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}

func getEnvRegex(key string, defaultValue *regexp.Regexp) *regexp.Regexp {
	if value := os.Getenv(key); value != "" {
		if regex, err := regexp.Compile(value); err == nil {
			return regex
		}
	}
	return defaultValue
}

// HasRadarrConfig returns true if Radarr API is configured
func (c *Config) HasRadarrConfig() bool {
	return c.RadarrURL != "" && c.RadarrAPIKey != ""
}

// HasSonarrConfig returns true if Sonarr API is configured
func (c *Config) HasSonarrConfig() bool {
	return c.SonarrURL != "" && c.SonarrAPIKey != ""
}

// UseFilesystemScan returns true if filesystem scanning should be used
// (when no Radarr/Sonarr config is present but media paths are configured)
func (c *Config) UseFilesystemScan() bool {
	return !c.HasRadarrConfig() && len(c.MediaLibraryPaths) > 0
}

// GetScanMode returns the current scan mode as a string (legacy, for backward compatibility)
func (c *Config) GetScanMode() string {
	if c.HasRadarrConfig() {
		return "radarr"
	}
	if len(c.MediaLibraryPaths) > 0 {
		return "filesystem"
	}
	return "disabled"
}

// GetMoviesImportMode returns the import mode for movies
func (c *Config) GetMoviesImportMode() string {
	if c.HasRadarrConfig() {
		return "radarr"
	}
	if len(c.MoviesLibraryPaths) > 0 {
		return "filesystem"
	}
	if len(c.MediaLibraryPaths) > 0 {
		return "filesystem"
	}
	return "disabled"
}

// GetSeriesImportMode returns the import mode for series
func (c *Config) GetSeriesImportMode() string {
	if c.HasSonarrConfig() {
		return "sonarr"
	}
	if len(c.SeriesLibraryPaths) > 0 {
		return "filesystem"
	}
	if len(c.MediaLibraryPaths) > 0 {
		return "filesystem"
	}
	return "disabled"
}

// GetMovieLibraryPaths returns the paths to scan for movies
func (c *Config) GetMovieLibraryPaths() []string {
	if len(c.MoviesLibraryPaths) > 0 {
		return c.MoviesLibraryPaths
	}
	return c.MediaLibraryPaths
}

// GetSeriesLibraryPaths returns the paths to scan for series
func (c *Config) GetSeriesLibraryPaths() []string {
	if len(c.SeriesLibraryPaths) > 0 {
		return c.SeriesLibraryPaths
	}
	return c.MediaLibraryPaths
}

// ParsePathMapping parses a path mapping string (format: "from:to") and returns from, to
func ParsePathMapping(mapping string) (string, string) {
	if mapping == "" {
		return "", ""
	}
	parts := strings.SplitN(mapping, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

// InitLogger initializes a zerolog logger with level from LOG_LEVEL environment variable
// Valid levels: TRACE, DEBUG, INFO, WARN, ERROR (case-insensitive, default: INFO)
// Sets the global GlobalLogger variable for use by services
func InitLogger() *zerolog.Logger {
	logLevel := strings.ToUpper(getEnv("LOG_LEVEL", "INFO"))

	// Parse and set zerolog level
	var zeroLevel zerolog.Level
	switch logLevel {
	case "TRACE":
		zeroLevel = zerolog.TraceLevel
	case "DEBUG":
		zeroLevel = zerolog.DebugLevel
	case "INFO":
		zeroLevel = zerolog.InfoLevel
	case "WARN":
		zeroLevel = zerolog.WarnLevel
	case "ERROR":
		zeroLevel = zerolog.ErrorLevel
	default:
		// Invalid level, warn and use INFO
		log.Warn().Str("invalid_level", logLevel).Msg("LOG_LEVEL not recognized, using INFO")
		zeroLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(zeroLevel)
	zerolog.DurationFieldFormat = zerolog.DurationFormatString
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	// Configure zerolog with pretty console output
	logger := log.With().Caller().Logger()
	logger = logger.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
		FormatLevel: func(i interface{}) string {
			// Colorize log level for better visibility
			levelStr := strings.ToUpper(fmt.Sprintf("%s", i))
			switch levelStr {
			case "TRACE":
				return fmt.Sprintf("\033[34m%s\033[0m", levelStr) // Blue
			case "DEBUG":
				return fmt.Sprintf("\033[36m%s\033[0m", levelStr) // Cyan
			case "INFO":
				return fmt.Sprintf("\033[32m%s\033[0m", levelStr) // Green
			case "WARN":
				return fmt.Sprintf("\033[33m%s\033[0m", levelStr) // Yellow
			case "ERROR":
				return fmt.Sprintf("\033[31m%s\033[0m", levelStr) // Red
			case "FATAL":
				return fmt.Sprintf("\033[35m%s\033[0m", levelStr) // Magenta
			case "PANIC":
				return fmt.Sprintf("\033[35m%s\033[0m", levelStr) // Magenta
			default:
				return levelStr
			}
		},
		TimeFormat: time.RFC3339,
	})

	// Set the global logger for service access
	GlobalLogger = &logger

	return &logger
}
