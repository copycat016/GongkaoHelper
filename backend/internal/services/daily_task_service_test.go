package services

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

func TestUpdateDailyTaskPatchPreservesOmittedFieldsAndClearsTouchedFields(t *testing.T) {
	service := newTestDailyTaskService(t)
	userID := uint(1)
	date := "2026-06-15"
	deadline := "2026-06-20"
	stage, err := service.CreateStageGoal(userID, StageGoalInput{Title: "阶段目标"})
	if err != nil {
		t.Fatalf("create stage goal: %v", err)
	}
	weekly, err := service.CreateWeeklyTask(userID, WeeklyTaskInput{Title: "周任务", WeekStart: date})
	if err != nil {
		t.Fatalf("create weekly task: %v", err)
	}
	task, err := service.CreateDailyTask(userID, DailyTaskInput{
		StageGoalID:      &stage.ID,
		WeeklyTaskID:     &weekly.ID,
		Title:            "原任务",
		Note:             "保留备注",
		Date:             &date,
		Deadline:         &deadline,
		EstimatedMinutes: 45,
	}, false)
	if err != nil {
		t.Fatalf("create daily task: %v", err)
	}

	updated, err := service.Update(userID, task.ID, DailyTaskInput{
		Title:   "新标题",
		Touched: map[string]bool{"title": true},
	})
	if err != nil {
		t.Fatalf("patch title: %v", err)
	}
	if updated.Note != "保留备注" || updated.Date == nil || updated.Deadline == nil {
		t.Fatalf("omitted fields were not preserved: %#v", updated)
	}
	if updated.StageGoalID == nil || *updated.StageGoalID != stage.ID || updated.WeeklyTaskID == nil || *updated.WeeklyTaskID != weekly.ID {
		t.Fatalf("omitted associations were not preserved: %#v", updated)
	}
	if updated.EstimatedMinutes != 45 {
		t.Fatalf("estimated minutes changed unexpectedly: %d", updated.EstimatedMinutes)
	}

	cleared, err := service.Update(userID, task.ID, DailyTaskInput{
		Note:             "",
		Date:             nil,
		Deadline:         nil,
		StageGoalID:      nil,
		WeeklyTaskID:     nil,
		EstimatedMinutes: 0,
		Touched: map[string]bool{
			"note":              true,
			"date":              true,
			"deadline":          true,
			"stage_goal_id":     true,
			"weekly_task_id":    true,
			"estimated_minutes": true,
		},
	})
	if err != nil {
		t.Fatalf("clear touched fields: %v", err)
	}
	if cleared.Note != "" || cleared.Date != nil || cleared.Deadline != nil || cleared.StageGoalID != nil || cleared.WeeklyTaskID != nil || cleared.EstimatedMinutes != 0 {
		t.Fatalf("touched fields were not cleared: %#v", cleared)
	}
}

func TestUpdateWeeklyTaskPatchPreservesAndClearsNullableFields(t *testing.T) {
	service := newTestDailyTaskService(t)
	userID := uint(1)
	weekStart := "2026-06-15"
	weekEnd := "2026-06-21"
	deadline := "2026-06-20"
	stage, err := service.CreateStageGoal(userID, StageGoalInput{Title: "阶段目标"})
	if err != nil {
		t.Fatalf("create stage goal: %v", err)
	}
	task, err := service.CreateWeeklyTask(userID, WeeklyTaskInput{
		StageGoalID: &stage.ID,
		Title:       "周任务",
		Note:        "周备注",
		WeekStart:   weekStart,
		WeekEnd:     &weekEnd,
		Deadline:    &deadline,
	})
	if err != nil {
		t.Fatalf("create weekly task: %v", err)
	}

	updated, err := service.UpdateWeeklyTask(userID, task.ID, WeeklyTaskInput{
		Status:  "doing",
		Touched: map[string]bool{"status": true},
	})
	if err != nil {
		t.Fatalf("patch weekly status: %v", err)
	}
	if updated.Note != "周备注" || updated.StageGoalID == nil || updated.WeekEnd == nil || updated.Deadline == nil {
		t.Fatalf("omitted weekly fields were not preserved: %#v", updated)
	}

	cleared, err := service.UpdateWeeklyTask(userID, task.ID, WeeklyTaskInput{
		Note:        "",
		StageGoalID: nil,
		WeekEnd:     nil,
		Deadline:    nil,
		Touched: map[string]bool{
			"note":          true,
			"stage_goal_id": true,
			"week_end":      true,
			"deadline":      true,
		},
	})
	if err != nil {
		t.Fatalf("clear weekly fields: %v", err)
	}
	if cleared.Note != "" || cleared.StageGoalID != nil || cleared.WeekEnd != nil || cleared.Deadline != nil {
		t.Fatalf("weekly nullable fields were not cleared: %#v", cleared)
	}
}

