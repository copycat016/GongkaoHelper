package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/models"
	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type MusicHandler struct {
	service *services.MusicService
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
	relativePath := filepath.Join("uploads", "music", storedName)
	if err := c.SaveUploadedFile(file, relativePath); err != nil {
		response.Error(c, http.StatusInternalServerError, 50054, "save music file failed")
		return
	}

	metadata := services.ReadAudioMetadata(relativePath)
	title := c.PostForm("title")
	if title == "" {
		title = metadata.Title
	}
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
		Genre:        metadata.Genre,
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

func (h *MusicHandler) PlaylistTracks(c *gin.Context) {
	playlistID, err := uintParam(c, "playlist_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40054, "invalid playlist id")
		return
	}
	tracks, err := h.service.PlaylistTracks(userIDFromRequest(c), playlistID)
	if err != nil {
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
