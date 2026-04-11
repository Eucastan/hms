package test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Eucastan/hms/clinical/internal/configs"
	"github.com/Eucastan/hms/clinical/internal/handlers"
	"github.com/Eucastan/hms/clinical/internal/repositories"
	"github.com/Eucastan/hms/clinical/internal/services"
	"github.com/Eucastan/hms/shared/pkg/auth"
)

type TestServer struct {
	*httptest.Server
	Router *gin.Engine
	DB     *gorm.DB
}

func NewClinicalTestServer(t *testing.T, db *gorm.DB) *TestServer {
	cfg := configs.Config{
		JWTSecret: "test-secret-very-long-32-chars-min",
	}

	logger := zap.NewNop()
	repo := repositories.NewDiagnosisRepo(db)
	Svc := services.NewDiagnosisService(repo, logger)
	handler := handlers.NewDiagnosisHandler(Svc)

	r := gin.Default()
	mw := auth.AuthMiddleware(cfg.JWTSecret)

	protected := r.Group("/api/v1", mw)
	{
		protected.POST("/diagnosis", auth.RequiredRole("doctor"), handler.Create)
		protected.GET("/diagnosis/:id", auth.RequiredRole("doctor"), handler.GetDiagnosis)
		protected.PUT("/diagnosis/:id", auth.RequiredRole("doctor"), handler.UpdateDiagnosis)
		protected.DELETE("/diagnosis/:id", auth.RequiredRole("doctor"), handler.DeleteDiagnosis)
	}

	return &TestServer{
		Server: httptest.NewServer(r),
		Router: r,
		DB:     db,
	}
}
