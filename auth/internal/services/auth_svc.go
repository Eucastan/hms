package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Eucastan/hms/auth/internal/configs"
	"github.com/Eucastan/hms/auth/internal/models"
	"github.com/Eucastan/hms/auth/internal/repositories"
	"github.com/Eucastan/hms/auth/internal/security"
	"github.com/Eucastan/hms/shared/pkg/utils"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type AuthService interface {
	Register(ctx context.Context, user *models.StaffCreateRequest) (*models.Staff, error)
	Login(ctx context.Context, input *models.LoginRequest) (string, *models.Staff, error)
	UpdateStaff(ctx context.Context, id uuid.UUID, input *models.StaffUpdateRequest, updaterRole string) error
	DeleteStaff(ctx context.Context, id uuid.UUID, deleterRole string) error
}

type AuthSvc struct {
	repo   repositories.AuthRepository
	cfg    *configs.Config
	logger *zap.Logger
}

var ErrValidation = errors.New("validation error")
var ErrWeakPassword = errors.New("weak password")
var ErrInvalidCredentials = errors.New("invalid credentials")

func NewAuthSvc(repo repositories.AuthRepository, cfg *configs.Config, logger *zap.Logger) AuthService {
	return &AuthSvc{repo: repo, cfg: cfg, logger: logger}
}

func (s *AuthSvc) Register(ctx context.Context, input *models.StaffCreateRequest) (*models.Staff, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tracer := otel.Tracer("auth-service")

	ctx, span := tracer.Start(ctx, "RegisterUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("email", input.Email),
	)

	if err := models.Validate.Struct(input); err != nil {
		s.logger.Error("validation failed",
			zap.Error(err),
		)
		return nil, ErrValidation
	}

	if err := security.ValidatePasswordStrength(input.Password); err != nil {
		return nil, ErrWeakPassword
	}

	hashed, err := security.Hashpwd(input.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &models.Staff{
		ID:        uuid.New(),
		Email:     input.Email,
		Password:  string(hashed),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Role:      models.Role(input.Role),
	}

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, errors.New("email already exists")
		}
		s.logger.Error("failed to create user",
			zap.Error(err),
			zap.String("email", input.Email),
		)
		return nil, errors.New("failed to create user")
	}

	s.logger.Info("user created successfully",
		zap.String("email", input.Email),
	)
	return created, nil

}

func (s *AuthSvc) Login(ctx context.Context, input *models.LoginRequest) (string, *models.Staff, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := models.Validate.Struct(input); err != nil {
		s.logger.Error("validation failed",
			zap.Error(err),
		)
		return "", nil, ErrValidation
	}

	exists, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		s.logger.Error("invalid email",
			zap.Error(err),
			zap.String("email", input.Email),
		)
		return "", nil, ErrInvalidCredentials
	}

	isOk := security.Checkpwd(exists.Password, input.Password)
	if !isOk {
		s.logger.Error("invalid password",
			zap.Error(err),
			zap.String("email", input.Email),
		)
		return "", nil, ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(exists.ID.String(), string(exists.Role), s.cfg.JWTSecret)
	if err != nil {
		s.logger.Error("failed to generate token",
			zap.Error(err),
		)
		return "", nil, errors.New("failed to generate token")
	}

	s.logger.Info("login successful")
	return token, exists, nil
}

func (s *AuthSvc) UpdateStaff(ctx context.Context, id uuid.UUID, input *models.StaffUpdateRequest, updaterRole string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := models.Validate.Struct(input); err != nil {
		s.logger.Error("validation failed",
			zap.Error(err),
		)
		return err
	}

	_, err := s.repo.FindByID(ctx, id.String())
	if err != nil {
		s.logger.Error("failed to get user ID",
			zap.Error(err),
		)
		return err
	}

	updates := map[string]any{}

	if input.FirstName != nil {
		updates["first_name"] = *input.FirstName
	}
	if input.LastName != nil {
		updates["last_name"] = *input.LastName
	}
	if input.Role != nil {
		// Admin can change role, others cannot
		if updaterRole != string(models.Admin) {
			return errors.New("only admin can change role")
		}
		updates["role"] = *input.Role
	}
	if input.Active != nil {
		updates["active"] = *input.Active
	}
	if input.Password != nil {
		if err := security.ValidatePasswordStrength(*input.Password); err != nil {
			return err
		}
		hashed, err := security.Hashpwd(*input.Password)
		if err != nil {
			s.logger.Error("password hash error",
				zap.Error(err),
			)
			return err
		}
		updates["password"] = string(hashed)
	}

	if len(updates) == 0 {
		return nil
	}

	s.logger.Info("update successful")
	return s.repo.Update(ctx, id, updates)
}

func (s *AuthSvc) DeleteStaff(ctx context.Context, id uuid.UUID, deleterRole string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if deleterRole != string(models.Admin) {
		return errors.New("only admin can delete users")
	}

	s.logger.Error("user successfully deleted",
		zap.String("id", id.String()),
	)
	return s.repo.Delete(ctx, id.String())
}
