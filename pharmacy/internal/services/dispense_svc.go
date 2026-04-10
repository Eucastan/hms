package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Eucastan/hms/pharmacy/internal/models"
	"github.com/Eucastan/hms/pharmacy/internal/repositories"
)

type DispenseService interface {
	CreateDispense(
		ctx context.Context,
		patientID,
		dispenseBy,
		prescriptionID,
		drugID uuid.UUID,
		quantity int32,
		notes string,
	) (*models.Dispense, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Dispense, error)
	CreatePrescription(ctx context.Context, doctorID uuid.UUID, input *models.CreatePrescriptionRequest) error
	FindPrescriptionByID(ctx context.Context, id uuid.UUID) (*models.Prescription, error)
	UpdateDispense(ctx context.Context, id, dispensedBy uuid.UUID, input *models.UpdateDispenseRequest) error
	FindDispenseByDrugID(ctx context.Context, drugID uuid.UUID) ([]*models.Dispense, error)
}

type DispenseSvc struct {
	repo   repositories.DispenseRepository
	logger *zap.Logger
}

var ErrDispenseNotFound = errors.New("dispense not found")
var ErrPrescriptionNotFound = errors.New("prescription not found")
var ErrFailedValidation = errors.New("validation failed")

func NewDispenseService(repo repositories.DispenseRepository, logger *zap.Logger) DispenseService {
	return &DispenseSvc{repo: repo, logger: logger}
}

func (s *DispenseSvc) CreateDispense(
	ctx context.Context,
	patientID,
	dispenseBy,
	prescriptionID,
	drugID uuid.UUID,
	quantity int32,
	notes string,
) (*models.Dispense, error) {

	if quantity <= 0 {
		s.logger.Warn("Invalid quantity in dispense request", zap.Int32("quantity", quantity))
		return nil, errors.New("quantity must be positive")
	}

	disp := &models.Dispense{
		ID:             uuid.New(),
		PatientID:      patientID,
		DispensedBy:    dispenseBy,
		PrescriptionID: prescriptionID,
		DrugID:         drugID,
		Quantity:       quantity,
		Notes:          notes,
	}

	dispense, err := s.repo.Create(ctx, disp)
	if err != nil {
		s.logger.Error("Failed to create dispense",
			zap.Error(err),
			zap.String("patient_id", patientID.String()),
			zap.String("drug_id", drugID.String()),
		)
		return nil, err
	}

	s.logger.Info("Dispense created successfully",
		zap.String("dispense_id", dispense.ID.String()),
		zap.String("patient_id", patientID.String()),
		zap.String("drug_id", drugID.String()),
		zap.Int32("quantity", quantity),
	)

	return dispense, nil
}

func (s *DispenseSvc) FindByID(ctx context.Context, id uuid.UUID) (*models.Dispense, error) {
	dispense, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Warn("Dispense not found", zap.String("dispense_id", id.String()))
		return nil, err
	}

	return dispense, nil
}

func (s *DispenseSvc) CreatePrescription(ctx context.Context, doctorID uuid.UUID, input *models.CreatePrescriptionRequest) error {
	prescription := &models.Prescription{
		ID:        uuid.New(),
		PatientID: uuid.MustParse(input.PatientID),
		DoctorID:  doctorID,
		DrugName:  input.DrugName,
		Dosage:    input.Dosage,
		Status:    "pending",
	}

	if err := s.repo.CreatePrescription(ctx, prescription); err != nil {
		s.logger.Error("Failed to create prescription", zap.Error(err))
		return err
	}

	s.logger.Info("Prescription created", zap.String("doctor_id", doctorID.String()))
	return nil
}

func (s *DispenseSvc) FindPrescriptionByID(ctx context.Context, id uuid.UUID) (*models.Prescription, error) {

	prescription, err := s.repo.FindPrescriptionByID(ctx, id)
	if err != nil {
		s.logger.Warn("Prescription not found", zap.String("prescription_id", id.String()))
		return nil, err
	}

	return prescription, nil
}

func (s *DispenseSvc) UpdateDispense(ctx context.Context, id, dispensedBy uuid.UUID, input *models.UpdateDispenseRequest) error {
	if err := models.Validate.Struct(input); err != nil {
		s.logger.Warn("Invalid update request",
			zap.String("dispense_id", id.String()),
			zap.String("dispensed_by", dispensedBy.String()),
		)
		return err
	}

	updates := make(map[string]any)
	if input.Quantity != nil {
		updates["quantity"] = *input.Quantity
	}

	if input.Notes != nil {
		updates["notes"] = *input.Notes
	}

	if len(updates) == 0 {
		return nil
	}

	if err := s.repo.Update(ctx, id, dispensedBy, updates); err != nil {
		s.logger.Error("Failed to update dispense", zap.Error(err), zap.String("dispense_id", id.String()))
		return err
	}

	s.logger.Info("Dispense updated succesfully", zap.String("dispense_id", id.String()))
	return nil

}

func (s *DispenseSvc) FindDispenseByDrugID(ctx context.Context, drugID uuid.UUID) ([]*models.Dispense, error) {
	dispense, err := s.repo.FindDispenseByDrugID(ctx, drugID)
	if err != nil {
		s.logger.Warn("Dispense not found by drug ID", zap.String("drug_id", drugID.String()))
		return nil, err
	}

	s.logger.Info("Dispense found by drug ID", zap.String("drug_id", drugID.String()))
	return dispense, nil
}
