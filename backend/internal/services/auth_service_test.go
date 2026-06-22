package services

import (
	"errors"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gkweb/backend/internal/config"
	"gkweb/backend/internal/models"
)

func TestBootstrapSingleOwnerRequiresPasswordInRelease(t *testing.T) {
	db := newAuthTestDB(t)
	service := NewAuthService(db, config.Config{
		GinMode:   "release",
		AuthMode:  "single",
		JWTSecret: "test-secret",
	})

	err := service.BootstrapSingleOwner()
	if err == nil || !strings.Contains(err.Error(), "AUTH_BOOTSTRAP_PASSWORD") {
		t.Fatalf("expected missing bootstrap password error, got %v", err)
	}
}

func TestBootstrapSingleOwnerCreatesReleaseOwnerWithConfiguredPassword(t *testing.T) {
	db := newAuthTestDB(t)
	service := NewAuthService(db, config.Config{
		GinMode:                  "release",
		AuthMode:                 "single",
		JWTSecret:                "test-secret",
		JWTExpireHours:           168,
		AuthBootstrapUsername:    "admin",
		AuthBootstrapPassword:    "correct horse battery staple",
		AuthBootstrapDisplayName: "Owner",
	})

	if err := service.BootstrapSingleOwner(); err != nil {
		t.Fatalf("bootstrap owner: %v", err)
	}

	result, err := service.Login("admin", "correct horse battery staple")
	if err != nil {
		t.Fatalf("login with configured password: %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
}

func TestBootstrapSingleOwnerRotatesDevelopmentPasswordInRelease(t *testing.T) {
	db := newAuthTestDB(t)
	devService := NewAuthService(db, config.Config{
		GinMode:        "debug",
		AuthMode:       "single",
		JWTSecret:      "test-secret",
		JWTExpireHours: 168,
	})
	if err := devService.BootstrapSingleOwner(); err != nil {
		t.Fatalf("bootstrap development owner: %v", err)
	}

	releaseService := NewAuthService(db, config.Config{
		GinMode:               "release",
		AuthMode:              "single",
		JWTSecret:             "test-secret",
		JWTExpireHours:        168,
		AuthBootstrapPassword: "correct horse battery staple",
	})
	if err := releaseService.BootstrapSingleOwner(); err != nil {
		t.Fatalf("rotate development password: %v", err)
	}

	if _, err := releaseService.Login(DefaultOwnerUsername, DevOwnerPassword); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected old development password to fail, got %v", err)
	}
	if _, err := releaseService.Login(DefaultOwnerUsername, "correct horse battery staple"); err != nil {
		t.Fatalf("expected rotated password to work: %v", err)
	}
}

func newAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}
