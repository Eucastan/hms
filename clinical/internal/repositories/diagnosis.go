package repositories

import (
	"context"
	"errors"

	"github.com/Eucastan/hms/clinical/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiagnosisRepository interface {
	Create(ctx context.Context, diagnosis *models.Diagnosis) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Diagnosis, error)
	Update(ctx context.Context, id, doctorID uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type DiagnosisRepo struct {
	DB *gorm.DB
}

var ErrNotFound = errors.New("record not found")

func NewDiagnosisRepo(db *gorm.DB) DiagnosisRepository {
	return &DiagnosisRepo{DB: db}
}

func (r *DiagnosisRepo) Create(ctx context.Context, diagnosis *models.Diagnosis) error {
	return r.DB.WithContext(ctx).Create(diagnosis).Error
}

func (r *DiagnosisRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Diagnosis, error) {
	var dgnsis models.Diagnosis
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&dgnsis).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &dgnsis, nil
}

func (r *DiagnosisRepo) Update(ctx context.Context, id, doctorID uuid.UUID, updates map[string]interface{}) error {
	return r.DB.WithContext(ctx).Model(&models.Diagnosis{}).Where("id = ? AND doctor_id = ?", id, doctorID).Updates(updates).Error
}

func (r *DiagnosisRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.DB.WithContext(ctx).Where("id = ?", id).Delete(&models.Diagnosis{}).Error
}
