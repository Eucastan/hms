package services

import (
	"context"

	"github.com/Eucastan/hms/lab/internal/models"
	"github.com/Eucastan/hms/lab/internal/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type LabRequestService interface {
	CreateLabRequest(ctx context.Context, patientID, requestBy uuid.UUID, testType, priority, status, notes string) (*models.LabTestRequest, error)
	GetLabRequestsByPatient(ctx context.Context, patientID uuid.UUID) ([]*models.LabTestRequest, error)
	UpdateLabRequest(ctx context.Context, id, requestBy uuid.UUID, inputs *models.LabRequestUpdate) error
	DeleteLabRequest(ctx context.Context, id uuid.UUID) error
}

type LabRequestSvc struct {
	repo   repositories.LabRequestRepository
	logger *zap.Logger
}

func NewLabRequestService(repo repositories.LabRequestRepository, logger *zap.Logger) LabRequestService {
	return &LabRequestSvc{repo: repo, logger: logger}
}

func (s *LabRequestSvc) CreateLabRequest(ctx context.Context, patientID, requestBy uuid.UUID, testType, priority, status, notes string) (*models.LabTestRequest, error) {

	labReq := &models.LabTestRequest{
		ID:        uuid.New(),
		PatientID: patientID,
		RequestBy: requestBy,
		TestType:  testType,
		Priority:  priority,
		Status:    status,
		Notes:     notes,
	}

	if err := s.repo.Create(ctx, labReq); err != nil {
		s.logger.Error("Failed to create lab request", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Lab request created successfully", zap.String("lab_request_id", labReq.ID.String()))
	return labReq, nil
}

func (s *LabRequestSvc) GetLabRequestsByPatient(ctx context.Context, patientID uuid.UUID) ([]*models.LabTestRequest, error) {
	labReqs, err := s.repo.FindByPatientID(ctx, patientID)
	if err != nil {
		s.logger.Error("Failed to fetch lab requests", zap.Error(err))
		return nil, err
	}

	return labReqs, nil
}

func (s *LabRequestSvc) UpdateLabRequest(ctx context.Context, id, requestBy uuid.UUID, input *models.LabRequestUpdate) error {

	if err := models.Validate.Struct(input); err != nil {
		s.logger.Warn("Invalid input for updating lab request", zap.Error(err))
		return err
	}

	updates := map[string]any{}

	if input.TestType != nil {
		updates["test_type"] = *input.TestType
	}
	if input.Priority != nil {
		updates["priority"] = *input.Priority
	}
	if input.Notes != nil {
		updates["notes"] = *input.Notes
	}

	if len(updates) == 0 {
		return nil
	}

	if err := s.repo.Update(ctx, id, requestBy, updates); err != nil {
		s.logger.Error("Failed to update lab request", zap.Error(err))
		return err
	}

	s.logger.Info("Lab request updated successfully", zap.String("lab_request_id", id.String()))
	return nil

}

func (s *LabRequestSvc) DeleteLabRequest(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete lab request", zap.Error(err))
		return err
	}

	s.logger.Info("Lab request deleted successfully", zap.String("lab_request_id", id.String()))
	return nil

}
