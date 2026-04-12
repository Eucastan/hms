package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Eucastan/hms/auth/internal/configs"
	"github.com/Eucastan/hms/auth/internal/handlers"
	"github.com/Eucastan/hms/auth/internal/models"
	"github.com/Eucastan/hms/auth/internal/repositories"
	"github.com/Eucastan/hms/auth/internal/security"
	"github.com/Eucastan/hms/auth/internal/services"
	"github.com/Eucastan/hms/shared/pkg/auth"
	"github.com/Eucastan/hms/shared/pkg/healthcheck"
	"github.com/Eucastan/hms/shared/pkg/logger"
	"github.com/Eucastan/hms/shared/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger.Init()
	defer logger.Sync()

	log := logger.Log
	log.Info("Starting Auth Service...")

	shutdown := tracing.InitTracer()
	defer shutdown(context.Background())

	cfg, err := configs.Load()
	if err != nil {
		logger.Log.Error("failed to load configuration", zap.Error(err))
	}

	db, err := configs.ConnectDB(cfg.DSN)
	if err != nil {
		logger.Log.Error("Database connection failed", zap.Error(err))
	}

	if err := db.AutoMigrate(&models.Staff{}); err != nil {
		logger.Log.Error("Failed to migrate Staff model", zap.Error(err))
	} else {
		logger.Log.Info("Database schema migrated successfully")
	}

	userRepo := repositories.NewAuthRepo(db)
	svc := services.NewAuthSvc(userRepo, cfg, log)
	authhandler := handlers.NewAuthHandler(svc)

	r := gin.Default()

	r.Use(auth.RateLimiter())
	r.Use(security.LoginRateLimiter())
	authMid := auth.AuthMiddleware(cfg.JWTSecret)

	// Public routes
	r.POST("/register", authhandler.Register)
	r.POST("/login", authhandler.Login)

	healthCheckHandler := healthcheck.NewHealthCheckHandlers(cfg.ServiceName, cfg.Version)
	r.GET("/health", healthCheckHandler.Health)
	r.GET("/liveness", healthCheckHandler.Liveness)
	r.GET("/readiness", healthCheckHandler.Readiness)

	// Protected routes
	protected := r.Group("/api", authMid)
	{
		protected.GET("/profile", authhandler.Profile)
		protected.PUT("/staff/:id", auth.RequiredRole("admin"), authhandler.UpdateStaff)
		protected.DELETE("/:id", auth.RequiredRole("admin"), authhandler.Delete)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: r,
	}

	go func() {
		log.Info("HTTPServer starting", zap.String("port", cfg.HTTPPort))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP Server failed to start", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-quit
	log.Info("Shutdown signal received. Starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	} else {
		log.Info("HTTP server stopped gracefully")
	}

	log.Info("Auth service server shutdown completed successfully")

}
