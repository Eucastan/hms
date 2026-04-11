package test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Eucastan/hms/lab/internal/configs"
	"github.com/Eucastan/hms/lab/internal/handlers"
	"github.com/Eucastan/hms/lab/internal/repositories"
	"github.com/Eucastan/hms/lab/internal/services"
	"github.com/Eucastan/hms/shared/pkg/auth"
)

type TestServer struct {
	*httptest.Server
	Router *gin.Engine
	DB     *gorm.DB
}

func NewLabTestServer(t *testing.T, db *gorm.DB) *TestServer {
	cfg := configs.Config{
		JWTSecret: "test-secret-very-long-32-chars-min",
	}

	logger := zap.NewNop()
	repo := repositories.NewLabRequestRepo(db)
	svc := services.NewLabRequestService(repo, logger)
	handler := handlers.NewLabRequestHandler(svc)

	resultrepo := repositories.NewLabResultRepo(db)
	resultsvc := services.NewLabResultService(resultrepo, logger)
	resulthandler := handlers.NewLabResultHandler(resultsvc)

	r := gin.Default()
	mw := auth.AuthMiddleware(cfg.JWTSecret)

	protected := r.Group("/api/v1", mw)
	{
		protected.POST("/lab-request", auth.RequiredRole("lab-tech", "doctor"), handler.CreateLabRequest)
		protected.GET("/lab-request/:id", auth.RequiredRole("lab-tech", "doctor"), handler.GetLabRequest)
		protected.DELETE("/lab-request/:id", auth.RequiredRole("lab-tech", "doctor"), handler.DeleteLabRequest)

		protected.POST("/lab-result", auth.RequiredRole("lab-tech"), resulthandler.CreateLabResult)
		protected.GET("/lab-result/:id", auth.RequiredRole("lab-tech", "doctor"), resulthandler.GetLabResult)
		protected.DELETE("/lab-result/:id", auth.RequiredRole("lab-tech"), resulthandler.DeleteLabResult)
	}

	return &TestServer{
		Server: httptest.NewServer(r),
		Router: r,
		DB:     db,
	}
}
