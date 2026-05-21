package models

import "time"

// 三层规划：StageGoal → WeeklyTask → DailyTask。
//
// 关联是单向「子指向父」，且都可选；上层节点可以独立存在；
// 下层节点也可以脱离父节点（临时任务）。
//
// 进度由子项完成情况聚合得到，不在父节点里冗余存数。

// StageGoal 阶段目标，例如「3 月底前行测平均 65+」「申论本月每周练 3 篇」。
type StageGoal struct {
	BaseModel
	Title        string     `json:"title" gorm:"size:200;not null"`
	Note         string     `json:"note" gorm:"type:text"`
	StartDate    *time.Time `json:"start_date" gorm:"index"`
	EndDate      *time.Time `json:"end_date" gorm:"index"`
	TargetMetric string     `json:"target_metric" gorm:"size:200"` // 自由文本目标指标
	Status       string     `json:"status" gorm:"size:20;index;default:active"`
	SortOrder    int        `json:"sort_order" gorm:"index;default:0"`
}

// WeeklyTask 周任务。week_start 是该周的周一本地零点，用作分组键。
//
// stage_goal_id 可空：允许临时插入与阶段目标无关的周任务。
type WeeklyTask struct {
	BaseModel
	StageGoalID   *uint      `json:"stage_goal_id" gorm:"index"`
	Title         string     `json:"title" gorm:"size:200;not null"`
	Note          string     `json:"note" gorm:"type:text"`
	WeekStart     time.Time  `json:"week_start" gorm:"index;not null"`
	DueDate       *time.Time `json:"due_date" gorm:"index"`
	Priority      string     `json:"priority" gorm:"size:20;index;default:normal"`
	Status        string     `json:"status" gorm:"size:20;index;default:todo"` // todo / doing / done / cancelled
	TargetMinutes int        `json:"target_minutes" gorm:"default:0"`
	SortOrder     int        `json:"sort_order" gorm:"index;default:0"`
}

// DailyTask 日任务。
//
// 字段说明：
//   - PlanDate：计划完成的日期。允许空：「有 DDL 但今天不一定做」就把 plan_date 留空，
//     前端可以放进「待安排」池。
//   - DueDate：硬 DDL，可空。和 plan_date 区分：plan_date 是「打算哪天做」，
//     due_date 是「最迟交付时间」。
//   - WeeklyTaskID / StageGoalID：归属，可空，独立任务也允许。
//   - Done / DoneAt：完成态。
type DailyTask struct {
	BaseModel
	WeeklyTaskID *uint      `json:"weekly_task_id" gorm:"index"`
	StageGoalID  *uint      `json:"stage_goal_id" gorm:"index"`
	Title        string     `json:"title" gorm:"size:200;not null"`
	Note         string     `json:"note" gorm:"type:text"`
	PlanDate     *time.Time `json:"plan_date" gorm:"index"`
	DueDate      *time.Time `json:"due_date" gorm:"index"`
	Priority     string     `json:"priority" gorm:"size:20;index;default:normal"`
	Done         bool       `json:"done" gorm:"index;default:false"`
	DoneAt       *time.Time `json:"done_at"`
	SortOrder    int        `json:"sort_order" gorm:"index;default:0"`
}
