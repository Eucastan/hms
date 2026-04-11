package services

import (
	"context"
	"errors"
	"time"

	"github.com/Eucastan/hms/patient/internal/models"
	"github.com/Eucastan/hms/patient/internal/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PatientService interface {
	CreatePatient(ctx context.Context, input models.PatientCreate) (*models.Patient, error)
	SearchPatient(ctx context.Context, name, hospitalNo string) ([]*models.Patient, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Patient, error)
	FindAdmissionByID(ctx context.Context, id uuid.UUID) (*models.Admission, error)
	UpdatePatient(ctx context.Context, id uuid.UUID, input models.PatientUpdate) error
	UpdateAdmission(ctx context.Context, id uuid.UUID, input *models.AdmissionUpdate) error
	DeletePatient(ctx context.Context, id uuid.UUID) error
}

type PatientSvc struct {
	ptRepo   repositories.PatientRepository
	admsRepo repositories.AdmissionRepository
	logger   *zap.Logger
}

var ErrNotFound = errors.New("record not found")
var ErrOnAdmissionCreation = errors.New("failed to create initial admission")
var ErrFailedValidation = errors.New("validation failed")

func NewPatientService(ptRepo repositories.PatientRepository, admsn repositories.AdmissionRepository, logger *zap.Logger) PatientService {
	return &PatientSvc{
		ptRepo:   ptRepo,
		admsRepo: admsn,
		logger:   logger,
	}
}

func (s *PatientSvc) createAdmission(ctx context.Context, patientID uuid.UUID, input models.PatientCreate) error {
	if !input.AutoCreateAdmission {
		return nil // skip silently
	}

	admission := &models.Admission{
		ID:         uuid.New(),
		PatientID:  patientID,
		AdmittedAt: time.Now(),
		Ward:       input.InitialWard,
		BedNumber:  "",
		Status:     "admitted",
		Reason:     input.InitialReason,
	}

	if err := s.admsRepo.Create(ctx, admission); err != nil {
		s.logger.Error(
			"WARNING: auto-admission failed for patient",
			zap.String("patient", patientID.String()),
			zap.String("err", err.Error()),
		)
		return ErrOnAdmissionCreation
	}

	return nil
}

func (s *PatientSvc) CreatePatient(ctx context.Context, input models.PatientCreate) (*models.Patient, error) {
	if err := models.Validate.Struct(input); err != nil {
		s.logger.Error(
			"Failed to validate",
			zap.String("err", err.Error()),
		)
		return nil, ErrFailedValidation
	}

	patient := &models.Patient{
		ID:            uuid.New(),
		HospitalNo:    input.HospitalNo,
		FirstName:     input.FirstName,
		LastName:      input.LastName,
		DateOfBirth:   input.DateOfBirth,
		Age:           input.Age,
		Gender:        input.Gender,
		Address:       input.Address,
		Phone:         input.Phone,
		NextOfKinName: input.NextOfKinName,
	}

	created, err := s.ptRepo.Create(ctx, patient)
	if err != nil {
		s.logger.Error("Failed to create patient", zap.Error(err))
		return nil, err
	}

	if err := s.createAdmission(ctx, created.ID, input); err != nil {
		return nil, err
	}

	// Reload to include admission if created
	fullPatient, err := s.ptRepo.FindByID(ctx, created.ID)
	if err != nil {
		return created, nil // fallback to basic patient
	}

	return fullPatient, nil
}

func (s *PatientSvc) SearchPatient(ctx context.Context, name, hospitalNo string) ([]*models.Patient, error) {

	if name == "" && hospitalNo == "" {
		return nil, errors.New("at least one of name or hospital_no is required")
	}

	return s.ptRepo.FindByNameAndHospitalNo(ctx, name, hospitalNo)

}

func (s *PatientSvc) FindByID(ctx context.Context, id uuid.UUID) (*models.Patient, error) {

	patients, err := s.ptRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to fetch patient", zap.Error(err))
		return nil, err
	}

	return patients, nil
}

func (s *PatientSvc) FindAdmissionByID(ctx context.Context, id uuid.UUID) (*models.Admission, error) {
	admsn, err := s.admsRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to fetch admission", zap.Error(err))
		return nil, err
	}

	return admsn, nil
}

func (s *PatientSvc) UpdatePatient(ctx context.Context, id uuid.UUID, input models.PatientUpdate) error {
	if err := models.Validate.Struct(input); err != nil {
		s.logger.Error(
			"failed validate",
			zap.String("err", err.Error()),
		)
		return ErrFailedValidation
	}

	updates := make(map[string]any)

	if input.FirstName != nil {
		updates["first_name"] = *input.FirstName
	}
	if input.LastName != nil {
		updates["last_name"] = *input.LastName
	}
	if input.DateOfBirth != nil {
		updates["date_of_birth"] = *input.DateOfBirth
	}
	if input.Age != nil {
		updates["age"] = *input.Age
	}
	if input.Gender != nil {
		updates["gender"] = *input.Gender
	}
	if input.Address != nil {
		updates["address"] = *input.Address
	}
	if input.Phone != nil {
		updates["phone"] = *input.Phone
	}
	if input.NextOfKinName != nil {
		updates["next_of_kin_name"] = *input.NextOfKinName
	}

	if err := s.ptRepo.Update(ctx, id, updates); err != nil {
		s.logger.Error("Failed to update patient", zap.Error(err))
		return err
	}

	s.logger.Info("Patient updated successfully", zap.String("patient_id", id.String()))
	return nil
}

func (s *PatientSvc) UpdateAdmission(ctx context.Context, id uuid.UUID, input *models.AdmissionUpdate) error {
	if err := models.Validate.Struct(input); err != nil {
		s.logger.Error(
			"failed validate",
			zap.String("err", err.Error()),
		)
		return ErrFailedValidation
	}

	updates := make(map[string]any)

	if input.Ward != nil {
		updates["ward"] = *input.Ward
	}

	if input.BedNumber != nil {
		updates["bed_number"] = *input.BedNumber
	}

	if input.Status != nil {
		updates["status"] = *input.Status
	}

	if input.Reason != nil {
		updates["reason"] = *input.Reason
	}

	if err := s.admsRepo.Update(ctx, id, updates); err != nil {
		s.logger.Error("Failed to update admission", zap.Error(err))
		return err
	}

	s.logger.Info("Admission updated successfully", zap.String("admission_id", id.String()))
	return nil
}

func (s *PatientSvc) DeletePatient(ctx context.Context, id uuid.UUID) error {
	if err := s.ptRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete patient", zap.Error(err))
		return err
	}

	s.logger.Info("Patient deleted successfully", zap.String("patient_id", id.String()))
	return nil
}
