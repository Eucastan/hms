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

	"github.com/Eucastan/hms/patient/internal/configs"
	"github.com/Eucastan/hms/patient/internal/grpc/server"
	"github.com/Eucastan/hms/patient/internal/models"
	"github.com/Eucastan/hms/patient/internal/repositories"
	"github.com/Eucastan/hms/patient/internal/services"
	"github.com/Eucastan/hms/shared/pkg/grpcserver"
	"github.com/Eucastan/hms/shared/pkg/logger"
	patient "github.com/Eucastan/hms/shared/pkg/proto/patient"
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

	if err := db.AutoMigrate(&models.Patient{}, &models.Admission{}); err != nil {
		log.Fatal("Failed to migrate database", zap.Error(err))
	}

	patientRepo := repositories.NewPatientRepo(db)
	admissionRepo := repositories.NewAdmissionRepo(db)
	svc := services.NewPatientService(patientRepo, admissionRepo, log)

	listenAddr, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	defer listenAddr.Close()

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.AuthInterceptor(cfg.JWTSecret)),
	)
	patient.RegisterPatientServiceServer(grpcServer, server.NewPatientServer(svc))

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

	log.Info("Patient gRPC server shutdown completed")
}
