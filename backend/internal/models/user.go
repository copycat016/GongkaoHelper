package models

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"size:80;not null;uniqueIndex"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	DisplayName  string    `json:"display_name" gorm:"size:120"`
	Role         string    `json:"role" gorm:"size:40;not null;default:owner"`
	Enabled      bool      `json:"enabled" gorm:"not null;default:true;index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
