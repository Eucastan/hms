package services

import (
	"context"

	"github.com/Eucastan/hms/lab/internal/models"
	"github.com/Eucastan/hms/lab/internal/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type LabResultService interface {
	CreateLabResult(ctx context.Context, performBy, patientID uuid.UUID, testType, resultValue, unit, referenceRange, comments string, verified bool) (*models.LabResult, error)
	GetByPatientID(ctx context.Context, patientID uuid.UUID) ([]*models.LabResult, error)
	UpdateLabResult(ctx context.Context, id, performedBy uuid.UUID, inputs *models.LabResultUpdate) error
	DeleteLabResult(ctx context.Context, id uuid.UUID) error
}

type LabResultSvc struct {
	repo   repositories.LabResultRepository
	logger *zap.Logger
}

func NewLabResultService(repo repositories.LabResultRepository, logger *zap.Logger) LabResultService {
	return &LabResultSvc{repo: repo, logger: logger}
}

func (s *LabResultSvc) CreateLabResult(ctx context.Context, performBy, patientID uuid.UUID, testType, resultValue, unit, referenceRange, comments string, verified bool) (*models.LabResult, error) {

	labResult := &models.LabResult{
		ID:             uuid.New(),
		PerformedBy:    performBy,
		PatientID:      patientID,
		TestType:       testType,
		ResultValue:    resultValue,
		Unit:           unit,
		ReferenceRange: referenceRange,
		Comments:       comments,
		Verified:       verified,
	}

	if err := s.repo.Create(ctx, labResult); err != nil {
		s.logger.Error("Failed to create lab result", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Lab result created successfully", zap.String("lab_result_id", labResult.ID.String()))
	return labResult, nil
}

func (s *LabResultSvc) GetByPatientID(ctx context.Context, patientID uuid.UUID) ([]*models.LabResult, error) {
	labResults, err := s.repo.FindByPatientID(ctx, patientID)
	if err != nil {
		s.logger.Error("Failed to fetch lab results", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Lab results fetched successfully", zap.String("patient_id", patientID.String()))
	return labResults, nil
}

func (s *LabResultSvc) UpdateLabResult(ctx context.Context, id, performedBy uuid.UUID, inputs *models.LabResultUpdate) error {

	updates := make(map[string]any)

	if inputs.ResultValue != nil {
		updates["result_value"] = &inputs.ResultValue
	}

	if inputs.Unit != nil {
		updates["unit"] = &inputs.Unit
	}

	if inputs.ReferenceRange != nil {
		updates["reference_range"] = &inputs.ReferenceRange
	}

	if inputs.Comments != nil {
		updates["comments"] = &inputs.Comments
	}

	if inputs.Verified != nil {
		updates["verified"] = &inputs.Verified
	}

	if err := s.repo.Update(ctx, id, performedBy, updates); err != nil {
		s.logger.Error("Failed to update lab result", zap.Error(err))
		return err
	}

	s.logger.Info("Lab result updated successfully", zap.String("lab_result_id", id.String()))
	return nil
}

func (s *LabResultSvc) DeleteLabResult(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete lab result", zap.Error(err))
		return err
	}

	s.logger.Info("Lab result deleted successfully", zap.String("lab_result_id", id.String()))
	return nil
}
