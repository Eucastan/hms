package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Eucastan/hms/lab/internal/configs"
	"github.com/Eucastan/hms/lab/internal/handlers"
	"github.com/Eucastan/hms/lab/internal/models"
	"github.com/Eucastan/hms/lab/internal/repositories"
	"github.com/Eucastan/hms/lab/internal/services"
	"github.com/Eucastan/hms/shared/pkg/auth"
	"github.com/Eucastan/hms/shared/pkg/healthcheck"
	"github.com/Eucastan/hms/shared/pkg/logger"
)

func main() {
	logger.Init()
	defer logger.Sync()

	log := logger.Log
	log.Info("Starting Lab Service...")

	cfg, err := configs.Load()
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	db, err := configs.ConnectDB(cfg.DSN)
	if err != nil {
		log.Fatal("Database connection failed", zap.Error(err))
	}

	if err := db.AutoMigrate(&models.LabResult{}, &models.LabTestRequest{}); err != nil {
		log.Fatal("Failed to migrate models", zap.Error(err))
	} else {
		log.Info("Database schema migrated successfully")
	}

	// Lab Request server
	labReqRepo := repositories.NewLabRequestRepo(db)
	labReqService := services.NewLabRequestService(labReqRepo, log)
	labReqHandler := handlers.NewLabRequestHandler(labReqService)

	// Lab Result server
	labResultRepo := repositories.NewLabResultRepo(db)
	labResultService := services.NewLabResultService(labResultRepo, log)
	labResultHandler := handlers.NewLabResultHandler(labResultService)

	r := gin.Default()
	r.Use(auth.RateLimiter())
	r.Use(auth.AuthMiddleware(cfg.JWTSecret))

	healthHandler := healthcheck.NewHealthCheckHandlers(cfg.ServiceName, cfg.Version)
	r.GET("/health", healthHandler.Health)
	r.GET("/liveness", healthHandler.Liveness)
	r.GET("/readiness", healthHandler.Readiness)

	r.Group("api/v1")
	{
		// Lab request routes
		r.POST("/labrequest", auth.RequiredRole("labtech", "doctor"), labReqHandler.CreateLabRequest)
		r.GET("/labrequest/:id", auth.RequiredRole("labtech"), labReqHandler.GetLabRequest)
		r.PUT("/labrequest/:id", auth.RequiredRole("labtech"), labReqHandler.UpdateLabRequest)
		r.DELETE("/labrequest/:id", auth.RequiredRole("labtech"), labReqHandler.DeleteLabRequest)

		// Lab result routes
		r.POST("/labresult", auth.RequiredRole("labtech"), labResultHandler.CreateLabResult)
		r.GET("/labresult/:id", auth.RequiredRole("labtech"), labResultHandler.GetLabResult)
		r.PUT("/labresult/:id", auth.RequiredRole("labtech"), labResultHandler.UpdateLabResult)
		r.DELETE("/labresult/:id", auth.RequiredRole("labtech"), labResultHandler.DeleteLabResult)
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

	log.Info("Lab service server shutdown completed successfully")
}
