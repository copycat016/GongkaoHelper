package config

import (
	"fmt"
	"os"
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
}

func Load() Config {
	return Config{
		ServerPort: getEnv("SERVER_PORT", "21080"),
		GinMode:    getEnv("GIN_MODE", "debug"),

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
