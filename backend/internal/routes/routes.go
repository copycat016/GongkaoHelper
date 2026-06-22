package routes

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gkweb/backend/internal/config"
	"gkweb/backend/internal/handlers"
	"gkweb/backend/internal/middleware"
	"gkweb/backend/internal/services"
)

func Register(router *gin.Engine, db *gorm.DB, cfg config.Config) {
	healthHandler := handlers.NewHealthHandler()
	dbHandler := handlers.NewDBHandler(db)
	authService := services.NewAuthService(db, cfg)
	authHandler := handlers.NewAuthHandler(authService)
	llmHandler := handlers.NewLLMHandler(services.NewLLMService(db))
	promptHandler := handlers.NewPromptHandler(services.NewPromptService(db))
	pomodoroHandler := handlers.NewPomodoroHandler(services.NewPomodoroService(db))
	studyHandler := handlers.NewStudyHandler(services.NewStudyService(db))
	taskHandler := handlers.NewDailyTaskHandler(services.NewDailyTaskService(db))
	musicHandler := handlers.NewMusicHandler(services.NewMusicService(db))
	ocrHandler := handlers.NewOCRHandler(services.NewBaiduOCRService(db, cfg))
	backupHandler := handlers.NewBackupHandler(services.NewBackupService(db))
	essayHandler := handlers.NewEssayHandler(services.NewEssayService(db))
	pdfHandler := handlers.NewPDFHandler()
	themeHandler := handlers.NewThemeHandler(services.NewThemeService(db))

	// 音频文件不会被修改（文件名含时间戳），设置长缓存
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/uploads/music/") {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		c.Next()
	})
	router.Static("/uploads", "./uploads")

	api := router.Group("/api")
	{
		api.GET("/health", healthHandler.Health)
		api.GET("/db/ping", dbHandler.Ping)

		authPublic := api.Group("/auth")
		{
			authPublic.POST("/login", authHandler.Login)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthRequired(db, cfg))

		authProtected := protected.Group("/auth")
		{
			authProtected.GET("/me", authHandler.Me)
		}

		llm := protected.Group("/llm")
		{
			llm.GET("/providers", llmHandler.ListProviders)
			llm.POST("/providers", llmHandler.CreateProvider)
			llm.GET("/providers/:id/models", llmHandler.FetchProviderModels)
			llm.PUT("/providers/:id", llmHandler.UpdateProvider)
			llm.DELETE("/providers/:id", llmHandler.DeleteProvider)

			llm.GET("/models", llmHandler.ListModels)
			llm.POST("/models", llmHandler.CreateModel)
			llm.PUT("/models/:id", llmHandler.UpdateModel)
			llm.DELETE("/models/:id", llmHandler.DeleteModel)
		}

		prompts := protected.Group("/prompts")
		{
			prompts.GET("", promptHandler.List)
			prompts.POST("", promptHandler.Create)
			prompts.PUT("/:id", promptHandler.Update)
			prompts.DELETE("/:id", promptHandler.Delete)
		}

		pomodoro := protected.Group("/pomodoro")
		{
			pomodoro.POST("/sessions", pomodoroHandler.CreateSession)
			pomodoro.GET("/stats/today", pomodoroHandler.TodayStats)
		}

		logs := protected.Group("/logs")
		{
			logs.GET("", studyHandler.ListLogs)
			logs.GET("/stats", studyHandler.LogStats)
			logs.POST("", studyHandler.CreateLog)
		}

		tasks := protected.Group("/tasks")
		{
			tasks.GET("", taskHandler.List)
			tasks.GET("/summary", taskHandler.Summary)
			tasks.POST("", taskHandler.Create)
			tasks.PUT("/:id", taskHandler.Update)
			tasks.POST("/:id/toggle", taskHandler.Toggle)
			tasks.DELETE("/:id", taskHandler.Delete)
		}

		planning := protected.Group("/planning")
		{
			planning.GET("/daily-tasks", taskHandler.ListDailyTasks)
			planning.POST("/daily-tasks", taskHandler.CreateDailyTask)
			planning.PUT("/daily-tasks/:id", taskHandler.Update)
			planning.POST("/daily-tasks/:id/toggle", taskHandler.Toggle)
			planning.DELETE("/daily-tasks/:id", taskHandler.Delete)
			planning.GET("/weekly-tasks", taskHandler.ListWeeklyTasks)
			planning.POST("/weekly-tasks", taskHandler.CreateWeeklyTask)
			planning.PUT("/weekly-tasks/:id", taskHandler.UpdateWeeklyTask)
			planning.POST("/weekly-tasks/:id/materialize", taskHandler.MaterializeWeeklyTask)
			planning.DELETE("/weekly-tasks/:id", taskHandler.DeleteWeeklyTask)
			planning.GET("/stage-goals", taskHandler.ListStageGoals)
			planning.POST("/stage-goals", taskHandler.CreateStageGoal)
			planning.PUT("/stage-goals/:id", taskHandler.UpdateStageGoal)
			planning.DELETE("/stage-goals/:id", taskHandler.DeleteStageGoal)
			planning.GET("/stage-items", taskHandler.ListStageItems)
			planning.POST("/stage-items", taskHandler.CreateStageItem)
			planning.PUT("/stage-items/:id", taskHandler.UpdateStageItem)
			planning.DELETE("/stage-items/:id", taskHandler.DeleteStageItem)
		}

		music := protected.Group("/music")
		{
			music.GET("/playlists", musicHandler.ListPlaylists)
			music.POST("/playlists", musicHandler.CreatePlaylist)
			music.PUT("/playlists/:playlist_id", musicHandler.UpdatePlaylist)
			music.DELETE("/playlists/:playlist_id", musicHandler.DeletePlaylist)
			music.GET("/tracks", musicHandler.ListTracks)
			music.POST("/tracks", musicHandler.UploadTrack)
			music.DELETE("/tracks/:track_id", musicHandler.DeleteTrack)
			music.GET("/tracks/:track_id/playlists", musicHandler.TrackPlaylists)
			music.POST("/tracks/:track_id/metadata/lookup", musicHandler.LookupTrackMetadata)
			music.PUT("/tracks/:track_id/metadata", musicHandler.ApplyTrackMetadata)
			music.POST("/tracks/:track_id/lyrics/fetch", musicHandler.FetchTrackLyrics)
			music.GET("/playlists/:playlist_id/tracks", musicHandler.PlaylistTracks)
			music.POST("/playlists/:playlist_id/tracks/:track_id", musicHandler.AddTrackToPlaylist)
			music.DELETE("/playlists/:playlist_id/tracks/:track_id", musicHandler.RemoveTrackFromPlaylist)
			music.PUT("/playlists/:playlist_id/sort", musicHandler.UpdatePlaylistSort)
		}

		ocr := protected.Group("/ocr")
		{
			ocr.GET("/engines", ocrHandler.Engines)
			ocr.GET("/scenes", ocrHandler.Scenes)
			ocr.GET("/config", ocrHandler.Config)
			ocr.PUT("/config", ocrHandler.UpdateConfig)
			ocr.GET("/usage/month", ocrHandler.MonthUsage)
			ocr.POST("/recognize", ocrHandler.Recognize)
		}

		backup := protected.Group("/backup")
		{
			backup.GET("/export", backupHandler.Export)
		}

		essay := protected.Group("/essay")
		{
			essay.GET("/documents", essayHandler.ListDocuments)
			essay.POST("/documents", essayHandler.CreateDocument)
			essay.DELETE("/documents/:id", essayHandler.DeleteDocument)
			essay.POST("/documents/:id/parse", essayHandler.ParseDocument)
			essay.POST("/documents/:id/debug-boundary", essayHandler.DebugBoundary)
			essay.GET("/documents/:id/sections", essayHandler.ListSections)
			essay.GET("/documents/:id/chunks", essayHandler.ListChunks)
			essay.POST("/documents/:id/classify", essayHandler.ClassifyChunks)
			essay.POST("/documents/:id/assemble", essayHandler.AssembleQuestions)
			essay.GET("/documents/:id/questions", essayHandler.ListQuestions)
			essay.POST("/questions", essayHandler.CreateQuestion)
			essay.PUT("/questions/:id", essayHandler.UpdateQuestion)
			essay.DELETE("/questions/:id", essayHandler.DeleteQuestion)
			essay.POST("/questions/:id/relations", essayHandler.ReplaceQuestionRelations)
			essay.POST("/questions/:id/review", essayHandler.ReviewAnswer)
			essay.PUT("/sections/:id", essayHandler.UpdateSection)
		}

		pdf := protected.Group("/pdf")
		{
			pdf.GET("/parser-info", pdfHandler.ParserInfo)
			pdf.POST("/parse-tool", pdfHandler.ParseTool)
			pdf.POST("/parse-test", pdfHandler.ParseTest)
		}

		theme := protected.Group("/theme")
		{
			theme.GET("", themeHandler.Get)
			theme.POST("", themeHandler.Save)
		}
	}
}
