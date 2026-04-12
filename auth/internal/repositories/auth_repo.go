package repositories

import (
	"context"
	"errors"

	"github.com/Eucastan/hms/auth/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthRepository interface {
	Create(ctx context.Context, user *models.Staff) (*models.Staff, error)
	FindByEmail(ctx context.Context, email string) (*models.Staff, error)
	FindByID(ctx context.Context, id string) (*models.Staff, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]any) error
	Delete(ctx context.Context, id string) error
}

type AuthRepo struct {
	DB *gorm.DB
}

func NewAuthRepo(db *gorm.DB) AuthRepository {
	return &AuthRepo{DB: db}
}

func (r *AuthRepo) Create(ctx context.Context, user *models.Staff) (*models.Staff, error) {
	if err := r.DB.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *AuthRepo) FindByEmail(ctx context.Context, email string) (*models.Staff, error) {
	var user models.Staff
	err := r.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepo) FindByID(ctx context.Context, id string) (*models.Staff, error) {
	var user models.Staff
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}

		return nil, err
	}

	return &user, nil
}

func (r *AuthRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	// Protect critical fields
	delete(updates, "id")
	delete(updates, "email")
	delete(updates, "created_at")

	return r.DB.WithContext(ctx).
		Model(&models.Staff{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *AuthRepo) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return r.DB.WithContext(ctx).Where("id = ?", id).Delete(&models.Staff{ID: uid}).Error
}
