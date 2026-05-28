package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerPort         string
	DBPath             string
	TMDBAPIKey         string
	TVDBAPIKey         string
	RadarrURL          string
	SonarrURL          string
	RadarrAPIKey       string
	SonarrAPIKey       string
	RadarrPathMapping  string   // path mapping for Radarr (format: "from:to")
	SonarrPathMapping  string   // path mapping for Sonarr (format: "from:to")
	ScanInterval       int      // hours between scans (0 = disabled)
	MediaLibraryPaths  []string // directories to scan for media files
	MoviesLibraryPaths []string // directories to scan for movies (optional)
	SeriesLibraryPaths []string // directories to scan for series (optional)
	SkipFolders        []string // folder names to skip during scanning
	MediainfoPath      string   // path to mediainfo binary
	ScanTimeout        int      // timeout in seconds per file

	// Authentication settings
	AuthMode          string // none, simple, oidc
	AuthAdminUsername string // admin username (for simple auth)
	AuthAdminPassword string // admin password (for simple auth)
	AuthSessionSecret string // secret for signing session tokens
	AuthSessionMaxAge int    // session duration in hours (default: 168 = 7 days)

	// OIDC settings
	OIDCIssuerURL      string   // OIDC issuer URL (e.g., https://accounts.google.com)
	OIDCClientID       string   // OIDC client ID
	OIDCClientSecret   string   // OIDC client secret
	OIDCRedirectURL    string   // Redirect URL after OIDC authentication
	OIDCScopes         []string // OIDC scopes (defaults to openid, profile, email)
	OIDCRolesClaim     string   // Roles claim name to check for admin role (e.g., "groups" or "roles")
	OIDCAdminRoleValue string   // Claim value that grants admin role (e.g., "admins" or "indexarr-admin")
	OIDCUsernameClaim  string   // Claim to use for username (default: "preferred_username" or "email")
	OIDCAutoCreateUser bool     // Automatically create users on first OIDC login
}

func Load() *Config {
	// Generate random session secret if not provided
	sessionSecret := getEnv("AUTH_SESSION_SECRET", "")
	if sessionSecret == "" {
		sessionSecret = generateRandomSecret()
	}

	// Default OIDC scopes
	defaultScopes := []string{"openid", "profile", "email"}
	oidcScopes := getEnvList("OIDC_SCOPES", defaultScopes)
	if len(oidcScopes) == 0 {
		oidcScopes = defaultScopes
	}

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
		MediainfoPath:      getEnv("MEDIAINFO_PATH", "mediainfo"),
		ScanTimeout:        getEnvInt("SCAN_TIMEOUT", 30),

		// Authentication
		AuthMode:          getEnv("AUTH_MODE", "none"),
		AuthAdminUsername: getEnv("AUTH_ADMIN_USERNAME", ""),
		AuthAdminPassword: getEnv("AUTH_ADMIN_PASSWORD", ""),
		AuthSessionSecret: sessionSecret,
		AuthSessionMaxAge: getEnvInt("AUTH_SESSION_MAX_AGE", 168), // 7 days

		// OIDC
		OIDCIssuerURL:      getEnv("OIDC_ISSUER_URL", ""),
		OIDCClientID:       getEnv("OIDC_CLIENT_ID", ""),
		OIDCClientSecret:   getEnv("OIDC_CLIENT_SECRET", ""),
		OIDCRedirectURL:    getEnv("OIDC_REDIRECT_URL", ""),
		OIDCScopes:         oidcScopes,
		OIDCRolesClaim:     getEnv("OIDC_ROLES_CLAIM", ""),
		OIDCAdminRoleValue: getEnv("OIDC_ADMIN_ROLE_VALUE", ""),
		OIDCUsernameClaim:  getEnv("OIDC_USERNAME_CLAIM", "preferred_username"),
		OIDCAutoCreateUser: getEnvBool("OIDC_AUTO_CREATE_USER", true),
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

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		lower := strings.ToLower(value)
		if lower == "true" || lower == "1" || lower == "yes" {
			return true
		}
		if lower == "false" || lower == "0" || lower == "no" {
			return false
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

// generateRandomSecret generates a random 32-byte hex string for session signing
func generateRandomSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a default (not ideal, but better than crashing)
		return "indexarr-default-secret-change-me"
	}
	return hex.EncodeToString(bytes)
}

// HasAuthEnabled returns true if authentication is enabled
func (c *Config) HasAuthEnabled() bool {
	return c.AuthMode == "simple" || c.AuthMode == "oidc"
}

// IsSimpleAuth returns true if simple authentication mode is enabled
func (c *Config) IsSimpleAuth() bool {
	return c.AuthMode == "simple"
}

// IsOIDCAuth returns true if OIDC authentication mode is enabled
func (c *Config) IsOIDCAuth() bool {
	return c.AuthMode == "oidc"
}

// HasOIDCConfig returns true if OIDC is properly configured
func (c *Config) HasOIDCConfig() bool {
	return c.OIDCIssuerURL != "" && c.OIDCClientID != "" && c.OIDCClientSecret != ""
}
