package models

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())

	// Custom uuid validator
	Validate.RegisterValidation("uuid4", func(fl validator.FieldLevel) bool {
		_, err := uuid.Parse(fl.Field().String())
		return err == nil
	})

	// custom enum validators
	Validate.RegisterValidation("drug_form", func(fl validator.FieldLevel) bool {
		v := fl.Field().String()
		return v == "tablet" || v == "syrup" || v == "injection" || v == "capsule" // etc.
	})
}

type CreatePrescriptionRequest struct {
	PatientID string `json:"patient_id" validate:"required,uuid4"`
	DrugName  string `json:"drug_name"  validate:"required,max=200"`
	Dosage    string `json:"dosage"     validate:"required,max=100"`
}

type CreateDispenseRequest struct {
	PatientID      string `json:"patient_id"       validate:"required,uuid4"`
	PrescriptionID string `json:"prescription_id"  validate:"omitempty,uuid4"`
	DrugID         string `json:"drug_id"          validate:"required,uuid4"`
	Quantity       int32  `json:"quantity"         validate:"required,gt=0"`
	Notes          string `json:"notes"            validate:"omitempty,max=1000"`
}

type UpdateDispenseRequest struct {
	Quantity *int32  `json:"quantity" validate:"omitempty,gt=0"`
	Notes    *string `json:"notes"    validate:"omitempty,max=1000"`
}

type DrugCreateRequest struct {
	Name         string  `json:"name"           validate:"required,max=200"`
	GenericName  string  `json:"generic_name"   validate:"required,max=200"`
	Form         string  `json:"form"           validate:"required,drug_form"`
	Strength     string  `json:"strength"       validate:"required,max=50"`
	PackSize     int32   `json:"pack_size"      validate:"required,gt=0"`
	Stock        int32   `json:"stock"          validate:"required,gte=0"`
	PricePerUnit float64 `json:"price_per_unit" validate:"required,gt=0"`
	MinStock     int32   `json:"min_stock"      validate:"omitempty,gte=0"`
}

type DrugUpdateRequest struct {
	Name         *string  `json:"name"          validate:"omitempty,max=200"`
	GenericName  *string  `json:"generic_name"  validate:"omitempty,max=200"`
	Form         *string  `json:"form"          validate:"omitempty,drug_form"`
	Strength     *string  `json:"strength"      validate:"omitempty,max=50"`
	PackSize     *int32   `json:"pack_size"     validate:"omitempty,gt=0"`
	Stock        *int32   `json:"stock"         validate:"omitempty,gte=0"`
	PricePerUnit *float64 `json:"price_per_unit" validate:"omitempty,gt=0"`
	MinStock     *int32   `json:"min_stock"     validate:"omitempty,gte=0"`
}

type BillChargeReq struct {
	PatientID   string  `json:"patient_id" binding:"required" validate:"required,max=100"`
	SourceRefID string  `json:"source_ref_id" binding:"required" validate:"required,max=100"`
	Description string  `json:"description" binding:"required" validate:"required,max=100"`
	Quantity    int32   `json:"quantity" binding:"required" validate:"required,max=100"`
	UnitPrice   float64 `json:"unit_price" binding:"required" validate:"required,max=100"`
}
