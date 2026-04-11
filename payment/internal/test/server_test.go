package test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Eucastan/hms/payment/internal/configs"
	"github.com/Eucastan/hms/payment/internal/handlers"
	"github.com/Eucastan/hms/payment/internal/repositories"
	"github.com/Eucastan/hms/payment/internal/services"
	"github.com/Eucastan/hms/shared/pkg/auth"
)

type TestServer struct {
	*httptest.Server
	Router *gin.Engine
	DB     *gorm.DB
}

func NewPaymentTestServer(t *testing.T, db *gorm.DB) *TestServer {
	cfg := configs.Config{
		JWTSecret: "test-secret-very-long-32-chars-min",
	}

	logger := zap.NewNop()
	billRepo := repositories.NewBillingRepository(db)
	Svc := services.NewBillingService(billRepo, logger)
	handler := handlers.NewBillingHandler(Svc)

	r := gin.Default()
	mw := auth.AuthMiddleware(cfg.JWTSecret)

	protected := r.Group("/api/v1", mw)
	{
		protected.POST("/bill", auth.RequiredRole("accountant"), handler.CreateBillCharge)
		protected.GET("/invoice/:id", auth.RequiredRole("accountant"), handler.GetInvoice)
		protected.POST("/refund", auth.RequiredRole("accountant"), handler.RefundBillCharge)
	}

	return &TestServer{
		Server: httptest.NewServer(r),
		Router: r,
		DB:     db,
	}
}
