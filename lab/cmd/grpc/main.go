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

	"github.com/Eucastan/hms/lab/internal/configs"
	"github.com/Eucastan/hms/lab/internal/grpc/server"
	"github.com/Eucastan/hms/lab/internal/models"
	"github.com/Eucastan/hms/lab/internal/repositories"
	"github.com/Eucastan/hms/lab/internal/services"
	"github.com/Eucastan/hms/shared/pkg/grpcserver"
	"github.com/Eucastan/hms/shared/pkg/logger"
	labpb "github.com/Eucastan/hms/shared/pkg/proto/lab"
)

func main() {
	logger.Init()
	defer logger.Sync()

	log := logger.Log
	log.Info("Starting gRPC Lab Service...")

	cfg, err := configs.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	db, err := configs.ConnectDB(cfg.DSN)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	if err := db.AutoMigrate(&models.LabTestRequest{}, &models.LabResult{}); err != nil {
		log.Fatal("Failed to migrate database", zap.Error(err))
	}

	labReq := repositories.NewLabRequestRepo(db)
	labReqSvc := services.NewLabRequestService(labReq, log)

	// Lab Results
	labRes := repositories.NewLabResultRepo(db)
	labResSvc := services.NewLabResultService(labRes, log)

	listenAddr, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	defer listenAddr.Close()

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.AuthInterceptor(cfg.JWTSecret)),
	)
	labpb.RegisterLabServiceServer(grpcServer, server.NewLabServiceServer(labReqSvc, labResSvc))

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

	log.Info("Lab gRPC server shutdown completed")
}
