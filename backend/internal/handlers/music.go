package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gkweb/backend/internal/models"
	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type MusicHandler struct {
	service *services.MusicService
}

type applyMusicMetadataPayload struct {
	Source      string `json:"source"`
	ExternalID  string `json:"external_id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	ReleaseDate string `json:"release_date"`
	Year        string `json:"year"`
	Genre       string `json:"genre"`
	CoverURL    string `json:"cover_url"`
}

type updateMusicPlaylistPayload struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
}

type sortPlaylistTracksPayload struct {
	TrackIDs []uint `json:"track_ids"`
	Tracks   []struct {
		ID        uint `json:"id"`
		TrackID   uint `json:"track_id"`
		SortOrder int  `json:"sort_order"`
	} `json:"tracks"`
}

func NewMusicHandler(service *services.MusicService) *MusicHandler {
	return &MusicHandler{service: service}
}

func (h *MusicHandler) ListPlaylists(c *gin.Context) {
	playlists, err := h.service.ListPlaylists(userIDFromRequest(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50051, "list playlists failed")
		return
	}
	response.Success(c, playlists)
}

func (h *MusicHandler) CreatePlaylist(c *gin.Context) {
	var playlist models.MusicPlaylist
	if err := c.ShouldBindJSON(&playlist); err != nil {
		response.Error(c, http.StatusBadRequest, 40051, "invalid playlist payload")
		return
	}
	playlist.Name = strings.TrimSpace(playlist.Name)
	if playlist.Name == "" {
		response.Error(c, http.StatusBadRequest, 40051, "playlist name is required")
		return
	}
	playlist.UserID = userIDFromRequest(c)
	if playlist.OwnerRole == "" {
		playlist.OwnerRole = "root"
	}
	if err := h.service.CreatePlaylist(&playlist); err != nil {
		response.Error(c, http.StatusInternalServerError, 50052, "create playlist failed")
		return
	}
	response.Success(c, playlist)
}

func (h *MusicHandler) UpdatePlaylist(c *gin.Context) {
	playlistID, err := uintParam(c, "playlist_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40054, "invalid playlist id")
		return
	}
	var payload updateMusicPlaylistPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40051, "invalid playlist payload")
		return
	}
	playlist, err := h.service.UpdatePlaylist(userIDFromRequest(c), playlistID, services.PlaylistUpdate{
		Name:        payload.Name,
		Description: payload.Description,
		Enabled:     payload.Enabled,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, 40057, "playlist not found")
			return
		}
		if strings.Contains(err.Error(), "playlist name") {
			response.Error(c, http.StatusBadRequest, 40051, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, 50057, "update playlist failed")
		return
	}
	response.Success(c, playlist)
}

func (h *MusicHandler) ListTracks(c *gin.Context) {
	tracks, err := h.service.ListTracks(userIDFromRequest(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50053, "list tracks failed")
		return
	}
	response.Success(c, tracks)
}

func (h *MusicHandler) UploadTrack(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40052, "music file is required")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedAudioExt(ext) {
		response.Error(c, http.StatusBadRequest, 40053, "unsupported audio format")
		return
	}

	storedName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	if err := os.MkdirAll(filepath.Join("uploads", "music"), 0755); err != nil {
		response.Error(c, http.StatusInternalServerError, 50054, "prepare music upload directory failed")
		return
	}
	_ = os.MkdirAll(filepath.Join("uploads", "music", "covers"), 0755)
	relativePath := filepath.Join("uploads", "music", storedName)
	if err := c.SaveUploadedFile(file, relativePath); err != nil {
		response.Error(c, http.StatusInternalServerError, 50054, "save music file failed")
		return
	}

	metadata := services.ReadAudioMetadata(relativePath, file.Filename)

	// 优先级: 手动提交 > ID3 标签(已含 filename fallback) > 原始文件名
	title := firstNonEmpty(c.PostForm("title"), metadata.Title)
	if title == "" {
		title = strings.TrimSuffix(file.Filename, ext)
	}
	artist := firstNonEmpty(c.PostForm("artist"), metadata.Artist)
	album := firstNonEmpty(c.PostForm("album"), metadata.Album)

	track := models.MusicTrack{
		BaseModel:    models.BaseModel{UserID: userIDFromRequest(c)},
		Title:        title,
		Artist:       artist,
		Album:        album,
		Year:         metadata.Year,
		ReleaseDate:  metadata.ReleaseDate,
		Genre:        metadata.Genre,
		CoverURL:     metadata.CoverURL,
		DurationSec:  metadata.DurationSec,
		OriginalName: file.Filename,
		FilePath:     relativePath,
		PublicURL:    "/uploads/music/" + storedName,
		MimeType:     file.Header.Get("Content-Type"),
		SizeBytes:    file.Size,
		OwnerRole:    "root",
		Enabled:      true,
	}

	if err := h.service.CreateTrack(&track); err != nil {
		response.Error(c, http.StatusInternalServerError, 50055, "create music track failed")
		return
	}

	if rawPlaylistID := c.PostForm("playlist_id"); rawPlaylistID != "" {
		if playlistID, err := strconv.ParseUint(rawPlaylistID, 10, 64); err == nil && playlistID > 0 {
			_ = h.service.AddTrackToPlaylist(track.UserID, uint(playlistID), track.ID)
		}
	}

	response.Success(c, track)
}

func (h *MusicHandler) LookupTrackMetadata(c *gin.Context) {
	trackID, err := uintParam(c, "track_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40055, "invalid track id")
		return
	}
	candidates, err := h.service.LookupTrackMetadata(userIDFromRequest(c), trackID)
	if err != nil {
		writeServiceError(c, err, "lookup track metadata failed")
		return
	}
	response.Success(c, candidates)
}

func (h *MusicHandler) ApplyTrackMetadata(c *gin.Context) {
	trackID, err := uintParam(c, "track_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40055, "invalid track id")
		return
	}
	var payload applyMusicMetadataPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40056, "invalid metadata payload")
		return
	}
	track, err := h.service.ApplyTrackMetadata(userIDFromRequest(c), trackID, services.MusicMetadataCandidate{
		Source:      payload.Source,
		ExternalID:  payload.ExternalID,
		Title:       payload.Title,
		Artist:      payload.Artist,
		Album:       payload.Album,
		ReleaseDate: payload.ReleaseDate,
		Year:        payload.Year,
		Genre:       payload.Genre,
		CoverURL:    payload.CoverURL,
	})
	if err != nil {
		writeServiceError(c, err, "apply track metadata failed")
		return
	}
	response.Success(c, track)
}

func (h *MusicHandler) FetchTrackLyrics(c *gin.Context) {
	trackID, err := uintParam(c, "track_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40055, "invalid track id")
		return
	}
	track, err := h.service.FetchTrackLyrics(userIDFromRequest(c), trackID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, 40058, "track not found")
			return
		}
		response.Error(c, http.StatusBadRequest, 40064, err.Error())
		return
	}
	response.Success(c, track)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (h *MusicHandler) AddTrackToPlaylist(c *gin.Context) {
	playlistID, err := uintParam(c, "playlist_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40054, "invalid playlist id")
		return
	}
	trackID, err := uintParam(c, "track_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40055, "invalid track id")
		return
	}

	if err := h.service.AddTrackToPlaylist(userIDFromRequest(c), playlistID, trackID); err != nil {
		writeServiceError(c, err, "add track to playlist failed")
		return
	}
	response.Success(c, gin.H{"added": true})
}

func (h *MusicHandler) RemoveTrackFromPlaylist(c *gin.Context) {
	playlistID, err := uintParam(c, "playlist_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40054, "invalid playlist id")
		return
	}
	trackID, err := uintParam(c, "track_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40055, "invalid track id")
		return
	}

	if err := h.service.RemoveTrackFromPlaylist(userIDFromRequest(c), playlistID, trackID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, 40059, "playlist track not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, 50059, "remove track from playlist failed")
		return
	}
	response.Success(c, gin.H{"removed": true})
}

func (h *MusicHandler) UpdatePlaylistSort(c *gin.Context) {
	playlistID, err := uintParam(c, "playlist_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40054, "invalid playlist id")
		return
	}
	var payload sortPlaylistTracksPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40060, "invalid playlist sort payload")
		return
	}
	trackIDs := payload.TrackIDs
	if len(trackIDs) == 0 && len(payload.Tracks) > 0 {
		sort.SliceStable(payload.Tracks, func(i, j int) bool {
			return payload.Tracks[i].SortOrder < payload.Tracks[j].SortOrder
		})
		for _, item := range payload.Tracks {
			trackID := item.TrackID
			if trackID == 0 {
				trackID = item.ID
			}
			if trackID > 0 {
				trackIDs = append(trackIDs, trackID)
			}
		}
	}

	if err := h.service.UpdatePlaylistSort(userIDFromRequest(c), playlistID, trackIDs); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, 40059, "playlist track not found")
			return
		}
		if strings.Contains(err.Error(), "duplicate track id") {
			response.Error(c, http.StatusBadRequest, 40060, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, 50060, "update playlist sort failed")
		return
	}
	response.Success(c, gin.H{"sorted": true})
}

func (h *MusicHandler) DeletePlaylist(c *gin.Context) {
	playlistID, err := uintParam(c, "playlist_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40054, "invalid playlist id")
		return
	}
	if err := h.service.DeletePlaylist(userIDFromRequest(c), playlistID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, 40057, "playlist not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, 50057, "delete playlist failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *MusicHandler) TrackPlaylists(c *gin.Context) {
	trackID, err := uintParam(c, "track_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40055, "invalid track id")
		return
	}
	playlists, err := h.service.TrackPlaylists(userIDFromRequest(c), trackID)
	if err != nil {
		writeServiceError(c, err, "list track playlists failed")
		return
	}
	response.Success(c, playlists)
}

func (h *MusicHandler) DeleteTrack(c *gin.Context) {
	trackID, err := uintParam(c, "track_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40055, "invalid track id")
		return
	}
	if err := h.service.DeleteTrack(userIDFromRequest(c), trackID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, 40058, "track not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, 50058, "delete track failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *MusicHandler) PlaylistTracks(c *gin.Context) {
	playlistID, err := uintParam(c, "playlist_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40054, "invalid playlist id")
		return
	}
	tracks, err := h.service.PlaylistTracks(userIDFromRequest(c), playlistID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, 40057, "playlist not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, 50056, "list playlist tracks failed")
		return
	}
	response.Success(c, tracks)
}

func allowedAudioExt(ext string) bool {
	switch ext {
	case ".mp3", ".flac", ".wav", ".ogg", ".m4a", ".aac":
		return true
	default:
		return false
	}
}
