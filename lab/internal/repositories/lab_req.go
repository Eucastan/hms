package repositories

import (
	"context"
	"fmt"

	"github.com/Eucastan/hms/lab/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LabRequestRepository interface {
	Create(ctx context.Context, labReq *models.LabTestRequest) error
	FindByPatientID(ctx context.Context, patientID uuid.UUID) ([]*models.LabTestRequest, error)
	Update(ctx context.Context, id, requestedBy uuid.UUID, updates map[string]any) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type LabRequestRepo struct {
	DB *gorm.DB
}

func NewLabRequestRepo(db *gorm.DB) LabRequestRepository {
	return &LabRequestRepo{DB: db}
}

func (r *LabRequestRepo) Create(ctx context.Context, labReq *models.LabTestRequest) error {
	return r.DB.WithContext(ctx).Create(labReq).Error
}

func (r *LabRequestRepo) FindByPatientID(ctx context.Context, patientID uuid.UUID) ([]*models.LabTestRequest, error) {
	var labReqs []*models.LabTestRequest
	err := r.DB.WithContext(ctx).Where("patient_id = ?", patientID).Find(&labReqs).Error
	if err != nil {
		return nil, err
	}

	return labReqs, nil
}

func (r *LabRequestRepo) Update(ctx context.Context, id, requestBy uuid.UUID, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	// Prevent changing immutable fields
	delete(updates, "patient_id")
	delete(updates, "request_by")
	delete(updates, "created_at")

	result := r.DB.WithContext(ctx).
		Model(&models.LabTestRequest{}).
		Where("id = ? AND request_by = ?", id, requestBy).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("lab request not found or not owned by user")
	}

	return nil
}

func (r *LabRequestRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.DB.WithContext(ctx).Where("id = ?", id).Delete(&models.LabTestRequest{}).Error
}
