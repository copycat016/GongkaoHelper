package models

type MusicPlaylist struct {
	BaseModel
	Name        string `json:"name" gorm:"size:160;not null"`
	Description string `json:"description" gorm:"size:1000"`
	OwnerRole   string `json:"owner_role" gorm:"size:40;not null;default:root"`
	Enabled     bool   `json:"enabled" gorm:"not null;default:true"`
	TrackCount  int64  `json:"track_count" gorm:"->;-:migration"`
}

type MusicTrack struct {
	BaseModel
	Title          string `json:"title" gorm:"size:240;not null"`
	Artist         string `json:"artist" gorm:"size:160"`
	Album          string `json:"album" gorm:"size:160"`
	Year           string `json:"year" gorm:"size:40"`
	ReleaseDate    string `json:"release_date" gorm:"size:40"`
	Genre          string `json:"genre" gorm:"size:80"`
	CoverURL       string `json:"cover_url" gorm:"size:1000"`
	Lyrics         string `json:"lyrics" gorm:"type:text"`
	LyricsType     string `json:"lyrics_type" gorm:"size:40"`
	LyricsSource   string `json:"lyrics_source" gorm:"size:80"`
	ExternalSource string `json:"external_source" gorm:"size:80"`
	ExternalID     string `json:"external_id" gorm:"size:180"`
	DurationSec    int    `json:"duration_sec" gorm:"not null;default:0"`
	OriginalName   string `json:"original_name" gorm:"size:300;not null"`
	FilePath       string `json:"-" gorm:"size:500;not null"`
	PublicURL      string `json:"public_url" gorm:"size:500;not null"`
	MimeType       string `json:"mime_type" gorm:"size:120"`
	SizeBytes      int64  `json:"size_bytes" gorm:"not null;default:0"`
	OwnerRole      string `json:"owner_role" gorm:"size:40;not null;default:root"`
	Enabled        bool   `json:"enabled" gorm:"not null;default:true"`
}

type MusicPlaylistTrack struct {
	BaseModel
	PlaylistID uint `json:"playlist_id" gorm:"not null;index;uniqueIndex:idx_music_playlist_track_user_playlist_track"`
	TrackID    uint `json:"track_id" gorm:"not null;index;uniqueIndex:idx_music_playlist_track_user_playlist_track"`
	SortOrder  int  `json:"sort_order" gorm:"not null;default:0"`
}
