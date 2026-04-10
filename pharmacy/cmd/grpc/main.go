package grpc

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/Eucastan/hms/pharmacy/internal/configs"
	"github.com/Eucastan/hms/pharmacy/internal/grpc/server"
	"github.com/Eucastan/hms/pharmacy/internal/models"
	"github.com/Eucastan/hms/pharmacy/internal/repositories"
	"github.com/Eucastan/hms/pharmacy/internal/services"
	"github.com/Eucastan/hms/shared/pkg/grpcserver"
	"github.com/Eucastan/hms/shared/pkg/logger"
	pharm "github.com/Eucastan/hms/shared/pkg/proto/pharmacy"
)

func main() {
	logger.Init()
	defer logger.Sync()

	log := logger.Log

	log.Info("Starting Pharmacy gRPC Server...")

	cfg, err := configs.Load()
	if err != nil {
		log.Fatal("Failed to load configs", zap.Error(err))
	}

	db, err := configs.ConnectDB(cfg.DSN)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}

	if err := db.AutoMigrate(&models.Dispense{}); err != nil {
		log.Fatal("failed to migrate database", zap.Error(err))
	}

	dispense := repositories.NewDispenseRepository(db)
	dispenseSvc := services.NewDispenseService(dispense, log)

	listenAddr, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	defer listenAddr.Close()

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.AuthInterceptor(cfg.JWTSecret)),
	)
	pharm.RegisterPharmacyServiceServer(
		grpcServer,
		server.NewPharmacyServer(dispenseSvc),
	)

	go func() {
		log.Info("gRPC server listening", zap.String("port", cfg.GRPCPort))
		if err := grpcServer.Serve(listenAddr); err != nil {
			log.Fatal("gRPC server failed", zap.Error(err))
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

	log.Info("Pharmacy gRPC server shutdown completed")
}
