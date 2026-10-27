package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config represents all application configuration loaded and validated from environment variables.
type Config struct {
	Port     int    `env:"PORT" envDefault:"8080"`
	Env      string `env:"ENV" envDefault:"development"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	// Database Configuration
	DBDSN                string `env:"DB_DSN,required"`
	DBMaxOpenConns       int    `env:"DB_MAX_OPEN_CONNS" envDefault:"50"`
	DBMaxIdleConns       int    `env:"DB_MAX_IDLE_CONNS" envDefault:"25"`
	DBConnMaxLifetimeMin int    `env:"DB_CONN_MAX_LIFETIME_MIN" envDefault:"15"`

	// Redis Configuration
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	// Security & JWT
	JWTSecret           string `env:"JWT_SECRET,required"`
	JWTAccessTTLMinutes int    `env:"JWT_ACCESS_TTL_MINUTES" envDefault:"15"`
	JWTRefreshTTLDays   int    `env:"JWT_REFRESH_TTL_DAYS" envDefault:"7"`

	// Media Storage
	MediaStorageDir string `env:"MEDIA_STORAGE_DIR" envDefault:"./data/media"`
	MaxUploadSizeMB int64  `env:"MAX_UPLOAD_SIZE_MB" envDefault:"10"`

	// Cloudflare CDN
	CloudflareZoneID   string `env:"CLOUDFLARE_ZONE_ID"`
	CloudflareAPIToken string `env:"CLOUDFLARE_API_TOKEN"`

	// CORS Allowed Origins
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000,http://localhost:4321,http://localhost:5173"`
}

// Load reads optional .env files and parses all required environment variables into Config.
func Load() (*Config, error) {
	// Attempt to load .env file from common candidate paths
	candidates := []string{
		os.Getenv("ENV_FILE"),
		".env",
		"apps/api/.env",
		"../.env",
		"../../.env",
	}
	for _, path := range candidates {
		if path != "" {
			loadDotEnv(path)
		}
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Validate secret length for production security
	if len(cfg.JWTSecret) < 16 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 16 characters for security")
	}

	return cfg, nil
}

// loadDotEnv reads key=value lines from a .env file into the environment if not already set.
func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Strip optional quotes
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
	}
}