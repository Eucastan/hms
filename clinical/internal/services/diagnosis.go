package services

import (
	"context"
	"errors"

	"github.com/Eucastan/hms/clinical/internal/dto"
	"github.com/Eucastan/hms/clinical/internal/models"
	"github.com/Eucastan/hms/clinical/internal/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DiagnosisService interface {
	CreateDiagnosis(ctx context.Context, doctorID uuid.UUID, input *dto.DiagnosisCreateRequest) (*models.Diagnosis, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Diagnosis, error)
	UpdateDiagnosis(ctx context.Context, id, doctorID uuid.UUID, updates map[string]interface{}) error
	DeleteDiagnosis(ctx context.Context, id uuid.UUID) error
}

type DiagnosisSvc struct {
	repo   repositories.DiagnosisRepository
	logger *zap.Logger
}

func NewDiagnosisService(repo repositories.DiagnosisRepository, logger *zap.Logger) DiagnosisService {
	return &DiagnosisSvc{repo: repo, logger: logger}
}

func (s *DiagnosisSvc) CreateDiagnosis(ctx context.Context, doctorID uuid.UUID, input *dto.DiagnosisCreateRequest) (*models.Diagnosis, error) {
	if err := dto.Validate.Struct(input); err != nil {
		return nil, err
	}

	diagnosis := &models.Diagnosis{
		ID:          uuid.New(),
		PatientID:   input.PatientID,
		AdmissionID: input.AdmissionID,
		DoctorID:    doctorID,
		Code:        input.Code,
		Description: input.Description,
	}

	if err := s.repo.Create(ctx, diagnosis); err != nil {
		return nil, err
	}

	return diagnosis, nil

}

func (s *DiagnosisSvc) FindByID(ctx context.Context, id uuid.UUID) (*models.Diagnosis, error) {

	if resp, err := s.repo.FindByID(ctx, id); err != nil {

		if errors.Is(err, repositories.ErrNotFound) {
			return nil, repositories.ErrNotFound
		}

		return nil, err

	} else {
		return resp, nil
	}

}

func (s *DiagnosisSvc) UpdateDiagnosis(ctx context.Context, id, doctorID uuid.UUID, updates map[string]interface{}) error {

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	// Prevent updating protected fields
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	delete(updates, "patient_id")
	delete(updates, "doctor_id")
	delete(updates, "admission_id")

	if len(updates) == 0 {
		return errors.New("no valid fields to update")
	}

	return s.repo.Update(ctx, id, doctorID, updates)
}

func (s *DiagnosisSvc) DeleteDiagnosis(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
