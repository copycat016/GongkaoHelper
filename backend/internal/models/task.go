package models

import "time"

// 任务规划采用可选关联模型，但长期规划建议按目标 -> 阶段子项 -> 推进任务 -> 日任务拆解。
//
// StageGoal 表示长期或阶段性成果；StageItem 表示目标下的可交付子项；
// WeeklyTask 表示推进任务，既可以是标准周任务，也可以是中短期专项推进项；
// DailyTask 表示某天真正执行的事项。只设置 DDL 不等于设置执行日。

// StageGoal 阶段目标，例如「3 月底前行测平均 65+」「申论本月每周练 3 篇」。
type StageGoal struct {
	BaseModel
	Title        string     `json:"title" gorm:"size:200;not null"`
	Note         string     `json:"note" gorm:"type:text"`
	StartDate    *time.Time `json:"start_date" gorm:"index"`
	EndDate      *time.Time `json:"end_date" gorm:"index"`
	TargetText   string     `json:"target_text" gorm:"column:target_metric;size:200"`
	Status       string     `json:"status" gorm:"size:20;index;default:active"`
	ProgressMode string     `json:"progress_mode" gorm:"size:30;default:weekly"`
	SortOrder    int        `json:"sort_order" gorm:"index;default:0"`
}

// StageItem 是长期目标下的阶段子项，例如「资料分析稳定 85%」「申论小题模板成型」。
type StageItem struct {
	BaseModel
	StageGoalID  uint       `json:"stage_goal_id" gorm:"not null;index"`
	Title        string     `json:"title" gorm:"size:200;not null"`
	Note         string     `json:"note" gorm:"type:text"`
	StartDate    *time.Time `json:"start_date" gorm:"index"`
	EndDate      *time.Time `json:"end_date" gorm:"index"`
	TargetText   string     `json:"target_text" gorm:"column:target_metric;size:200"`
	Status       string     `json:"status" gorm:"size:20;index;default:active"`
	ProgressMode string     `json:"progress_mode" gorm:"size:30;default:weekly"`
	SortOrder    int        `json:"sort_order" gorm:"index;default:0"`
}

// WeeklyTask 是推进任务。TaskKind=standard_week 时表示标准周任务；
// TaskKind=special_project 时表示可跨周的中短期专项推进项。
// 推进任务默认只留在周计划/专项计划页；只有显式安排或拆解时才进入 DailyTask。
type WeeklyTask struct {
	BaseModel
	StageGoalID *uint      `json:"stage_goal_id" gorm:"index"`
	StageItemID *uint      `json:"stage_item_id" gorm:"index"`
	TaskKind    string     `json:"task_kind" gorm:"size:30;index;default:standard_week"`
	Title       string     `json:"title" gorm:"size:200;not null"`
	Note        string     `json:"note" gorm:"type:text"`
	WeekStart   time.Time  `json:"week_start" gorm:"index;not null"`
	WeekEnd     *time.Time `json:"week_end" gorm:"index"`
	StartDate   *time.Time `json:"start_date" gorm:"index"`
	EndDate     *time.Time `json:"end_date" gorm:"index"`
	Deadline    *time.Time `json:"deadline" gorm:"column:due_date;index"`
	Priority    string     `json:"priority" gorm:"size:20;index;default:normal"`
	Status      string     `json:"status" gorm:"size:20;index;default:todo"` // todo / doing / done / cancelled
	ExecuteMode string     `json:"execute_mode" gorm:"size:30;index;default:weekly_todo"`
	SortOrder   int        `json:"sort_order" gorm:"index;default:0"`
}

// DailyTask 日任务。Date 是计划执行日；Deadline 是硬 DDL。
// Date 允许为空，用于待安排池；只有 Date 等于某天时才进入对应日计划。
type DailyTask struct {
	BaseModel
	WeeklyTaskID     *uint      `json:"weekly_task_id" gorm:"index"`
	StageGoalID      *uint      `json:"stage_goal_id" gorm:"index"`
	StageItemID      *uint      `json:"stage_item_id" gorm:"index"`
	Title            string     `json:"title" gorm:"size:200;not null"`
	Note             string     `json:"note" gorm:"type:text"`
	Date             *time.Time `json:"date" gorm:"column:plan_date;index"`
	Deadline         *time.Time `json:"deadline" gorm:"column:due_date;index"`
	Status           string     `json:"status" gorm:"size:20;index;default:todo"`
	Priority         string     `json:"priority" gorm:"size:20;index;default:normal"`
	EstimatedMinutes int        `json:"estimated_minutes" gorm:"default:0"`
	Origin           string     `json:"origin" gorm:"size:40;index;default:manual"`
	Done             bool       `json:"done" gorm:"index;default:false"`
	DoneAt           *time.Time `json:"done_at"`
	SortOrder        int        `json:"sort_order" gorm:"index;default:0"`
}
