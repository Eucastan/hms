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

	"github.com/Eucastan/hms/patient/internal/configs"
	"github.com/Eucastan/hms/patient/internal/handlers"
	"github.com/Eucastan/hms/patient/internal/models"
	"github.com/Eucastan/hms/patient/internal/repositories"
	"github.com/Eucastan/hms/patient/internal/services"
	"github.com/Eucastan/hms/shared/pkg/auth"
	"github.com/Eucastan/hms/shared/pkg/healthcheck"
	"github.com/Eucastan/hms/shared/pkg/logger"
)

func main() {
	logger.Init()
	defer logger.Sync()

	log := logger.Log
	log.Info("Starting Patient Service...")

	cfg, err := configs.Load()
	if err != nil {
		log.Fatal("Error loading configuration", zap.Error(err))
	}

	db, err := configs.ConnectDB(cfg.DSN)
	if err != nil {
		log.Fatal("Failed to connect to database %v", zap.Error(err))
	}

	if err := db.AutoMigrate(&models.Patient{}, &models.Admission{}); err != nil {
		log.Fatal("Failed to migrate models", zap.Error(err))
	} else {
		log.Info("Database schema migrated successfully")
	}

	// Patient service
	patientRepo := repositories.NewPatientRepo(db)
	admissionRepo := repositories.NewAdmissionRepo(db)
	patientService := services.NewPatientService(patientRepo, admissionRepo, log)
	handler := handlers.NewPatientHandler(patientService)

	r := gin.Default()
	r.Use(auth.RateLimiter())
	r.Use(auth.AuthMiddleware(cfg.JWTSecret))

	healthHandler := healthcheck.NewHealthCheckHandlers(cfg.ServiceName, cfg.Version)
	r.GET("/health", healthHandler.Health)
	r.GET("/liveness", healthHandler.Liveness)
	r.GET("/readiness", healthHandler.Readiness)

	r.Group("/api/v1", auth.RequiredRole("admin", "doctor"))
	{
		// Patient records routes
		r.POST("/patient", handler.CreatePatientRecord)
		r.GET("/patient", handler.GetPatientRecord)
		r.PUT("/patient/:id", handler.UpdatePatientRecord)
		r.DELETE("/patient/:id", handler.DeletePatientRecord)

		// Admission routes
		r.GET("/admission", handler.GetAdmissionRecords)
		r.PUT("/admission/:id", handler.UpdateAdmissionRecord)
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

	log.Info("Patient service server shutdown completed successfully")

}
