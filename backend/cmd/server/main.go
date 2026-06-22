package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/config"
	"gkweb/backend/internal/database"
	"gkweb/backend/internal/middleware"
	"gkweb/backend/internal/routes"
	"gkweb/backend/internal/services"
)

// 嵌入前端构建产物。
// 构建镜像时会把 ./dist 拷贝到 backend/cmd/server/web/。
// 本地开发可能 web 目录为空（仅含 .gitkeep），此时不会启用静态文件服务。
//
//go:embed all:web
var webFS embed.FS

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required when GIN_MODE=release")
	}
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
	if err := services.NewAuthService(db, cfg).BootstrapSingleOwner(); err != nil {
		log.Fatalf("bootstrap owner user: %v", err)
	}

	router := gin.New()
	router.Use(middleware.RequestID())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	routes.Register(router, db, cfg)

	registerEmbeddedWeb(router)

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("start server: %v", err)
	}
}

// registerEmbeddedWeb 把嵌入的前端 dist 暴露在根路径，
// 并对前端路由（非 /api、/uploads）回退到 index.html，支持 SPA。
func registerEmbeddedWeb(router *gin.Engine) {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Printf("embedded web disabled: %v", err)
		return
	}
	// 检查是否包含 index.html，没有就跳过（开发态构建未拷贝 dist 时）
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		log.Printf("embedded web disabled: index.html not found in embedded fs")
		return
	}

	httpFS := http.FS(sub)
	fileServer := http.FileServer(httpFS)

	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// API 与上传路径不应进入 SPA 回退
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/uploads/") {
			c.Status(http.StatusNotFound)
			return
		}
		// 已存在的静态资源（assets/*.js、*.css、图片等）直接由 FileServer 处理
		trimmed := strings.TrimPrefix(path, "/")
		if trimmed != "" {
			if f, err := httpFS.Open(trimmed); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}
		// 其余前端路由统一回退到 index.html
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
