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

type BillChargeRequest struct {
	PatientID     string  `json:"patient_id"      validate:"required,uuid4"`
	SourceRefID   string  `json:"source_ref_id"   validate:"required,uuid4"`
	ReferenceType string  `json:"reference_type"  validate:"required,oneof=drug lab consultation ward"`
	Description   string  `json:"description"     validate:"required,max=500"`
	Quantity      int32   `json:"quantity"        validate:"required,gt=0"`
	UnitPrice     float64 `json:"unit_price"      validate:"required,gt=0"`
	CreatedBy     string  `json:"created_by"      validate:"required"`
}

type PaymentRequest struct {
	InvoiceID     string  `json:"invoice_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required"` // cash, card, insurance
}
