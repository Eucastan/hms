package repositories

import (
	"context"

	"github.com/Eucastan/hms/lab/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LabResultRepository interface {
	Create(ctx context.Context, lab *models.LabResult) error
	FindByPatientID(ctx context.Context, patientID uuid.UUID) ([]*models.LabResult, error)
	Update(ctx context.Context, id, performedBy uuid.UUID, updates map[string]any) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type LabResultRepo struct {
	DB *gorm.DB
}

func NewLabResultRepo(db *gorm.DB) LabResultRepository {
	return &LabResultRepo{DB: db}
}

func (r *LabResultRepo) Create(ctx context.Context, lab *models.LabResult) error {
	return r.DB.WithContext(ctx).Create(lab).Error
}

func (r *LabResultRepo) FindByPatientID(ctx context.Context, patientID uuid.UUID) ([]*models.LabResult, error) {
	var labs []*models.LabResult
	err := r.DB.WithContext(ctx).Where("patient_id = ?", patientID).Find(&labs).Error
	if err != nil {
		return nil, err
	}

	return labs, nil
}

func (r *LabResultRepo) Update(ctx context.Context, id, performedBy uuid.UUID, updates map[string]any) error {
	return r.DB.WithContext(ctx).
		Model(&models.LabResult{}).
		Where("id = ? AND performed_by = ?", id, performedBy).
		Updates(updates).Error
}

func (r *LabResultRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.DB.WithContext(ctx).Where("id = ?", id).Delete(&models.LabResult{}).Error
}
