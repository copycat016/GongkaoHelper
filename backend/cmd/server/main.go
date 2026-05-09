package main

import (
	"log"
	"time"
	_ "time/tzdata"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/config"
	"gkweb/backend/internal/database"
	"gkweb/backend/internal/middleware"
	"gkweb/backend/internal/routes"
)

func main() {
	cfg := config.Load()
	location, err := time.LoadLocation(cfg.DBTimeZone)
	if err != nil {
		log.Fatalf("load timezone %s: %v", cfg.DBTimeZone, err)
	}
	time.Local = location
	gin.SetMode(cfg.GinMode)

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	router := gin.New()
	router.Use(middleware.RequestID())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	routes.Register(router, db)

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
