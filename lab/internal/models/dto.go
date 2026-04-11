package models

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
	// Optional: register custom tag for uuid4
	Validate.RegisterValidation("uuid4", func(fl validator.FieldLevel) bool {
		_, err := uuid.Parse(fl.Field().String())
		return err == nil
	})
}

type LabRequestCreate struct {
	PatientID string `json:"patient_id" validate:"required,uuid4"`
	TestType  string `json:"test_type"  validate:"required,max=100"`
	Priority  string `json:"priority"   validate:"required,oneof=routine urgent stat"`
	Notes     string `json:"notes"      validate:"omitempty,max=2000"`
}

type LabRequestUpdate struct {
	TestType *string `json:"test_type" validate:"omitempty,max=100"`
	Priority *string `json:"priority"  validate:"omitempty,oneof=routine urgent stat"`
	Notes    *string `json:"notes"     validate:"omitempty,max=2000"`
}

type LabResultCreate struct {
	PatientID      string `json:"patient_id"       validate:"required,uuid4"`
	TestType       string `json:"test_type"        validate:"required,max=100"`
	ResultValue    string `json:"result_value"     validate:"required,max=500"`
	Unit           string `json:"unit"             validate:"omitempty,max=50"`
	ReferenceRange string `json:"reference_range"  validate:"omitempty,max=100"`
	Comments       string `json:"comments"         validate:"omitempty,max=2000"`
	Verified       bool   `json:"verified"`
}

type LabResultUpdate struct {
	ResultValue    *string `json:"result_value"    validate:"omitempty,max=500"`
	Unit           *string `json:"unit"            validate:"omitempty,max=50"`
	ReferenceRange *string `json:"reference_range" validate:"omitempty,max=100"`
	Comments       *string `json:"comments"        validate:"omitempty,max=2000"`
	Verified       *bool   `json:"verified"`
}
