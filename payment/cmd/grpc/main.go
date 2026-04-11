package http

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/Eucastan/hms/payment/internal/configs"
	"github.com/Eucastan/hms/payment/internal/grpc/server"
	"github.com/Eucastan/hms/payment/internal/models"
	"github.com/Eucastan/hms/payment/internal/repositories"
	"github.com/Eucastan/hms/payment/internal/services"
	"github.com/Eucastan/hms/shared/pkg/grpcserver"
	"github.com/Eucastan/hms/shared/pkg/logger"
	pay "github.com/Eucastan/hms/shared/pkg/proto/billing"
)

func main() {
	logger.Init()
	defer logger.Sync()

	log := logger.Log
	log.Info("Starting gRPC Patient Service...")

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
		log.Fatal("Failed to migrate database", zap.Error(err))
	}

	invoiceRepo := repositories.NewBillingRepository(db)
	invoiceService := services.NewBillingService(invoiceRepo, log)

	listenAddr, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatal("Failed to listen on port", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.AuthInterceptor(cfg.JWTSecret)),
	)
	pay.RegisterBillingServiceServer(grpcServer, server.NewBillingServer(invoiceService))

	go func() {
		if err := grpcServer.Serve(listenAddr); err != nil {
			log.Fatal("gRPC serve failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-quit
	log.Info("Shutdown signal received. Stopping gRPC server...")

	// Graceful stop with timeout
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	grpcServer.GracefulStop()

	// Close DB
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	log.Info("Payment gRPC server shutdown completed")
}
