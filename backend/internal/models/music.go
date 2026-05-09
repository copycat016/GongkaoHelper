package models

type MusicPlaylist struct {
	BaseModel
	Name        string `json:"name" gorm:"size:160;not null"`
	Description string `json:"description" gorm:"size:1000"`
	OwnerRole   string `json:"owner_role" gorm:"size:40;not null;default:root"`
	Enabled     bool   `json:"enabled" gorm:"not null;default:true"`
}

type MusicTrack struct {
	BaseModel
	Title        string `json:"title" gorm:"size:240;not null"`
	Artist       string `json:"artist" gorm:"size:160"`
	Album        string `json:"album" gorm:"size:160"`
	Year         string `json:"year" gorm:"size:40"`
	Genre        string `json:"genre" gorm:"size:80"`
	DurationSec  int    `json:"duration_sec" gorm:"not null;default:0"`
	OriginalName string `json:"original_name" gorm:"size:300;not null"`
	FilePath     string `json:"file_path" gorm:"size:500;not null"`
	PublicURL    string `json:"public_url" gorm:"size:500;not null"`
	MimeType     string `json:"mime_type" gorm:"size:120"`
	SizeBytes    int64  `json:"size_bytes" gorm:"not null;default:0"`
	OwnerRole    string `json:"owner_role" gorm:"size:40;not null;default:root"`
	Enabled      bool   `json:"enabled" gorm:"not null;default:true"`
}

type MusicPlaylistTrack struct {
	BaseModel
	PlaylistID uint `json:"playlist_id" gorm:"not null;index"`
	TrackID    uint `json:"track_id" gorm:"not null;index"`
	SortOrder  int  `json:"sort_order" gorm:"not null;default:0"`
}