func TestStageItemPatchPreservesAndClearsOptionalFields(t *testing.T) {
	service := newTestDailyTaskService(t)
	userID := uint(1)
	start := "2026-06-01"
	end := "2026-06-30"
	goal, err := service.CreateStageGoal(userID, StageGoalInput{Title: "长期目标"})
	if err != nil {
		t.Fatalf("create stage goal: %v", err)
	}
	item, err := service.CreateStageItem(userID, StageItemInput{
		StageGoalID: goal.ID,
		Title:       "资料分析专项",
		Note:        "保持手感",
		StartDate:   &start,
		EndDate:     &end,
		TargetText:  "稳定 85%",
	})
	if err != nil {
		t.Fatalf("create stage item: %v", err)
	}

	updated, err := service.UpdateStageItem(userID, item.ID, StageItemInput{
		Status:  "done",
		Touched: map[string]bool{"status": true},
	})
	if err != nil {
		t.Fatalf("patch stage item: %v", err)
	}
	if updated.Note != "保持手感" || updated.StartDate == nil || updated.EndDate == nil || updated.TargetText != "稳定 85%" {
		t.Fatalf("omitted item fields were not preserved: %#v", updated)
	}

	cleared, err := service.UpdateStageItem(userID, item.ID, StageItemInput{
		Note:       "",
		StartDate:  nil,
		EndDate:    nil,
		TargetText: "",
		Touched: map[string]bool{
			"note":        true,
			"start_date":  true,
			"end_date":    true,
			"target_text": true,
		},
	})
	if err != nil {
		t.Fatalf("clear stage item: %v", err)
	}
	if cleared.Note != "" || cleared.StartDate != nil || cleared.EndDate != nil || cleared.TargetText != "" {
		t.Fatalf("stage item optional fields were not cleared: %#v", cleared)
	}
}

func TestWeeklyTaskBindsStageItemAndDerivesStageGoal(t *testing.T) {
	service := newTestDailyTaskService(t)
	userID := uint(1)
	goal, err := service.CreateStageGoal(userID, StageGoalInput{Title: "长期目标"})
	if err != nil {
		t.Fatalf("create stage goal: %v", err)
	}
	item, err := service.CreateStageItem(userID, StageItemInput{StageGoalID: goal.ID, Title: "资料分析专项"})
	if err != nil {
		t.Fatalf("create stage item: %v", err)
	}

	task, err := service.CreateWeeklyTask(userID, WeeklyTaskInput{
		StageItemID: &item.ID,
		Title:       "本周资料分析",
		WeekStart:   "2026-06-15",
	})
	if err != nil {
		t.Fatalf("create weekly task: %v", err)
	}
	if task.StageItemID == nil || *task.StageItemID != item.ID {
		t.Fatalf("stage item was not bound: %#v", task)
	}
	if task.StageGoalID == nil || *task.StageGoalID != goal.ID {
		t.Fatalf("stage goal was not derived from stage item: %#v", task)
	}
	if task.TaskKind != "standard_week" {
		t.Fatalf("unexpected task kind: %s", task.TaskKind)
	}
}

