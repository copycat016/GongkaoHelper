package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

type MusicService struct {
	db *gorm.DB
}

type MusicMetadataCandidate struct {
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

type LyricsLookupResult struct {
	Source     string `json:"source"`
	ExternalID string `json:"external_id"`
	TrackName  string `json:"track_name"`
	ArtistName string `json:"artist_name"`
	AlbumName  string `json:"album_name"`
	Type       string `json:"type"`
	Lyrics     string `json:"lyrics"`
}

type PlaylistUpdate struct {
	Name        *string
	Description *string
	Enabled     *bool
}

func NewMusicService(db *gorm.DB) *MusicService {
	return &MusicService{db: db}
}

func (s *MusicService) ListPlaylists(userID uint) ([]models.MusicPlaylist, error) {
	var playlists []models.MusicPlaylist
	err := s.db.
		Model(&models.MusicPlaylist{}).
		Select("music_playlists.*, COUNT(music_playlist_tracks.id) AS track_count").
		Joins("LEFT JOIN music_playlist_tracks ON music_playlist_tracks.playlist_id = music_playlists.id AND music_playlist_tracks.user_id = music_playlists.user_id").
		Where("music_playlists.user_id = ?", userID).
		Group("music_playlists.id").
		Order("music_playlists.created_at desc").
		Find(&playlists).Error
	return playlists, err
}

func (s *MusicService) CreatePlaylist(playlist *models.MusicPlaylist) error {
	return s.db.Create(playlist).Error
}

func (s *MusicService) UpdatePlaylist(userID uint, playlistID uint, update PlaylistUpdate) (*models.MusicPlaylist, error) {
	var playlist models.MusicPlaylist
	if err := s.db.Where("user_id = ? AND id = ?", userID, playlistID).First(&playlist).Error; err != nil {
		return nil, err
	}
	if update.Name != nil {
		name := strings.TrimSpace(*update.Name)
		if name == "" {
			return nil, errors.New("playlist name is required")
		}
		playlist.Name = name
	}
	if update.Description != nil {
		playlist.Description = strings.TrimSpace(*update.Description)
	}
	if update.Enabled != nil {
		playlist.Enabled = *update.Enabled
	}
	if err := s.db.Save(&playlist).Error; err != nil {
		return nil, err
	}
	return &playlist, nil
}

func (s *MusicService) ListTracks(userID uint) ([]models.MusicTrack, error) {
	var tracks []models.MusicTrack
	err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&tracks).Error
	if err == nil {
		s.ensureTrackCoverURLs(tracks)
	}
	return tracks, err
}

func (s *MusicService) CreateTrack(track *models.MusicTrack) error {
	return s.db.Create(track).Error
}

func (s *MusicService) AddTrackToPlaylist(userID uint, playlistID uint, trackID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := ensurePlaylistExists(tx, userID, playlistID); err != nil {
			return err
		}
		if err := ensureTrackExists(tx, userID, trackID); err != nil {
			return err
		}

		var existing models.MusicPlaylistTrack
		err := tx.Where("user_id = ? AND playlist_id = ? AND track_id = ?", userID, playlistID, trackID).First(&existing).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		sortOrder, err := nextPlaylistSortOrder(tx, userID, playlistID)
		if err != nil {
			return err
		}
		link := models.MusicPlaylistTrack{
			BaseModel:  models.BaseModel{UserID: userID},
			PlaylistID: playlistID,
			TrackID:    trackID,
			SortOrder:  sortOrder,
		}
		return tx.Create(&link).Error
	})
}

func (s *MusicService) PlaylistTracks(userID uint, playlistID uint) ([]models.MusicTrack, error) {
	if err := ensurePlaylistExists(s.db, userID, playlistID); err != nil {
		return nil, err
	}
	var tracks []models.MusicTrack
	err := s.db.
		Table("music_tracks").
		Select("music_tracks.*").
		Joins("JOIN music_playlist_tracks ON music_playlist_tracks.track_id = music_tracks.id").
		Where("music_playlist_tracks.user_id = ? AND music_playlist_tracks.playlist_id = ?", userID, playlistID).
		Order("music_playlist_tracks.sort_order asc, music_playlist_tracks.created_at asc").
		Find(&tracks).Error
	if err == nil {
		s.ensureTrackCoverURLs(tracks)
	}
	return tracks, err
}

