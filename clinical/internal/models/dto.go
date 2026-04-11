package models

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())

	Validate.RegisterValidation("uuid4", func(fl validator.FieldLevel) bool {
		_, err := uuid.Parse(fl.Field().String())
		return err == nil
	})
}

type CreatePrescriptionRequest struct {
	PatientID string `json:"patient_id" binding:"required"`
	DoctorID  string `json:"doctor_id" binding:"required"`
	DrugName  string `json:"drug_name" binding:"required"`
	Dosage    string `json:"dosage" binding:"required"`
}

type LabRequest struct {
	PatientID string `json:"patient_id" binding:"required"`
	RequestBy string `json:"request_by" binding:"required"`
	TestType  string `json:"test_type" binding:"required"`
	Priority  string `json:"priority" binding:"required"`
	Notes     string `json:"notes" binding:"required"`
}

type LabRequestUpdate struct {
	TestType *string `json:"test_type" validate:"omitempty,max=100"`
	Priority *string `json:"priority"  validate:"omitempty,oneof=routine urgent stat"`
	Notes    *string `json:"notes"     validate:"omitempty,max=2000"`
}

type Patient struct {
	HospitalNo    string `json:"hospital_no" binding:"required"` // file/hospital number
	FirstName     string `json:"first_name" binding:"required"`
	LastName      string `json:"last_name" binding:"required"`
	Gender        string `json:"gender" binding:"required"`
	Address       string `json:"address" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	NextOfKinName string `json:"next_of_kin_name" binding:"required"`
}