func TestMovingStageItemSyncsAttachedWeeklyAndDailyTasks(t *testing.T) {
	service := newTestDailyTaskService(t)
	userID := uint(1)
	oldGoal, err := service.CreateStageGoal(userID, StageGoalInput{Title: "旧目标"})
	if err != nil {
		t.Fatalf("create old goal: %v", err)
	}
	newGoal, err := service.CreateStageGoal(userID, StageGoalInput{Title: "新目标"})
	if err != nil {
		t.Fatalf("create new goal: %v", err)
	}
	item, err := service.CreateStageItem(userID, StageItemInput{StageGoalID: oldGoal.ID, Title: "阶段子项"})
	if err != nil {
		t.Fatalf("create stage item: %v", err)
	}
	weekly, err := service.CreateWeeklyTask(userID, WeeklyTaskInput{
		StageItemID: &item.ID,
		Title:       "周任务",
		WeekStart:   "2026-06-15",
	})
	if err != nil {
		t.Fatalf("create weekly task: %v", err)
	}
	if _, err := service.MaterializeWeeklyTask(userID, weekly.ID, WeeklyTaskMaterializeInput{
		Mode: "scheduled_day",
		Date: ptrString("2026-06-16"),
	}); err != nil {
		t.Fatalf("materialize weekly task: %v", err)
	}

	if _, err := service.UpdateStageItem(userID, item.ID, StageItemInput{
		StageGoalID: newGoal.ID,
		Touched:     map[string]bool{"stage_goal_id": true},
	}); err != nil {
		t.Fatalf("move stage item: %v", err)
	}

	var movedWeekly models.WeeklyTask
	if err := service.db.First(&movedWeekly, weekly.ID).Error; err != nil {
		t.Fatalf("fetch moved weekly task: %v", err)
	}
	if movedWeekly.StageGoalID == nil || *movedWeekly.StageGoalID != newGoal.ID {
		t.Fatalf("weekly task did not sync moved stage goal: %#v", movedWeekly)
	}
	tasks, err := service.ListDailyTasks(userID, "", false, "", "")
	if err != nil {
		t.Fatalf("list daily tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].StageGoalID == nil || *tasks[0].StageGoalID != newGoal.ID {
		t.Fatalf("daily task did not sync moved stage goal: %#v", tasks)
	}
}

func TestUpdateStageGoalPatchPreservesAndClearsOptionalFields(t *testing.T) {
	service := newTestDailyTaskService(t)
	userID := uint(1)
	start := "2026-06-01"
	end := "2026-06-30"
	goal, err := service.CreateStageGoal(userID, StageGoalInput{
		Title:      "阶段目标",
		Note:       "阶段备注",
		StartDate:  &start,
		EndDate:    &end,
		TargetText: "稳定 70+",
	})
	if err != nil {
		t.Fatalf("create stage goal: %v", err)
	}

	updated, err := service.UpdateStageGoal(userID, goal.ID, StageGoalInput{
		Status:  "done",
		Touched: map[string]bool{"status": true},
	})
	if err != nil {
		t.Fatalf("patch stage status: %v", err)
	}
	if updated.Note != "阶段备注" || updated.StartDate == nil || updated.EndDate == nil || updated.TargetText != "稳定 70+" {
		t.Fatalf("omitted stage fields were not preserved: %#v", updated)
	}

	cleared, err := service.UpdateStageGoal(userID, goal.ID, StageGoalInput{
		Note:       "",
		StartDate:  nil,
		EndDate:    nil,
		TargetText: "",
		Touched: map[string]bool{
			"note":        true,
			"start_date":  true,
			"end_date":    true,
			"target_text": true,
		},
	})
	if err != nil {
		t.Fatalf("clear stage fields: %v", err)
	}
	if cleared.Note != "" || cleared.StartDate != nil || cleared.EndDate != nil || cleared.TargetText != "" {
		t.Fatalf("stage optional fields were not cleared: %#v", cleared)
	}
}

func TestMaterializeWeeklyTaskCreatesAndUpdatesIdempotently(t *testing.T) {
	service := newTestDailyTaskService(t)
	userID := uint(1)
	weekStart := "2026-06-15"
	weekEnd := "2026-06-21"
	deadline := "2026-06-20"
	stage, err := service.CreateStageGoal(userID, StageGoalInput{Title: "阶段目标"})
	if err != nil {
		t.Fatalf("create stage goal: %v", err)
	}
	weekly, err := service.CreateWeeklyTask(userID, WeeklyTaskInput{
		StageGoalID: &stage.ID,
		Title:       "刷题复盘",
		Note:        "周任务备注",
		WeekStart:   weekStart,
		WeekEnd:     &weekEnd,
		Deadline:    &deadline,
		Priority:    "high",
	})
	if err != nil {
		t.Fatalf("create weekly task: %v", err)
	}

	first, err := service.MaterializeWeeklyTask(userID, weekly.ID, WeeklyTaskMaterializeInput{
		Mode:    "split_to_days",
		Dates:   []string{"2026-06-16", "2026-06-17", "2026-06-17"},
		Replace: true,
	})
	if err != nil {
		t.Fatalf("materialize weekly task: %v", err)
	}
	if first.CreatedCount != 2 || first.UpdatedCount != 0 || len(first.DailyTasks) != 2 {
		t.Fatalf("unexpected first materialize result: %#v", first)
	}
	if first.DailyTasks[0].Title != "刷题复盘" || first.DailyTasks[0].StageGoalID == nil || *first.DailyTasks[0].StageGoalID != stage.ID {
		t.Fatalf("daily task did not inherit weekly fields: %#v", first.DailyTasks[0])
	}
	if first.DailyTasks[0].Origin != "weekly_materialized" {
		t.Fatalf("daily task origin was not marked: %#v", first.DailyTasks[0])
	}

	renamed, err := service.UpdateWeeklyTask(userID, weekly.ID, WeeklyTaskInput{
		Title:   "刷题复盘改名",
		Touched: map[string]bool{"title": true},
	})
	if err != nil {
		t.Fatalf("rename weekly task: %v", err)
	}
	second, err := service.MaterializeWeeklyTask(userID, renamed.ID, WeeklyTaskMaterializeInput{
		Mode:    "split_to_days",
		Dates:   []string{"2026-06-16", "2026-06-17"},
		Replace: true,
	})
	if err != nil {
		t.Fatalf("materialize weekly task again: %v", err)
	}
	if second.CreatedCount != 0 || second.UpdatedCount != 2 || len(second.DailyTasks) != 2 {
		t.Fatalf("unexpected second materialize result: %#v", second)
	}

	tasks, err := service.ListDailyTasks(userID, "", false, "", "")
	if err != nil {
		t.Fatalf("list daily tasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("materialization created duplicates, got %d tasks", len(tasks))
	}
	for _, task := range tasks {
		if task.Title != "刷题复盘改名" {
			t.Fatalf("existing daily task was not synced from weekly title: %#v", task)
		}
	}

	third, err := service.MaterializeWeeklyTask(userID, renamed.ID, WeeklyTaskMaterializeInput{
		Mode:    "split_to_days",
		Dates:   []string{"2026-06-16"},
		Replace: true,
	})
	if err != nil {
		t.Fatalf("replace materialized daily tasks: %v", err)
	}
	if third.DeletedCount != 1 || third.UpdatedCount != 1 || third.CreatedCount != 0 {
		t.Fatalf("unexpected replace result: %#v", third)
	}
	tasks, err = service.ListDailyTasks(userID, "", false, "", "")
	if err != nil {
		t.Fatalf("list daily tasks after replace: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("replace should remove stale materialized tasks, got %d tasks", len(tasks))
	}
}

func TestMaterializeWeeklyTaskRejectsOutOfRangeDatesAtomically(t *testing.T) {
	service := newTestDailyTaskService(t)
	userID := uint(1)
	weekly, err := service.CreateWeeklyTask(userID, WeeklyTaskInput{
		Title:     "周任务",
		WeekStart: "2026-06-15",
	})
	if err != nil {
		t.Fatalf("create weekly task: %v", err)
	}

	_, err = service.MaterializeWeeklyTask(userID, weekly.ID, WeeklyTaskMaterializeInput{
		Mode:  "split_to_days",
		Dates: []string{"2026-06-16", "2026-06-30"},
	})
	if err == nil {
		t.Fatal("expected out-of-range materialization to fail")
	}
	tasks, err := service.ListDailyTasks(userID, "", false, "", "")
	if err != nil {
		t.Fatalf("list daily tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("transaction was not atomic, got %#v", tasks)
	}
}

func TestSpecialProjectWeeklyTaskOverlapsWeekAndMaterializesWithinWindow(t *testing.T) {
	service := newTestDailyTaskService(t)
	userID := uint(1)
	start := "2026-06-10"
	end := "2026-06-20"
	goal, err := service.CreateStageGoal(userID, StageGoalInput{Title: "长期目标"})
	if err != nil {
		t.Fatalf("create stage goal: %v", err)
	}
	item, err := service.CreateStageItem(userID, StageItemInput{StageGoalID: goal.ID, Title: "申论模板专项"})
	if err != nil {
		t.Fatalf("create stage item: %v", err)
	}
	weekly, err := service.CreateWeeklyTask(userID, WeeklyTaskInput{
		StageItemID: &item.ID,
		TaskKind:    "special_project",
		Title:       "申论模板短项",
		StartDate:   &start,
		EndDate:     &end,
	})
	if err != nil {
		t.Fatalf("create special project: %v", err)
	}

	overlapping, err := service.ListWeeklyTasks(userID, "2026-06-15", "", "")
	if err != nil {
		t.Fatalf("list overlapping tasks: %v", err)
	}
	if len(overlapping) != 1 || overlapping[0].ID != weekly.ID {
		t.Fatalf("special project should overlap target week: %#v", overlapping)
	}

	result, err := service.MaterializeWeeklyTask(userID, weekly.ID, WeeklyTaskMaterializeInput{
		Mode: "scheduled_day",
		Date: ptrString("2026-06-18"),
	})
	if err != nil {
		t.Fatalf("materialize special project: %v", err)
	}
	if result.CreatedCount != 1 || len(result.DailyTasks) != 1 {
		t.Fatalf("unexpected materialize result: %#v", result)
	}
	if result.DailyTasks[0].StageItemID == nil || *result.DailyTasks[0].StageItemID != item.ID {
		t.Fatalf("daily task did not inherit stage item: %#v", result.DailyTasks[0])
	}

	_, err = service.MaterializeWeeklyTask(userID, weekly.ID, WeeklyTaskMaterializeInput{
		Mode: "scheduled_day",
		Date: ptrString("2026-06-22"),
	})
	if err == nil {
		t.Fatal("expected out-of-window materialization to fail")
	}
	tasks, err := service.ListDailyTasks(userID, "", false, "", "")
	if err != nil {
		t.Fatalf("list daily tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("failed materialization should be atomic, got %#v", tasks)
	}
}

func newTestDailyTaskService(t *testing.T) *DailyTaskService {
	t.Helper()
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.StageGoal{}, &models.StageItem{}, &models.WeeklyTask{}, &models.DailyTask{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewDailyTaskService(db)
}

func ptrString(value string) *string {
	return &value
}
