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

	"github.com/Eucastan/hms/clinical/internal/configs"
	"github.com/Eucastan/hms/clinical/internal/grpc/client"
	"github.com/Eucastan/hms/clinical/internal/handlers"
	"github.com/Eucastan/hms/clinical/internal/models"
	"github.com/Eucastan/hms/clinical/internal/repositories"
	"github.com/Eucastan/hms/clinical/internal/services"
	"github.com/Eucastan/hms/shared/pkg/auth"
	"github.com/Eucastan/hms/shared/pkg/healthcheck"
	"github.com/Eucastan/hms/shared/pkg/logger"
)

func main() {
	logger.Init()
	defer logger.Sync()

	log := logger.Log
	log.Info("Starting Clinical service...")

	cfg, err := configs.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	db, err := configs.ConnectDB(cfg.DSN)
	if err != nil {
		log.Fatal("Database connection failed", zap.Error(err))
	}

	if err := db.AutoMigrate(&models.Diagnosis{}); err != nil {
		log.Fatal("Failed to migrate Diagnosis model", zap.Error(err))
	} else {
		log.Info("Database schema migrated successfully")
	}

	// Diagnosis server
	diagnosRepo := repositories.NewDiagnosisRepo(db)
	diagnosisService := services.NewDiagnosisService(diagnosRepo, log)
	diagnosHandler := handlers.NewDiagnosisHandler(diagnosisService)

	clients := client.NewSendToAllClient(cfg.GRPCPatient, cfg.GRPCLab, cfg.GRPCPharm)
	clientsHandler := handlers.NewGRPCClientHandler(*clients)

	r := gin.Default()
	r.Use(auth.RateLimiter())
	r.Use(auth.AuthMiddleware(cfg.JWTSecret))

	healthHandler := healthcheck.NewHealthCheckHandlers(cfg.ServiceName, cfg.Version)
	r.GET("/health", healthHandler.Health)
	r.GET("/liveness", healthHandler.Liveness)
	r.GET("/readiness", healthHandler.Readiness)

	diagnoses := r.Group("api", auth.RequiredRole("doctor"))
	{
		// Diagnosis routes
		diagnoses.POST("/v1/diagnoses", diagnosHandler.Create)
		diagnoses.GET("/v1/diagnoses/:id", diagnosHandler.GetDiagnosis)
		diagnoses.PUT("/v1/diagnoses/:id", diagnosHandler.UpdateDiagnosis)
		diagnoses.DELETE("/v1/diagnoses/:id", diagnosHandler.DeleteDiagnosis)
	}

	c := r.Group("/api/v1", auth.RequiredRole("doctor"))
	{
		// GRPC Clients routes
		c.POST("/prescription", auth.RequiredRole("doctor", "pharmacist"), clientsHandler.CreatePrescription)
		c.POST("/lab-request", auth.RequiredRole("doctor", "labtech"), clientsHandler.CreateLabRequest)
		c.PUT("/lab-request/:id", auth.RequiredRole("doctor", "labtech"), clientsHandler.UpdateLabRequest)
		c.GET("/patient/:id", auth.RequiredRole("doctor", "admin"), clientsHandler.GetPatient)
		c.GET("/patients/search", auth.RequiredRole("doctor", "admin"), clientsHandler.SearchPatient)
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

	log.Info("Clinical service server shutdown completed successfully")
}
