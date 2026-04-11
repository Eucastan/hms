package test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Eucastan/hms/patient/internal/configs"
	"github.com/Eucastan/hms/patient/internal/handlers"
	"github.com/Eucastan/hms/patient/internal/repositories"
	"github.com/Eucastan/hms/patient/internal/services"
	"github.com/Eucastan/hms/shared/pkg/auth"
)

type TestServer struct {
	*httptest.Server
	Router *gin.Engine
	DB     *gorm.DB
}

func NewPatientTestServer(t *testing.T, db *gorm.DB) *TestServer {
	cfg := configs.Config{
		JWTSecret: "test-secret-very-long-32-chars-min",
	}

	logger := zap.NewNop()
	ptRepo := repositories.NewPatientRepo(db)
	admsnRepo := repositories.NewAdmissionRepo(db)
	svc := services.NewPatientService(ptRepo, admsnRepo, logger)
	handler := handlers.NewPatientHandler(svc)

	r := gin.Default()
	mw := auth.AuthMiddleware(cfg.JWTSecret)

	protected := r.Group("/api/v1", mw)
	{
		protected.POST("/patient", auth.RequiredRole("admin"), handler.CreatePatientRecord)
		protected.GET("/patient/:id", auth.RequiredRole("admin"), handler.GetPatientRecord)
		protected.PUT("/patient/:id", auth.RequiredRole("admin"), handler.UpdatePatientRecord)
		protected.DELETE("/patient/:id", auth.RequiredRole("admin"), handler.DeletePatientRecord)
		protected.GET("/admission", auth.RequiredRole("admin"), handler.GetAdmissionRecords)
		protected.PUT("/admission", auth.RequiredRole("admin"), handler.UpdateAdmissionRecord)
	}

	return &TestServer{
		Server: httptest.NewServer(r),
		Router: r,
		DB:     db,
	}
}
