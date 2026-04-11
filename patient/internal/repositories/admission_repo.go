package repositories

import (
	"context"
	"errors"

	"github.com/Eucastan/hms/patient/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdmissionRepository interface {
	Create(ctx context.Context, admission *models.Admission) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Admission, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]any) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type AdmissionRepo struct {
	DB *gorm.DB
}

func NewAdmissionRepo(db *gorm.DB) AdmissionRepository {
	return &AdmissionRepo{DB: db}
}

func (r *AdmissionRepo) Create(ctx context.Context, admission *models.Admission) error {
	return r.DB.WithContext(ctx).Create(admission).Error
}

func (r *AdmissionRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Admission, error) {
	var admission models.Admission
	err := r.DB.WithContext(ctx).Preload("Patient").Where("id = ?", id).First(&admission).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("record not found")
	}

	return &admission, err
}

func (r *AdmissionRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	protected := []string{"id", "patient_id", "admitted_at", "created_at", "updated_at"}
	for _, key := range protected {
		delete(updates, key)
	}

	return r.DB.WithContext(ctx).
		Model(&models.Admission{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *AdmissionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.DB.WithContext(ctx).Where("id = ?", id).Delete(&models.Admission{}).Error
}
