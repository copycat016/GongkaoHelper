package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gkweb/backend/internal/config"
	"gkweb/backend/internal/handlers"
	"gkweb/backend/internal/services"
)

func Register(router *gin.Engine, db *gorm.DB) {
	cfg := config.Load()
	healthHandler := handlers.NewHealthHandler()
	dbHandler := handlers.NewDBHandler(db)
	llmHandler := handlers.NewLLMHandler(services.NewLLMService(db))
	promptHandler := handlers.NewPromptHandler(services.NewPromptService(db))
	mistakeHandler := handlers.NewMistakeHandler(services.NewMistakeService(db))
	pomodoroHandler := handlers.NewPomodoroHandler(services.NewPomodoroService(db))
	studyHandler := handlers.NewStudyHandler(services.NewStudyService(db))
	musicHandler := handlers.NewMusicHandler(services.NewMusicService(db))
	ocrHandler := handlers.NewOCRHandler(services.NewBaiduOCRService(db, cfg))
	backupHandler := handlers.NewBackupHandler(services.NewBackupService(db))
	essayHandler := handlers.NewEssayHandler(services.NewEssayService(db))
	pdfHandler := handlers.NewPDFHandler()

	router.Static("/uploads", "./uploads")

	api := router.Group("/api")
	{
		api.GET("/health", healthHandler.Health)
		api.GET("/db/ping", dbHandler.Ping)

		llm := api.Group("/llm")
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

		prompts := api.Group("/prompts")
		{
			prompts.GET("", promptHandler.List)
			prompts.POST("", promptHandler.Create)
			prompts.PUT("/:id", promptHandler.Update)
			prompts.DELETE("/:id", promptHandler.Delete)
		}

		mistakes := api.Group("/mistakes")
		{
			mistakes.GET("", mistakeHandler.List)
			mistakes.POST("", mistakeHandler.Create)
			mistakes.GET("/:id", mistakeHandler.Get)
			mistakes.PUT("/:id", mistakeHandler.Update)
			mistakes.DELETE("/:id", mistakeHandler.Delete)
			mistakes.POST("/:id/review", mistakeHandler.Review)
		}

		pomodoro := api.Group("/pomodoro")
		{
			pomodoro.POST("/sessions", pomodoroHandler.CreateSession)
			pomodoro.GET("/stats/today", pomodoroHandler.TodayStats)
		}

		logs := api.Group("/logs")
		{
			logs.GET("", studyHandler.ListLogs)
			logs.GET("/stats", studyHandler.LogStats)
			logs.POST("", studyHandler.CreateLog)
		}

		plans := api.Group("/plans")
		{
			plans.GET("", studyHandler.ListPlans)
			plans.POST("", studyHandler.CreatePlan)
			plans.PUT("/:id", studyHandler.UpdatePlan)
			plans.DELETE("/:id", studyHandler.DeletePlan)
			plans.POST("/:id/complete", studyHandler.CompletePlan)
		}

		calendar := api.Group("/calendar")
		{
			calendar.GET("/events", studyHandler.CalendarEvents)
		}

		music := api.Group("/music")
		{
			music.GET("/playlists", musicHandler.ListPlaylists)
			music.POST("/playlists", musicHandler.CreatePlaylist)
			music.GET("/tracks", musicHandler.ListTracks)
			music.POST("/tracks", musicHandler.UploadTrack)
			music.GET("/playlists/:playlist_id/tracks", musicHandler.PlaylistTracks)
			music.POST("/playlists/:playlist_id/tracks/:track_id", musicHandler.AddTrackToPlaylist)
		}

		ocr := api.Group("/ocr")
		{
			ocr.GET("/engines", ocrHandler.Engines)
			ocr.GET("/scenes", ocrHandler.Scenes)
			ocr.GET("/config", ocrHandler.Config)
			ocr.PUT("/config", ocrHandler.UpdateConfig)
			ocr.GET("/usage/month", ocrHandler.MonthUsage)
			ocr.POST("/recognize", ocrHandler.Recognize)
		}

		backup := api.Group("/backup")
		{
			backup.GET("/export", backupHandler.Export)
		}

		essay := api.Group("/essay")
		{
			essay.GET("/documents", essayHandler.ListDocuments)
			essay.POST("/documents", essayHandler.CreateDocument)
			essay.POST("/documents/:id/parse", essayHandler.ParseDocument)
			essay.GET("/documents/:id/chunks", essayHandler.ListChunks)
			essay.POST("/documents/:id/classify", essayHandler.ClassifyChunks)
			essay.POST("/documents/:id/assemble", essayHandler.AssembleQuestions)
			essay.GET("/documents/:id/questions", essayHandler.ListQuestions)
			essay.POST("/questions/:id/review", essayHandler.ReviewAnswer)
		}

		pdf := api.Group("/pdf")
		{
			pdf.POST("/parse-test", pdfHandler.ParseTest)
		}
	}
}
