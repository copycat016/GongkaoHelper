package services

import (
	"time"

	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

type BackupExport struct {
	ExportedAt time.Time      `json:"exported_at"`
	Version    string         `json:"version"`
	Tables     map[string]any `json:"tables"`
	Summary    map[string]int `json:"summary"`
	Notes      []string       `json:"notes"`
}

type BackupService struct {
	db *gorm.DB
}

func NewBackupService(db *gorm.DB) *BackupService {
	return &BackupService{db: db}
}

func (s *BackupService) Export(includeSecrets bool) (*BackupExport, error) {
	providers := []models.LLMProvider{}
	modelsList := []models.LLMModel{}
	prompts := []models.PromptTemplate{}
	mistakes := []models.Mistake{}
	pomodoros := []models.PomodoroSession{}
	logs := []models.StudyLog{}
	plans := []models.StudyPlan{}
	playlists := []models.MusicPlaylist{}
	tracks := []models.MusicTrack{}
	playlistTracks := []models.MusicPlaylistTrack{}
	ocrTasks := []models.OCRTask{}
	essayDocuments := []models.EssayDocument{}
	essayChunks := []models.EssayChunk{}
	essayQuestions := []models.EssayQuestion{}
	essayQuestionChunks := []models.EssayQuestionChunk{}
	essayReviews := []models.EssayReview{}

	queries := []struct {
		name string
		dest any
	}{
		{"llm_providers", &providers},
		{"llm_models", &modelsList},
		{"prompt_templates", &prompts},
		{"mistakes", &mistakes},
		{"pomodoro_sessions", &pomodoros},
		{"study_logs", &logs},
		{"study_plans", &plans},
		{"music_playlists", &playlists},
		{"music_tracks", &tracks},
		{"music_playlist_tracks", &playlistTracks},
		{"ocr_tasks", &ocrTasks},
		{"essay_documents", &essayDocuments},
		{"essay_chunks", &essayChunks},
		{"essay_questions", &essayQuestions},
		{"essay_question_chunks", &essayQuestionChunks},
		{"essay_reviews", &essayReviews},
	}

	for _, query := range queries {
		if err := s.db.Order("id asc").Find(query.dest).Error; err != nil {
			return nil, err
		}
	}

	if !includeSecrets {
		for index := range providers {
			providers[index].APIKey = ""
		}
	}

	tables := map[string]any{
		"llm_providers":         providers,
		"llm_models":            modelsList,
		"prompt_templates":      prompts,
		"mistakes":              mistakes,
		"pomodoro_sessions":     pomodoros,
		"study_logs":            logs,
		"study_plans":           plans,
		"music_playlists":       playlists,
		"music_tracks":          tracks,
		"music_playlist_tracks": playlistTracks,
		"ocr_tasks":             ocrTasks,
		"essay_documents":       essayDocuments,
		"essay_chunks":          essayChunks,
		"essay_questions":       essayQuestions,
		"essay_question_chunks": essayQuestionChunks,
		"essay_reviews":         essayReviews,
	}

	summary := map[string]int{
		"llm_providers":         len(providers),
		"llm_models":            len(modelsList),
		"prompt_templates":      len(prompts),
		"mistakes":              len(mistakes),
		"pomodoro_sessions":     len(pomodoros),
		"study_logs":            len(logs),
		"study_plans":           len(plans),
		"music_playlists":       len(playlists),
		"music_tracks":          len(tracks),
		"music_playlist_tracks": len(playlistTracks),
		"ocr_tasks":             len(ocrTasks),
		"essay_documents":       len(essayDocuments),
		"essay_chunks":          len(essayChunks),
		"essay_questions":       len(essayQuestions),
		"essay_question_chunks": len(essayQuestionChunks),
		"essay_reviews":         len(essayReviews),
	}

	notes := []string{
		"该文件是业务数据 JSON 备份，不包含上传的音乐、图片等二进制文件。",
		"如未勾选包含敏感配置，LLM Provider API Key 会被置空。",
		"百度 OCR 本地配置文件 backend/data/ocr_config.json 不属于数据库表，未包含在本导出中。",
	}

	return &BackupExport{
		ExportedAt: time.Now(),
		Version:    "gkweb-backup-v1",
		Tables:     tables,
		Summary:    summary,
		Notes:      notes,
	}, nil
}
