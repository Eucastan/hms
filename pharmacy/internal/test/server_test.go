package test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Eucastan/hms/pharmacy/internal/configs"
	"github.com/Eucastan/hms/pharmacy/internal/grpc/client"
	"github.com/Eucastan/hms/pharmacy/internal/handlers"
	"github.com/Eucastan/hms/pharmacy/internal/repositories"
	"github.com/Eucastan/hms/pharmacy/internal/services"
	"github.com/Eucastan/hms/shared/pkg/auth"
)

type TestServer struct {
	*httptest.Server
	Router *gin.Engine
	DB     *gorm.DB
}

func NewPharmacyTestServer(t *testing.T, db *gorm.DB) *TestServer {
	grpcServer := StartGRPCTestServer(t)
	t.Cleanup(func() { grpcServer.Close() })

	cfg := configs.Config{
		JWTSecret:   "test-secret-very-long-32-chars-min",
		GRPCBilling: grpcServer.Addr,
		GRPCPatient: grpcServer.Addr,
	}

	logger := zap.NewNop()
	drugRepo := repositories.NewDrugRepository(db)
	drugSvc := services.NewDrugService(drugRepo, logger)
	drugHandler := handlers.NewDrugHandler(drugSvc)

	c := client.NewClients(cfg.GRPCBilling, cfg.GRPCPatient)
	repo := repositories.NewDispenseRepository(db)
	svc := services.NewDispenseService(repo, logger)

	// integration test
	h := handlers.NewDispenseHandler(svc, c)

	r := gin.Default()
	mw := auth.AuthMiddleware(cfg.JWTSecret)

	protected := r.Group("/api/v1", mw)
	{
		protected.POST("/drug", auth.RequiredRole("pharmacist"), drugHandler.CreateDrug)
		protected.GET("/drug/:id", auth.RequiredRole("pharmacist"), drugHandler.GetDrug)
		protected.DELETE("/drug/:id", auth.RequiredRole("pharmacist"), drugHandler.DeleteDrug)

		// Real gRPC server service-to-service tests (fake implementation)
		protected.POST("/dispense", auth.RequiredRole("pharmacist"), h.CreateDispense)
		protected.GET("/dispense/:id", auth.RequiredRole("pharmacist"), h.GetDispense)
	}

	return &TestServer{
		Server: httptest.NewServer(r),
		Router: r,
		DB:     db,
	}
}
