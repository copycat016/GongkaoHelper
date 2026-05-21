package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gkweb/backend/internal/config"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	dialector, err := dialector(cfg)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(cfg.DBDriver, "sqlite") {
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	} else {
		sqlDB.SetMaxOpenConns(20)
		sqlDB.SetMaxIdleConns(10)
	}
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func dialector(cfg config.Config) (gorm.Dialector, error) {
	switch strings.ToLower(cfg.DBDriver) {
	case "", "sqlite":
		if err := ensureSQLiteDir(cfg.SQLitePath); err != nil {
			return nil, err
		}
		return sqlite.Open(cfg.SQLiteDSN()), nil
	case "postgres", "postgresql":
		return postgres.Open(cfg.PostgresDSN()), nil
	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER %q", cfg.DBDriver)
	}
}

func ensureSQLiteDir(path string) error {
	if path == "" || strings.HasPrefix(path, "file:") {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return sqlDB.PingContext(pingCtx)
}