func (s *MusicService) RemoveTrackFromPlaylist(userID uint, playlistID uint, trackID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := ensurePlaylistExists(tx, userID, playlistID); err != nil {
			return err
		}
		result := tx.Where("user_id = ? AND playlist_id = ? AND track_id = ?", userID, playlistID, trackID).Delete(&models.MusicPlaylistTrack{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return normalizePlaylistSort(tx, userID, playlistID)
	})
}

func (s *MusicService) UpdatePlaylistSort(userID uint, playlistID uint, trackIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := ensurePlaylistExists(tx, userID, playlistID); err != nil {
			return err
		}
		var links []models.MusicPlaylistTrack
		if err := tx.Where("user_id = ? AND playlist_id = ?", userID, playlistID).Order("sort_order asc, created_at asc").Find(&links).Error; err != nil {
			return err
		}

		byTrackID := make(map[uint]models.MusicPlaylistTrack, len(links))
		for _, link := range links {
			byTrackID[link.TrackID] = link
		}

		seen := make(map[uint]bool, len(trackIDs))
		ordered := make([]models.MusicPlaylistTrack, 0, len(links))
		for _, trackID := range trackIDs {
			if seen[trackID] {
				return errors.New("duplicate track id in sort payload")
			}
			link, ok := byTrackID[trackID]
			if !ok {
				return gorm.ErrRecordNotFound
			}
			seen[trackID] = true
			ordered = append(ordered, link)
		}
		for _, link := range links {
			if !seen[link.TrackID] {
				ordered = append(ordered, link)
			}
		}

		for index, link := range ordered {
			if err := tx.Model(&models.MusicPlaylistTrack{}).
				Where("id = ? AND user_id = ?", link.ID, userID).
				Update("sort_order", index).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *MusicService) LookupTrackMetadata(userID uint, trackID uint) ([]MusicMetadataCandidate, error) {
	track, err := s.getTrack(userID, trackID)
	if err != nil {
		return nil, err
	}
	filenameArtist, filenameTitle := ParseFilenameMetadata(firstNonEmptyString(track.OriginalName, track.Title))
	candidates := make([]MusicMetadataCandidate, 0, 24)
	if filenameTitle != "" || filenameArtist != "" {
		candidates = append(candidates, MusicMetadataCandidate{
			Source:     "filename",
			ExternalID: fmt.Sprintf("%d", track.ID),
			Title:      firstNonEmptyString(filenameTitle, track.Title),
			Artist:     firstNonEmptyString(filenameArtist, track.Artist),
			Album:      track.Album,
			Year:       track.Year,
			Genre:      track.Genre,
			CoverURL:   track.CoverURL,
		})
	}
	if track.Title != "" || track.Artist != "" || track.Album != "" || track.CoverURL != "" {
		candidates = append(candidates, MusicMetadataCandidate{
			Source:      "local",
			ExternalID:  fmt.Sprintf("%d", track.ID),
			Title:       track.Title,
			Artist:      track.Artist,
			Album:       track.Album,
			ReleaseDate: track.ReleaseDate,
			Year:        track.Year,
			Genre:       track.Genre,
			CoverURL:    track.CoverURL,
		})
	}

	// 构建 iTunes 搜索查询：优先使用 DB 中已有的 title/artist（来自 ID3 标签），
	// 其次使用文件名解析结果，最后使用原始文件名。
	// DB 字段通常比文件名解析更可靠，因为文件名可能已被重命名或格式不标准。
	bestArtist := firstNonEmptyString(track.Artist, filenameArtist)
	bestTitle := firstNonEmptyString(track.Title, filenameTitle)

	// 构建查询: 如果 artist 和 title 都有，用 "artist title" 获得更精准的搜索结果
	var queryParts []string
	if bestArtist != "" {
		queryParts = append(queryParts, bestArtist)
	}
	if bestTitle != "" {
		queryParts = append(queryParts, bestTitle)
	}
	query := strings.TrimSpace(strings.Join(queryParts, " "))

	// 如果 title/artist 都为空，尝试用原始文件名（去掉扩展名）作为搜索词
	if query == "" {
		query = strings.TrimSpace(track.OriginalName)
		if ext := strings.LastIndex(query, "."); ext > 0 {
			query = query[:ext]
		}
	}
	if query == "" {
		if len(candidates) > 0 {
			return candidates, nil
		}
		return nil, errors.New("track has no searchable metadata")
	}
	itunesCandidates, err := searchITunesMetadata(query)
	if err != nil {
		if len(candidates) > 0 {
			return candidates, nil
		}
		return nil, err
	}
	return append(candidates, itunesCandidates...), nil
}

func (s *MusicService) ApplyTrackMetadata(userID uint, trackID uint, candidate MusicMetadataCandidate) (*models.MusicTrack, error) {
	track, err := s.getTrack(userID, trackID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(candidate.Title) != "" {
		track.Title = candidate.Title
	}
	if strings.TrimSpace(candidate.Artist) != "" {
		track.Artist = candidate.Artist
	}
	if strings.TrimSpace(candidate.Album) != "" {
		track.Album = candidate.Album
	}
	if strings.TrimSpace(candidate.Year) != "" {
		track.Year = candidate.Year
	}
	if strings.TrimSpace(candidate.ReleaseDate) != "" {
		track.ReleaseDate = candidate.ReleaseDate
	}
	if strings.TrimSpace(candidate.Genre) != "" {
		track.Genre = candidate.Genre
	}
	if strings.TrimSpace(candidate.CoverURL) != "" {
		track.CoverURL = candidate.CoverURL
	}
	track.ExternalSource = candidate.Source
	track.ExternalID = candidate.ExternalID

	return track, s.db.Save(track).Error
}

func (s *MusicService) FetchTrackLyrics(userID uint, trackID uint) (*models.MusicTrack, error) {
	track, err := s.getTrack(userID, trackID)
	if err != nil {
		return nil, err
	}
	result, err := searchNeteaseLyrics(track)
	if err != nil {
		return nil, err
	}
	track.Lyrics = result.Lyrics
	track.LyricsType = result.Type
	track.LyricsSource = result.Source
	track.ExternalID = firstNonEmptyString(track.ExternalID, result.ExternalID)
	if err := s.db.Save(track).Error; err != nil {
		return nil, err
	}
	return track, nil
}

func (s *MusicService) DeletePlaylist(userID uint, playlistID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND playlist_id = ?", userID, playlistID).Delete(&models.MusicPlaylistTrack{}).Error; err != nil {
			return err
		}
		result := tx.Where("user_id = ? AND id = ?", userID, playlistID).Delete(&models.MusicPlaylist{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (s *MusicService) DeleteTrack(userID uint, trackID uint) error {
	var track models.MusicTrack
	if err := s.db.Where("user_id = ? AND id = ?", userID, trackID).First(&track).Error; err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND track_id = ?", userID, trackID).Delete(&models.MusicPlaylistTrack{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND id = ?", userID, trackID).Delete(&models.MusicTrack{}).Error; err != nil {
			return err
		}
		if track.FilePath != "" {
			_ = os.Remove(track.FilePath)
		}
		return nil
	})
}

func ensurePlaylistExists(db *gorm.DB, userID uint, playlistID uint) error {
	var count int64
	if err := db.Model(&models.MusicPlaylist{}).Where("user_id = ? AND id = ?", userID, playlistID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func ensureTrackExists(db *gorm.DB, userID uint, trackID uint) error {
	var count int64
	if err := db.Model(&models.MusicTrack{}).Where("user_id = ? AND id = ?", userID, trackID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func nextPlaylistSortOrder(db *gorm.DB, userID uint, playlistID uint) (int, error) {
	var links []models.MusicPlaylistTrack
	if err := db.Where("user_id = ? AND playlist_id = ?", userID, playlistID).Order("sort_order desc").Limit(1).Find(&links).Error; err != nil {
		return 0, err
	}
	if len(links) == 0 {
		return 0, nil
	}
	return links[0].SortOrder + 1, nil
}

func normalizePlaylistSort(db *gorm.DB, userID uint, playlistID uint) error {
	var links []models.MusicPlaylistTrack
	if err := db.Where("user_id = ? AND playlist_id = ?", userID, playlistID).Order("sort_order asc, created_at asc").Find(&links).Error; err != nil {
		return err
	}
	sort.SliceStable(links, func(i, j int) bool {
		if links[i].SortOrder == links[j].SortOrder {
			return links[i].CreatedAt.Before(links[j].CreatedAt)
		}
		return links[i].SortOrder < links[j].SortOrder
	})
	for index, link := range links {
		if link.SortOrder == index {
			continue
		}
		if err := db.Model(&models.MusicPlaylistTrack{}).
			Where("id = ? AND user_id = ?", link.ID, userID).
			Update("sort_order", index).Error; err != nil {
			return err
		}
	}
	return nil
}

// TrackPlaylists 返回指定曲目所属的所有歌单（含歌单名称）
func (s *MusicService) TrackPlaylists(userID uint, trackID uint) ([]models.MusicPlaylist, error) {
	if err := ensureTrackExists(s.db, userID, trackID); err != nil {
		return nil, err
	}
	var playlists []models.MusicPlaylist
	err := s.db.
		Table("music_playlists").
		Select("music_playlists.*").
		Joins("JOIN music_playlist_tracks ON music_playlist_tracks.playlist_id = music_playlists.id").
		Where("music_playlist_tracks.user_id = ? AND music_playlist_tracks.track_id = ?", userID, trackID).
		Order("music_playlists.name asc").
		Find(&playlists).Error
	return playlists, err
}

func (s *MusicService) getTrack(userID uint, trackID uint) (*models.MusicTrack, error) {
	var track models.MusicTrack
	if err := s.db.Where("user_id = ? AND id = ?", userID, trackID).First(&track).Error; err != nil {
		return nil, err
	}
	if track.CoverURL == "" && track.FilePath != "" {
		if coverURL := ExtractEmbeddedCover(track.FilePath); coverURL != "" {
			track.CoverURL = coverURL
			_ = s.db.Model(&track).Update("cover_url", coverURL).Error
		}
	}
	return &track, nil
}

func (s *MusicService) ensureTrackCoverURLs(tracks []models.MusicTrack) {
	for index := range tracks {
		if tracks[index].CoverURL != "" || tracks[index].FilePath == "" {
			continue
		}
		coverURL := ExtractEmbeddedCover(tracks[index].FilePath)
		if coverURL == "" {
			continue
		}
		tracks[index].CoverURL = coverURL
		if err := s.db.Model(&tracks[index]).Update("cover_url", coverURL).Error; err != nil {
			logTrackCoverUpdateError(tracks[index].ID, err)
		}
	}
}

func logTrackCoverUpdateError(trackID uint, err error) {
	if err != nil {
		fmt.Printf("music metadata: update cover_url for track %d failed: %v\n", trackID, err)
	}
}

func searchITunesMetadata(query string) ([]MusicMetadataCandidate, error) {
	endpoint := "https://itunes.apple.com/search?media=music&limit=25&term=" + url.QueryEscape(query)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Gkweb/0.1 music metadata lookup")

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("itunes metadata lookup failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("itunes metadata lookup returned http %d", resp.StatusCode)
	}

	var payload struct {
		Results []struct {
			TrackID        int64  `json:"trackId"`
			TrackName      string `json:"trackName"`
			ArtistName     string `json:"artistName"`
			CollectionName string `json:"collectionName"`
			ReleaseDate    string `json:"releaseDate"`
			PrimaryGenre   string `json:"primaryGenreName"`
			ArtworkURL100  string `json:"artworkUrl100"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	candidates := make([]MusicMetadataCandidate, 0, len(payload.Results))
	for _, item := range payload.Results {
		cover := strings.Replace(item.ArtworkURL100, "100x100bb", "600x600bb", 1)
		candidates = append(candidates, MusicMetadataCandidate{
			Source:      "itunes",
			ExternalID:  fmt.Sprintf("%d", item.TrackID),
			Title:       item.TrackName,
			Artist:      item.ArtistName,
			Album:       item.CollectionName,
			ReleaseDate: shortDate(item.ReleaseDate),
			Year:        yearFromDate(item.ReleaseDate),
			Genre:       item.PrimaryGenre,
			CoverURL:    cover,
		})
	}
	return candidates, nil
}

func searchNeteaseLyrics(track *models.MusicTrack) (*LyricsLookupResult, error) {
	trackName := strings.TrimSpace(track.Title)
	artistName := strings.TrimSpace(track.Artist)
	filenameArtist, filenameTitle := ParseFilenameMetadata(firstNonEmptyString(track.OriginalName, track.Title))
	if trackName == "" {
		trackName = filenameTitle
	}
	if artistName == "" {
		artistName = filenameArtist
	}
	if trackName == "" {
		return nil, errors.New("track has no title for lyrics lookup")
	}

	searchQueries := lyricsSearchQueries(track, trackName, artistName)
	var candidates []neteaseSongCandidate
	var lookupErrors []string
	for _, query := range searchQueries {
		items, err := searchNeteaseSongs(query)
		if err != nil {
			lookupErrors = append(lookupErrors, err.Error())
			continue
		}
		candidates = append(candidates, items...)
	}
	if len(candidates) == 0 {
		if len(lookupErrors) > 0 {
			return nil, fmt.Errorf("网易云音乐未找到歌曲，已尝试 %d 种查询：%s", len(searchQueries), strings.Join(uniqueStrings(lookupErrors), "; "))
		}
		return nil, errors.New("网易云音乐未找到歌曲")
	}

	scoredSongs := make([]scoredNeteaseSong, 0, len(candidates))
	seen := make(map[int64]bool, len(candidates))
	for _, item := range candidates {
		if item.ID != 0 && seen[item.ID] {
			continue
		}
		if item.ID != 0 {
			seen[item.ID] = true
		}
		score := lyricsMatchScore(track, item.TrackName, item.ArtistName, item.AlbumName)
		scoredSongs = append(scoredSongs, scoredNeteaseSong{item: item, score: score})
	}
	if len(scoredSongs) == 0 {
		return nil, errors.New("网易云音乐返回的歌曲候选为空")
	}
	sort.SliceStable(scoredSongs, func(i, j int) bool {
		return scoredSongs[i].score > scoredSongs[j].score
	})

	var lyricErrors []string
	for _, scored := range scoredSongs {
		lyrics, translatedLyrics, err := fetchNeteaseLyrics(scored.item.ID)
		if err != nil {
			lyricErrors = append(lyricErrors, err.Error())
			continue
		}
		mergedLyrics := mergeLrcLyrics(lyrics, translatedLyrics)
		if strings.TrimSpace(mergedLyrics) == "" {
			lyricErrors = append(lyricErrors, fmt.Sprintf("%s 无歌词内容", scored.item.TrackName))
			continue
		}
		return &LyricsLookupResult{
			Source:     "netease",
			ExternalID: fmt.Sprintf("%d", scored.item.ID),
			TrackName:  scored.item.TrackName,
			ArtistName: scored.item.ArtistName,
			AlbumName:  scored.item.AlbumName,
			Type:       "lrc",
			Lyrics:     mergedLyrics,
		}, nil
	}
	return nil, fmt.Errorf("网易云音乐找到候选歌曲，但没有可用歌词：%s", strings.Join(uniqueStrings(lyricErrors), "; "))
}

type neteaseSongCandidate struct {
	ID         int64
	TrackName  string
	ArtistName string
	AlbumName  string
}

type scoredNeteaseSong struct {
	item  neteaseSongCandidate
	score int
}

func searchNeteaseSongs(query string) ([]neteaseSongCandidate, error) {
	values := url.Values{}
	values.Set("s", strings.TrimSpace(query))
	values.Set("type", "1")
	values.Set("limit", "12")
	values.Set("offset", "0")
	values.Set("total", "true")
	values.Set("csrf_token", "")

	var payload struct {
		Result struct {
			Songs []struct {
				ID      int64  `json:"id"`
				Name    string `json:"name"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
				Album struct {
					Name string `json:"name"`
				} `json:"album"`
				AR []struct {
					Name string `json:"name"`
				} `json:"ar"`
				AL struct {
					Name string `json:"name"`
				} `json:"al"`
			} `json:"songs"`
		} `json:"result"`
	}
	if err := requestNeteaseJSON("https://music.163.com/api/cloudsearch/pc?"+values.Encode(), &payload); err != nil {
		return nil, err
	}
	candidates := make([]neteaseSongCandidate, 0, len(payload.Result.Songs))
	for _, song := range payload.Result.Songs {
		var artists []string
		for _, artist := range song.Artists {
			if strings.TrimSpace(artist.Name) != "" {
				artists = append(artists, strings.TrimSpace(artist.Name))
			}
		}
		for _, artist := range song.AR {
			if strings.TrimSpace(artist.Name) != "" {
				artists = append(artists, strings.TrimSpace(artist.Name))
			}
		}
		artists = uniqueStrings(artists)
		albumName := firstNonEmptyString(song.Album.Name, song.AL.Name)
		candidates = append(candidates, neteaseSongCandidate{
			ID:         song.ID,
			TrackName:  song.Name,
			ArtistName: strings.Join(artists, " / "),
			AlbumName:  albumName,
		})
	}
	return candidates, nil
}

func fetchNeteaseLyrics(songID int64) (string, string, error) {
	values := url.Values{}
	values.Set("id", fmt.Sprintf("%d", songID))
	values.Set("lv", "-1")
	values.Set("kv", "-1")
	values.Set("tv", "-1")
	var payload struct {
		Lrc struct {
			Lyric string `json:"lyric"`
		} `json:"lrc"`
		TLrc struct {
			Lyric string `json:"lyric"`
		} `json:"tlyric"`
	}
	if err := requestNeteaseJSON("https://music.163.com/api/song/lyric?"+values.Encode(), &payload); err != nil {
		return "", "", err
	}
	return payload.Lrc.Lyric, payload.TLrc.Lyric, nil
}

func requestNeteaseJSON(endpoint string, target any) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://music.163.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 Gkweb/0.1 music lyrics lookup")

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("netease lyrics lookup failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("netease lyrics lookup returned http %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func mergeLrcLyrics(lyrics string, translatedLyrics string) string {
	lyrics = strings.TrimSpace(lyrics)
	translatedLyrics = strings.TrimSpace(translatedLyrics)
	if lyrics == "" {
		return translatedLyrics
	}
	if translatedLyrics == "" {
		return lyrics
	}
	return lyrics + "\n" + translatedLyrics
}

func lyricsSearchQueries(track *models.MusicTrack, trackName string, artistName string) []string {
	var queries []string
	if artistName != "" && trackName != "" {
		queries = append(queries, artistName+" "+trackName)
	}
	if trackName != "" {
		queries = append(queries, trackName)
	}
	original := strings.TrimSpace(track.OriginalName)
	if original != "" {
		if ext := strings.LastIndex(original, "."); ext > 0 {
			original = original[:ext]
		}
		queries = append(queries, original)
	}
	return uniqueStrings(queries)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	var result []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func lyricsMatchScore(track *models.MusicTrack, trackName string, artistName string, albumName string) int {
	sourceTitle := normalizeMusicLookupText(track.Title)
	sourceArtist := normalizeMusicLookupText(track.Artist)
	sourceAlbum := normalizeMusicLookupText(track.Album)
	score := 0
	if sourceTitle != "" && sourceTitle == normalizeMusicLookupText(trackName) {
		score += 60
	} else if sourceTitle != "" && strings.Contains(normalizeMusicLookupText(trackName), sourceTitle) {
		score += 35
	}
	if sourceArtist != "" && sourceArtist == normalizeMusicLookupText(artistName) {
		score += 30
	}
	if sourceAlbum != "" && sourceAlbum == normalizeMusicLookupText(albumName) {
		score += 10
	}
	return score
}

func normalizeMusicLookupText(value string) string {
	replacer := strings.NewReplacer(
		"【", " ", "】", " ",
		"[", " ", "]", " ",
		"(", " ", ")", " ",
		"（", " ", "）", " ",
		"「", " ", "」", " ",
		"『", " ", "』", " ",
		"-", " ", "_", " ",
		"·", " ", "/", " ",
		"　", " ",
	)
	return strings.ToLower(strings.Join(strings.Fields(replacer.Replace(value)), " "))
}

func shortDate(value string) string {
	if len(value) >= 10 {
		return value[:10]
	}
	return value
}

func yearFromDate(value string) string {
	if len(value) >= 4 {
		return value[:4]
	}
	return ""
}
