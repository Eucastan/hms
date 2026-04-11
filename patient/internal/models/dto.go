package models

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

var Validate *validator.Validate
var hospitalNoRegex = regexp.MustCompile(`^[A-Z0-9-]+$`)

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())

	Validate.RegisterValidation("hospital_no", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return hospitalNoRegex.MatchString(value)
	})
}

type PatientCreate struct {
	HospitalNo    string `json:"hospital_no"    validate:"required,hospital_no,max=50"`
	FirstName     string `json:"first_name"     validate:"required,max=100"`
	LastName      string `json:"last_name"      validate:"required,max=100"`
	DateOfBirth   string `json:"date_of_birth"            validate:"required,datetime=2006-01-02"` // ISO date
	Age           int32  `json:"age"            validate:"gte=0,lte=150"`
	Gender        string `json:"gender"         validate:"required,oneof=M F O"`
	Address       string `json:"address"        validate:"max=500"`
	Phone         string `json:"phone"          validate:"max=30"` // can add e164 later
	NextOfKinName string `json:"next_of_kin_name" validate:"max=100"`

	AutoCreateAdmission bool   `json:"auto_create_admission"` // default false
	InitialWard         string `json:"initial_ward" validate:"omitempty,max=50"`
	InitialReason       string `json:"initial_reason" validate:"omitempty,max=1000"`
}

type PatientUpdate struct {
	HospitalNo    *string `json:"hospital_no"    validate:"omitempty,alphanumunicode,max=50"`
	FirstName     *string `json:"first_name"     validate:"omitempty,max=100"`
	LastName      *string `json:"last_name"      validate:"omitempty,max=100"`
	DateOfBirth   *string `json:"dob"            validate:"omitempty,datetime=2006-01-02"`
	Age           *int32  `json:"age"            validate:"omitempty,gte=0,lte=150"`
	Gender        *string `json:"gender"         validate:"omitempty,oneof=M F O"`
	Address       *string `json:"address"        validate:"omitempty,max=500"`
	Phone         *string `json:"phone"          validate:"omitempty,max=30"`
	NextOfKinName *string `json:"next_of_kin_name" validate:"omitempty,max=100"`
}

type AdmissionUpdate struct {
	Ward      *string `json:"ward"       validate:"omitempty,max=50"`
	BedNumber *string `json:"bed_number" validate:"omitempty,max=20"`
	Status    *string `json:"status"     validate:"omitempty,oneof=active transferred discharged"`
	Reason    *string `json:"reason"     validate:"omitempty,max=2000"`
}
