package services

import (
	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

type MusicService struct {
	db *gorm.DB
}

func NewMusicService(db *gorm.DB) *MusicService {
	return &MusicService{db: db}
}

func (s *MusicService) ListPlaylists(userID uint) ([]models.MusicPlaylist, error) {
	var playlists []models.MusicPlaylist
	err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&playlists).Error
	return playlists, err
}

func (s *MusicService) CreatePlaylist(playlist *models.MusicPlaylist) error {
	return s.db.Create(playlist).Error
}

func (s *MusicService) ListTracks(userID uint) ([]models.MusicTrack, error) {
	var tracks []models.MusicTrack
	err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&tracks).Error
	return tracks, err
}

func (s *MusicService) CreateTrack(track *models.MusicTrack) error {
	return s.db.Create(track).Error
}

func (s *MusicService) AddTrackToPlaylist(userID uint, playlistID uint, trackID uint) error {
	var count int64
	if err := s.db.Model(&models.MusicPlaylist{}).Where("user_id = ? AND id = ?", userID, playlistID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}

	link := models.MusicPlaylistTrack{
		BaseModel:  models.BaseModel{UserID: userID},
		PlaylistID: playlistID,
		TrackID:    trackID,
	}
	return s.db.FirstOrCreate(&link, "user_id = ? AND playlist_id = ? AND track_id = ?", userID, playlistID, trackID).Error
}

func (s *MusicService) PlaylistTracks(userID uint, playlistID uint) ([]models.MusicTrack, error) {
	var tracks []models.MusicTrack
	err := s.db.
		Table("music_tracks").
		Select("music_tracks.*").
		Joins("JOIN music_playlist_tracks ON music_playlist_tracks.track_id = music_tracks.id").
		Where("music_playlist_tracks.user_id = ? AND music_playlist_tracks.playlist_id = ?", userID, playlistID).
		Order("music_playlist_tracks.sort_order asc, music_playlist_tracks.created_at asc").
		Find(&tracks).Error
	return tracks, err
}
