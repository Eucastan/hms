package models

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())

	// UUID validator
	Validate.RegisterValidation("uuid4", func(fl validator.FieldLevel) bool {
		_, err := uuid.Parse(fl.Field().String())
		return err == nil
	})

	// Role enum validator
	Validate.RegisterValidation("role", func(fl validator.FieldLevel) bool {
		role := fl.Field().String()
		switch role {
		case string(Doctor), string(Admin), string(Nurse), string(Accountant):
			return true
		default:
			return false
		}
	})
}

type StaffCreateRequest struct {
	Email     string `json:"email"     validate:"required,email,max=255"`
	Password  string `json:"password"  validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name"  validate:"required,max=100"`
	Role      string `json:"role"       validate:"required,role"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8"`
}

type StaffUpdateRequest struct {
	FirstName *string `json:"first_name" validate:"omitempty,max=100"`
	LastName  *string `json:"last_name"  validate:"omitempty,max=100"`
	Role      *string `json:"role"       validate:"omitempty,role"`
	Active    *bool   `json:"active"     validate:"omitempty"`
	Password  *string `json:"password"   validate:"omitempty,min=8"` // optional password change
}

type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type PasswordResetConfirmRequest struct {
	Token       string `json:"token" validate:"required,max=255"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}
