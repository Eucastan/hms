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

	"github.com/Eucastan/hms/payment/internal/configs"
	"github.com/Eucastan/hms/payment/internal/handlers"
	"github.com/Eucastan/hms/payment/internal/models"
	"github.com/Eucastan/hms/payment/internal/repositories"
	"github.com/Eucastan/hms/payment/internal/services"
	"github.com/Eucastan/hms/shared/pkg/auth"
	"github.com/Eucastan/hms/shared/pkg/healthcheck"
	"github.com/Eucastan/hms/shared/pkg/logger"
)

func main() {
	logger.Init()
	defer logger.Sync()

	log := logger.Log
	log.Info("Starting Payment Service...")

	cfg, err := configs.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	db, err := configs.ConnectDB(cfg.DSN)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	if err := db.AutoMigrate(
		&models.Invoice{},
		&models.Payment{},
		&models.BillCharge{},
	); err != nil {
		log.Fatal("Failed to migrate models", zap.Error(err))
	} else {
		log.Info("Database schema migrated successfully")
	}

	billingRepo := repositories.NewBillingRepository(db)
	billingService := services.NewBillingService(billingRepo, log)
	billingHandler := handlers.NewBillingHandler(billingService)

	r := gin.Default()
	r.Use(auth.RateLimiter())
	api := r.Group("/api/v1")

	healthHandler := healthcheck.NewHealthCheckHandlers(cfg.ServiceName, cfg.Version)
	r.GET("/health", healthHandler.Health)
	r.GET("/liveness", healthHandler.Liveness)
	r.GET("/readiness", healthHandler.Readiness)

	auth := api.Use(auth.AuthMiddleware(cfg.JWTSecret), auth.RequiredRole("accountant"))
	{
		auth.GET("/invoice/:id", billingHandler.GetInvoice)
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

	log.Info("Payment service server shutdown completed successfully")
}
