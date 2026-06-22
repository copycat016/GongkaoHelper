package services

import (
	"os"
	"time"

	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

type BackupExport struct {
	ExportedAt time.Time      `json:"exported_at"`
	Version    string         `json:"version"`
	Tables     map[string]any `json:"tables"`
	Summary    map[string]int `json:"summary"`
	Files      []BackupFile   `json:"files"`
	Notes      []string       `json:"notes"`
}

type BackupFile struct {
	Table        string `json:"table"`
	RecordID     uint   `json:"record_id"`
	Path         string `json:"path"`
	PublicURL    string `json:"public_url,omitempty"`
	OriginalName string `json:"original_name,omitempty"`
	Exists       bool   `json:"exists"`
	SizeBytes    int64  `json:"size_bytes"`
}

type BackupService struct {
	db *gorm.DB
}

func NewBackupService(db *gorm.DB) *BackupService {
	return &BackupService{db: db}
}

func (s *BackupService) Export(includeSecrets bool) (*BackupExport, error) {
	users := []models.User{}
	providers := []models.LLMProvider{}
	modelsList := []models.LLMModel{}
	prompts := []models.PromptTemplate{}
	mistakes := []models.Mistake{}
	pomodoros := []models.PomodoroSession{}
	logs := []models.StudyLog{}
	plans := []models.StudyPlan{}
	stageGoals := []models.StageGoal{}
	stageItems := []models.StageItem{}
	weeklyTasks := []models.WeeklyTask{}
	dailyTasks := []models.DailyTask{}
	playlists := []models.MusicPlaylist{}
	tracks := []models.MusicTrack{}
	playlistTracks := []models.MusicPlaylistTrack{}
	ocrTasks := []models.OCRTask{}
	essayDocuments := []models.EssayDocument{}
	essayChunks := []models.EssayChunk{}
	essaySections := []models.EssaySection{}
	essayQuestions := []models.EssayQuestion{}
	essayQuestionChunks := []models.EssayQuestionChunk{}
	essaySectionRelations := []models.EssaySectionRelation{}
	essayReviews := []models.EssayReview{}
	themeConfigs := []models.ThemeConfig{}

	queries := []struct {
		name string
		dest any
	}{
		{"users", &users},
		{"llm_providers", &providers},
		{"llm_models", &modelsList},
		{"prompt_templates", &prompts},
		{"mistakes", &mistakes},
		{"pomodoro_sessions", &pomodoros},
		{"study_logs", &logs},
		{"study_plans", &plans},
		{"stage_goals", &stageGoals},
		{"stage_items", &stageItems},
		{"weekly_tasks", &weeklyTasks},
		{"daily_tasks", &dailyTasks},
		{"music_playlists", &playlists},
		{"music_tracks", &tracks},
		{"music_playlist_tracks", &playlistTracks},
		{"ocr_tasks", &ocrTasks},
		{"essay_documents", &essayDocuments},
		{"essay_chunks", &essayChunks},
		{"essay_sections", &essaySections},
		{"essay_questions", &essayQuestions},
		{"essay_question_chunks", &essayQuestionChunks},
		{"essay_section_relations", &essaySectionRelations},
		{"essay_reviews", &essayReviews},
		{"theme_configs", &themeConfigs},
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
		"users":                   users,
		"llm_providers":           providers,
		"llm_models":              modelsList,
		"prompt_templates":        prompts,
		"mistakes":                mistakes,
		"pomodoro_sessions":       pomodoros,
		"study_logs":              logs,
		"study_plans":             plans,
		"stage_goals":             stageGoals,
		"stage_items":             stageItems,
		"weekly_tasks":            weeklyTasks,
		"daily_tasks":             dailyTasks,
		"music_playlists":         playlists,
		"music_tracks":            tracks,
		"music_playlist_tracks":   playlistTracks,
		"ocr_tasks":               ocrTasks,
		"essay_documents":         essayDocuments,
		"essay_chunks":            essayChunks,
		"essay_sections":          essaySections,
		"essay_questions":         essayQuestions,
		"essay_question_chunks":   essayQuestionChunks,
		"essay_section_relations": essaySectionRelations,
		"essay_reviews":           essayReviews,
		"theme_configs":           themeConfigs,
	}

	summary := map[string]int{
		"users":                   len(users),
		"llm_providers":           len(providers),
		"llm_models":              len(modelsList),
		"prompt_templates":        len(prompts),
		"mistakes":                len(mistakes),
		"pomodoro_sessions":       len(pomodoros),
		"study_logs":              len(logs),
		"study_plans":             len(plans),
		"stage_goals":             len(stageGoals),
		"stage_items":             len(stageItems),
		"weekly_tasks":            len(weeklyTasks),
		"daily_tasks":             len(dailyTasks),
		"music_playlists":         len(playlists),
		"music_tracks":            len(tracks),
		"music_playlist_tracks":   len(playlistTracks),
		"ocr_tasks":               len(ocrTasks),
		"essay_documents":         len(essayDocuments),
		"essay_chunks":            len(essayChunks),
		"essay_sections":          len(essaySections),
		"essay_questions":         len(essayQuestions),
		"essay_question_chunks":   len(essayQuestionChunks),
		"essay_section_relations": len(essaySectionRelations),
		"essay_reviews":           len(essayReviews),
		"theme_configs":           len(themeConfigs),
	}

	files := collectBackupFiles(tracks, essayDocuments)

	notes := []string{
		"该文件是业务数据 JSON 备份，不包含上传的音乐、PDF、图片等二进制文件；files 字段仅记录数据库引用的文件清单。",
		"如未勾选包含敏感配置，LLM Provider API Key 会被置空。",
		"用户密码哈希不会通过 JSON 导出；恢复用户密码需走初始化或重置流程。",
		"百度 OCR 本地配置文件 backend/data/ocr_config.json 不属于数据库表，未包含在本导出中。",
	}

	return &BackupExport{
		ExportedAt: time.Now(),
		Version:    "gkweb-backup-v1",
		Tables:     tables,
		Summary:    summary,
		Files:      files,
		Notes:      notes,
	}, nil
}

func collectBackupFiles(tracks []models.MusicTrack, documents []models.EssayDocument) []BackupFile {
	files := make([]BackupFile, 0, len(tracks)+len(documents))
	for _, track := range tracks {
		if track.FilePath == "" {
			continue
		}
		files = append(files, buildBackupFile("music_tracks", track.ID, track.FilePath, track.PublicURL, track.OriginalName))
	}
	for _, document := range documents {
		if document.FilePath == "" {
			continue
		}
		files = append(files, buildBackupFile("essay_documents", document.ID, document.FilePath, "", document.OriginalName))
	}
	return files
}

func buildBackupFile(table string, recordID uint, path string, publicURL string, originalName string) BackupFile {
	file := BackupFile{
		Table:        table,
		RecordID:     recordID,
		Path:         path,
		PublicURL:    publicURL,
		OriginalName: originalName,
	}
	if info, err := os.Stat(path); err == nil {
		file.Exists = true
		file.SizeBytes = info.Size()
	}
	return file
}
