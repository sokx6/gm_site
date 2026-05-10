package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"

	"gm_site/internal/config"
	"gm_site/internal/database"
	"gm_site/internal/handler"
	"gm_site/internal/logger"
	"gm_site/internal/middleware"
	"gm_site/internal/repository"
	"gm_site/internal/service"
	ws "gm_site/internal/websocket"
)

func main() {
	configPath := flag.String("config", "config.local.yaml", "path to config file")
	logPath := flag.String("log", "", "log file path (default: derived from config name)")
	flag.Parse()

	var logFile string
	if *logPath != "" {
		logFile = *logPath
	} else {
		cfgName := strings.TrimSuffix(filepath.Base(*configPath), filepath.Ext(*configPath))
		logFile = fmt.Sprintf("/var/log/gm_site/%s.log", cfgName)
	}

	logger.Init(logFile)

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.L.Error("config load failed", "err", err)
		os.Exit(1)
	}

	// 连接数据库
	db, err := database.NewDatabase(cfg.Database.Path)
	if err != nil {
		logger.L.Error("database connect failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.L.Info("database connected")

	// 执行迁移
	if err := database.RunMigrations(db, "migrations"); err != nil {
		logger.L.Error("migrations failed", "err", err)
		os.Exit(1)
	}
	logger.L.Info("migrations applied")

	// 解析站点运行起始日期
	startDate, err := time.Parse("2006-01-02", cfg.Site.StartDate)
	if err != nil {
		logger.L.Error("site start_date parse failed", "err", err)
		os.Exit(1)
	}

	// 初始化依赖
	jwtService := service.NewJWTService(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpire,
		cfg.JWT.RefreshExpire,
	)

	emailSvc := service.NewSmtpEmailService(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
		cfg.SMTP.AdminEmail,
	)

	lskyClient := service.NewLskyClient(
		cfg.Lsky.BaseURL,
		cfg.Lsky.Email,
		cfg.Lsky.Password,
	)

	// 初始化仓库层
	userRepo := repository.NewUserRepository(db)
	albumRepo := repository.NewAlbumRepository(db)
	imageRepo := repository.NewImageRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	friendRepo := repository.NewFriendRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(userRepo, jwtService, emailSvc, cfg.Admin.Email)
	albumHandler := handler.NewAlbumHandler(albumRepo)
	imageHandler := handler.NewImageHandler(imageRepo, lskyClient, cfg.Upload.MaxSizeMB)
	commentHandler := handler.NewCommentHandler(commentRepo, imageRepo, userRepo, notificationRepo, emailSvc, db)
	friendSvc := service.NewFriendService(friendRepo, userRepo, emailSvc, notificationRepo)
	friendHandler := handler.NewFriendHandler(friendSvc, userRepo, notificationRepo)
	adminHandler := handler.NewAdminHandler(userRepo, emailSvc)
	siteHandler := handler.NewSiteHandler(cfg.Site.Name)

	// 初始化 WebSocket Hub
	wsHub := ws.NewHub()
	go wsHub.Run()

	// 初始化统计服务
	statsService := service.NewStatsService(wsHub, db, startDate)
	statsService.StartBroadcastLoop(10 * time.Second)

	// 创建 Echo 实例
	e := echo.New()

	// CORS 中间件
	e.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	// 请求日志中间件
	e.Use(middleware.LoggerMiddleware(logger.L))

	// 访客统计中间件（跳过健康检查）
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Path() == "/api/health" {
				return next(c)
			}
			go statsService.IncrementVisitor()
			return next(c)
		}
	})

	// ── 公开路由（支持可选认证） ──────────────────────────
	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/api/albums", albumHandler.ListAlbums, middleware.OptionalAuth(jwtService))
	e.GET("/api/images", imageHandler.ListImages, middleware.OptionalAuth(jwtService))
	e.GET("/api/images/search", imageHandler.SearchImages, middleware.OptionalAuth(jwtService))
	e.GET("/api/images/:id", imageHandler.GetImage)
	e.GET("/api/images/:id/comments", commentHandler.ListByImage)
	e.GET("/api/site", siteHandler.Info)

	// ── 认证路由（无需登录） ──────────────────────────────
	e.POST("/api/auth/register", authHandler.Register)
	e.POST("/api/auth/login", authHandler.Login)
	e.POST("/api/auth/refresh", authHandler.Refresh)

	// ── 受保护路由（需要登录） ────────────────────────────
	auth := e.Group("/api")
	auth.Use(middleware.AuthRequired(jwtService))
	auth.GET("/auth/me", authHandler.Me)
	auth.POST("/albums", albumHandler.CreateAlbum)
	auth.PUT("/albums/:id", albumHandler.UpdateAlbum)
	auth.PUT("/albums/:id/privacy", albumHandler.UpdatePrivacy)
	auth.DELETE("/albums/:id", albumHandler.DeleteAlbum)
	auth.POST("/images/upload", imageHandler.UploadImage)
	auth.PUT("/images/:id", imageHandler.UpdateImage)
	auth.PUT("/images/:id/privacy", imageHandler.UpdatePrivacy)
	auth.DELETE("/images/:id", imageHandler.DeleteImage)
	auth.POST("/images/:id/comments", commentHandler.Create)
	auth.DELETE("/comments/:id", commentHandler.Delete)
	auth.POST("/friends/request", friendHandler.SendRequest)
	auth.PUT("/friends/request/:id/accept", friendHandler.AcceptRequest)
	auth.PUT("/friends/request/:id/reject", friendHandler.RejectRequest)
	auth.GET("/friends/requests", friendHandler.GetRequests)
	auth.GET("/friends", friendHandler.GetFriends)
	auth.DELETE("/friends/:id", friendHandler.DeleteFriend)
	auth.GET("/notifications", friendHandler.GetNotifications)
	auth.PUT("/notifications/:id/read", friendHandler.MarkNotificationRead)
	auth.POST("/comments/:id/reply", commentHandler.Reply)

	// ── 管理员路由（需要登录 + admin 角色） ───────────────
	admin := e.Group("/api/admin")
	admin.Use(middleware.AuthRequired(jwtService), middleware.AdminRequired())
	admin.GET("/users/pending", adminHandler.ListPending)
	admin.PUT("/users/:id/approve", adminHandler.ApproveUser)
	admin.PUT("/users/:id/reject", adminHandler.RejectUser)

	// ── WebSocket ─────────────────────────────────────────
	e.GET("/api/ws", handler.ServeWS(wsHub))

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()
	stop()
	e.Logger.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		e.Logger.Fatal(err)
	}
}
