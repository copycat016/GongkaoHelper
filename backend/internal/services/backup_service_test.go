package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gkweb/backend/internal/database"
	"gkweb/backend/internal/models"
)

func TestBackupExportCoversCurrentBusinessTablesAndFileManifest(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	trackPath := filepath.Join(t.TempDir(), "track.mp3")
	if err := writeTestFile(trackPath); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	records := []any{
		&models.User{Username: "admin", PasswordHash: "secret", Role: "owner", Enabled: true},
		&models.StageGoal{BaseModel: models.BaseModel{UserID: 1}, Title: "阶段目标"},
		&models.StageItem{BaseModel: models.BaseModel{UserID: 1}, StageGoalID: 1, Title: "阶段子项"},
		&models.WeeklyTask{BaseModel: models.BaseModel{UserID: 1}, Title: "周任务", WeekStart: mustParseDay(t, "2026-06-15")},
		&models.DailyTask{BaseModel: models.BaseModel{UserID: 1}, Title: "日任务"},
		&models.EssaySection{BaseModel: models.BaseModel{UserID: 1}, DocumentID: 1, SectionType: "material", Content: "材料"},
		&models.EssaySectionRelation{BaseModel: models.BaseModel{UserID: 1}, DocumentID: 1, QuestionID: 1, SectionID: 1, RelationType: "question_material"},
		&models.ThemeConfig{BaseModel: models.BaseModel{UserID: 1}, Palette: "aozora"},
		&models.MusicTrack{BaseModel: models.BaseModel{UserID: 1}, Title: "曲目", OriginalName: "track.mp3", FilePath: trackPath, PublicURL: "/uploads/music/track.mp3"},
	}
	for _, record := range records {
		if err := db.Create(record).Error; err != nil {
			t.Fatalf("create record %T: %v", record, err)
		}
	}

	export, err := NewBackupService(db).Export(false)
	if err != nil {
		t.Fatalf("export backup: %v", err)
	}

	for _, table := range []string{
		"users",
		"stage_goals",
		"stage_items",
		"weekly_tasks",
		"daily_tasks",
		"essay_sections",
		"essay_section_relations",
		"theme_configs",
	} {
		if _, ok := export.Tables[table]; !ok {
			t.Fatalf("missing backup table %q", table)
		}
		if export.Summary[table] == 0 {
			t.Fatalf("summary for %q was not populated", table)
		}
	}
	if len(export.Files) != 1 {
		t.Fatalf("expected one file manifest entry, got %d", len(export.Files))
	}
	if !export.Files[0].Exists || export.Files[0].Path != trackPath || export.Files[0].Table != "music_tracks" {
		t.Fatalf("unexpected file manifest: %#v", export.Files[0])
	}
}

func writeTestFile(path string) error {
	return os.WriteFile(path, []byte("test"), 0600)
}

func mustParseDay(t *testing.T, value string) time.Time {
	t.Helper()
	day, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		t.Fatalf("parse day: %v", err)
	}
	return day
}
