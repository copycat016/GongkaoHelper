package database

import (
	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.LLMProvider{},
		&models.LLMModel{},
		&models.PromptTemplate{},
		&models.Mistake{},
		&models.PomodoroSession{},
		&models.StudyLog{},
		&models.StudyPlan{},
		&models.StageGoal{},
		&models.WeeklyTask{},
		&models.DailyTask{},
		&models.MusicPlaylist{},
		&models.MusicTrack{},
		&models.MusicPlaylistTrack{},
		&models.OCRTask{},
		&models.EssayDocument{},
		&models.EssayChunk{},
		&models.EssaySection{},
		&models.EssayQuestion{},
		&models.EssayQuestionChunk{},
		&models.EssaySectionRelation{},
		&models.EssayReview{},
	)
}
