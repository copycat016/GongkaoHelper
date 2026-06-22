package services

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	tokenauth "gkweb/backend/internal/auth"
	"gkweb/backend/internal/config"
	"gkweb/backend/internal/models"
)

const (
	DefaultOwnerID          uint   = 1
	DefaultOwnerUsername    string = "admin"
	DefaultOwnerDisplayName string = "Owner"
	DevOwnerPassword        string = "123456"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type AuthService struct {
	db  *gorm.DB
	cfg config.Config
}

type UserInfo struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	AuthMode    string `json:"auth_mode"`
}

type LoginResult struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int64    `json:"expires_in"`
	User        UserInfo `json:"user"`
}

func NewAuthService(db *gorm.DB, cfg config.Config) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

func (s *AuthService) BootstrapSingleOwner() error {
	if s.cfg.AuthMode != "single" {
		return nil
	}

	var count int64
	if err := s.db.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return s.rotateDevelopmentOwnerPassword()
	}

	username := strings.TrimSpace(s.cfg.AuthBootstrapUsername)
	if username == "" {
		username = DefaultOwnerUsername
	}
	displayName := strings.TrimSpace(s.cfg.AuthBootstrapDisplayName)
	if displayName == "" {
		displayName = DefaultOwnerDisplayName
	}
	password, err := s.bootstrapPassword(username)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{
		ID:           DefaultOwnerID,
		Username:     username,
		PasswordHash: string(hash),
		DisplayName:  displayName,
		Role:         "owner",
		Enabled:      true,
	}
	return s.db.Create(&user).Error
}

func (s *AuthService) bootstrapPassword(username string) (string, error) {
	password := strings.TrimSpace(s.cfg.AuthBootstrapPassword)
	if password == "" {
		if isReleaseMode(s.cfg.GinMode) {
			return "", errors.New("AUTH_BOOTSTRAP_PASSWORD is required for first startup in release mode")
		}
		return DevOwnerPassword, nil
	}
	if isReleaseMode(s.cfg.GinMode) && unsafeBootstrapPassword(username, password) {
		return "", errors.New("AUTH_BOOTSTRAP_PASSWORD must be at least 12 characters and not a common default")
	}
	return password, nil
}

func (s *AuthService) rotateDevelopmentOwnerPassword() error {
	var user models.User
	err := s.db.Where("id = ? AND username = ?", DefaultOwnerID, DefaultOwnerUsername).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(DevOwnerPassword)) != nil {
		return nil
	}

	replacement := strings.TrimSpace(s.cfg.AuthBootstrapPassword)
	if replacement == "" {
		if isReleaseMode(s.cfg.GinMode) {
			return errors.New("default admin still uses the development password; set AUTH_BOOTSTRAP_PASSWORD to rotate it before running in release mode")
		}
		return nil
	}
	if isReleaseMode(s.cfg.GinMode) && unsafeBootstrapPassword(user.Username, replacement) {
		return errors.New("AUTH_BOOTSTRAP_PASSWORD must be at least 12 characters and not a common default")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(replacement), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.Model(&user).Update("password_hash", string(hash)).Error
}

func isReleaseMode(mode string) bool {
	return strings.EqualFold(strings.TrimSpace(mode), "release")
}

func unsafeBootstrapPassword(username string, password string) bool {
	trimmed := strings.TrimSpace(password)
	normalized := strings.ToLower(trimmed)
	if len([]rune(trimmed)) < 12 {
		return true
	}
	switch normalized {
	case "123456", "12345678", "123456789", "password", "admin", "gkweb", "changeme":
		return true
	}
	return normalized == strings.ToLower(strings.TrimSpace(username))
}

func (s *AuthService) Login(username string, password string) (*LoginResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	var user models.User
	if err := s.db.Where("username = ? AND enabled = ?", username, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	ttl := time.Duration(s.cfg.JWTExpireHours) * time.Hour
	token, err := tokenauth.GenerateAccessToken(user, s.cfg.JWTSecret, ttl)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(ttl.Seconds()),
		User:        s.UserInfo(user),
	}, nil
}

func (s *AuthService) FindEnabledUser(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND enabled = ?", id, true).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) UserInfo(user models.User) UserInfo {
	return UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		AuthMode:    s.cfg.AuthMode,
	}
}
