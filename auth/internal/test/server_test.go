package test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Eucastan/hms/auth/internal/configs"
	"github.com/Eucastan/hms/auth/internal/handlers"
	"github.com/Eucastan/hms/auth/internal/repositories"
	"github.com/Eucastan/hms/auth/internal/services"
	"github.com/Eucastan/hms/shared/pkg/auth"
)

type TestServer struct {
	*httptest.Server
	Router *gin.Engine
	DB     *gorm.DB
}

func NewAuthServer(t *testing.T, db *gorm.DB) *TestServer {
	cfg := &configs.Config{
		JWTSecret: "testsecret",
	}

	repo := repositories.NewAuthRepo(db)
	logger := zap.NewNop()

	svc := services.NewAuthSvc(repo, cfg, logger)
	handler := handlers.NewAuthHandler(svc)

	r := gin.Default()
	mw := auth.AuthMiddleware(cfg.JWTSecret)

	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)

	protected := r.Group("admin/users", mw)
	{
		protected.DELETE("/:id", auth.RequiredRole("admin"), handler.Delete)
	}

	return &TestServer{
		Server: httptest.NewServer(r),
		Router: r,
		DB:     db,
	}
}
