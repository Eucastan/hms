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

	"github.com/Eucastan/hms/pharmacy/internal/configs"
	"github.com/Eucastan/hms/pharmacy/internal/grpc/client"
	"github.com/Eucastan/hms/pharmacy/internal/handlers"
	"github.com/Eucastan/hms/pharmacy/internal/models"
	"github.com/Eucastan/hms/pharmacy/internal/repositories"
	"github.com/Eucastan/hms/pharmacy/internal/services"
	"github.com/Eucastan/hms/shared/pkg/auth"
	"github.com/Eucastan/hms/shared/pkg/healthcheck"
	"github.com/Eucastan/hms/shared/pkg/logger"
)

func main() {
	logger.Init()
	defer logger.Sync()

	log := logger.Log
	log.Info("Starting Pharmacy Service...")

	cfg, err := configs.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	db, err := configs.ConnectDB(cfg.DSN)
	if err != nil {
		log.Fatal("Database connection failed", zap.Error(err))
	}

	if err := db.AutoMigrate(&models.Drug{}, &models.Dispense{}); err != nil {
		log.Fatal("Failed to migrate models", zap.Error(err))
	} else {
		log.Info("Database schema migrated successfully")
	}

	// GRPC Client
	client := client.NewClients(cfg.GRPCBilling, cfg.GRPCPatient)

	// Drug server
	drugRepo := repositories.NewDrugRepository(db)
	drugService := services.NewDrugService(drugRepo, log)
	drugHandler := handlers.NewDrugHandler(drugService)

	// Dispense server
	dispenseRepo := repositories.NewDispenseRepository(db)
	dispenseService := services.NewDispenseService(dispenseRepo, log)
	dispenseHandler := handlers.NewDispenseHandler(dispenseService, client)

	r := gin.Default()
	r.Use(auth.RateLimiter())
	authMid := auth.AuthMiddleware(cfg.JWTSecret)

	healthHandler := healthcheck.NewHealthCheckHandlers(cfg.ServiceName, cfg.Version)
	r.GET("/health", healthHandler.Health)
	r.GET("/liveness", healthHandler.Liveness)
	r.GET("/readiness", healthHandler.Readiness)

	r.Group("/api/v1", authMid)
	{
		// Drug request routes
		r.POST("/drug", auth.RequiredRole("pharmacist"), drugHandler.CreateDrug)
		r.GET("/drug/:id", auth.RequiredRole("pharmacist"), drugHandler.GetDrug)
		r.PUT("/drug/:id", auth.RequiredRole("pharmacist"), drugHandler.UpdateDrug)
		r.DELETE("/drug/:id", auth.RequiredRole("pharmacist"), drugHandler.DeleteDrug)

		// Dispense result routes
		r.POST("/dispense", auth.RequiredRole("pharmacist"), dispenseHandler.CreateDispense)
		r.GET("/dispense/:id", auth.RequiredRole("pharmacist"), dispenseHandler.GetDispense)
		r.GET("/prescription/:id", auth.RequiredRole("pharmacist", "doctor"), dispenseHandler.GetPrescription)
		r.GET("/dispense/drug/:id", auth.RequiredRole("pharmacist"), dispenseHandler.GetDispenseByDrugID)
		r.PUT("/dispense/:id", auth.RequiredRole("pharmacist"), dispenseHandler.UpdateDispense)
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

	log.Info("Pharmacy service server shutdown completed successfully")
}
