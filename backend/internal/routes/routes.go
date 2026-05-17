package routes

import (
	"strings"

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
	questionBankHandler := handlers.NewQuestionBankHandler(services.NewQuestionBankService(db))
	pomodoroHandler := handlers.NewPomodoroHandler(services.NewPomodoroService(db))
	studyHandler := handlers.NewStudyHandler(services.NewStudyService(db))
	musicHandler := handlers.NewMusicHandler(services.NewMusicService(db))
	ocrHandler := handlers.NewOCRHandler(services.NewBaiduOCRService(db, cfg))
	backupHandler := handlers.NewBackupHandler(services.NewBackupService(db))
	essayHandler := handlers.NewEssayHandler(services.NewEssayService(db))
	pdfHandler := handlers.NewPDFHandler()

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

		questions := api.Group("/questions")
		{
			questions.GET("", questionBankHandler.List)
			questions.GET("/:id", questionBankHandler.Get)
			questions.PUT("/:id", questionBankHandler.Update)
			questions.DELETE("/:id", questionBankHandler.Delete)
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
			essay.DELETE("/documents/:id", essayHandler.DeleteDocument)
			essay.POST("/documents/:id/parse", essayHandler.ParseDocument)
			essay.POST("/documents/:id/debug-boundary", essayHandler.DebugBoundary)
			essay.GET("/documents/:id/sections", essayHandler.ListSections)
			essay.GET("/documents/:id/chunks", essayHandler.ListChunks)
			essay.POST("/documents/:id/classify", essayHandler.ClassifyChunks)
			essay.POST("/documents/:id/assemble", essayHandler.AssembleQuestions)
			essay.GET("/documents/:id/questions", essayHandler.ListQuestions)
			essay.POST("/questions/:id/review", essayHandler.ReviewAnswer)
		}

		pdf := api.Group("/pdf")
		{
			pdf.POST("/parse-tool", pdfHandler.ParseTool)
			pdf.POST("/parse-test", pdfHandler.ParseTest)
		}
	}
}
