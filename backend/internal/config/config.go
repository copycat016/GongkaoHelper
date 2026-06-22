package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	ServerPort string
	GinMode    string

	DBDriver   string
	SQLitePath string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBTimeZone string

	BaiduOCREnabled       bool
	BaiduOCRAPIKey        string
	BaiduOCRSecretKey     string
	BaiduOCRMonthlyLimit  int
	BaiduOCRTimeoutSecond int

	AuthMode       string
	JWTSecret      string
	JWTExpireHours int

	AuthBootstrapUsername    string
	AuthBootstrapPassword    string
	AuthBootstrapDisplayName string

	CORSAllowedOrigins []string
}

func Load() Config {
	ginMode := getEnv("GIN_MODE", "debug")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" && ginMode != "release" {
		jwtSecret = randomSecret()
		log.Print("WARNING: JWT_SECRET is not set; generated a temporary development secret. Tokens will be invalid after restart.")
	}

	return Config{
		ServerPort: getEnv("SERVER_PORT", "21080"),
		GinMode:    ginMode,

		DBDriver:   getEnv("DB_DRIVER", "sqlite"),
		SQLitePath: getEnv("SQLITE_PATH", "./data/gkweb.db"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "21432"),
		DBUser:     getEnv("DB_USER", "gkweb"),
		DBPassword: getEnv("DB_PASSWORD", "gkweb_password"),
		DBName:     getEnv("DB_NAME", "gkweb"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		DBTimeZone: getEnv("DB_TIMEZONE", "Asia/Shanghai"),

		BaiduOCREnabled:       getEnv("BAIDU_OCR_ENABLED", "false") == "true",
		BaiduOCRAPIKey:        getEnv("BAIDU_OCR_API_KEY", ""),
		BaiduOCRSecretKey:     getEnv("BAIDU_OCR_SECRET_KEY", ""),
		BaiduOCRMonthlyLimit:  getEnvInt("BAIDU_OCR_MONTHLY_LIMIT", 200),
		BaiduOCRTimeoutSecond: getEnvInt("BAIDU_OCR_TIMEOUT_SECONDS", 30),

		AuthMode:       normalizeAuthMode(getEnv("AUTH_MODE", "single")),
		JWTSecret:      jwtSecret,
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24*7),

		AuthBootstrapUsername:    getEnv("AUTH_BOOTSTRAP_USERNAME", "admin"),
		AuthBootstrapPassword:    os.Getenv("AUTH_BOOTSTRAP_PASSWORD"),
		AuthBootstrapDisplayName: getEnv("AUTH_BOOTSTRAP_DISPLAY_NAME", "Owner"),

		CORSAllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", "http://localhost:21073,http://127.0.0.1:21073,http://localhost:5173,http://127.0.0.1:5173,http://localhost:4173,http://127.0.0.1:4173"),
	}
}

func (c Config) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
		c.DBTimeZone,
	)
}

func (c Config) SQLiteDSN() string {
	return c.SQLitePath
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
		return fallback
	}
	return parsed
}

func getEnvList(key string, fallback string) []string {
	value := getEnv(key, fallback)
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" || item == "*" {
			continue
		}
		items = append(items, item)
	}
	return items
}

func normalizeAuthMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "multi":
		return "multi"
	default:
		return "single"
	}
}

func randomSecret() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		log.Printf("WARNING: failed to generate JWT development secret: %v", err)
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
