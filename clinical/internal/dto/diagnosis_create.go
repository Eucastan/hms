package dto

import (
	"github.com/google/uuid"
)

type DiagnosisCreateRequest struct {
	PatientID   uuid.UUID `json:"patient_id"   validate:"required,uuid"`
	AdmissionID uuid.UUID `json:"admission_id" validate:"omitempty,uuid"`
	Code        string    `json:"code"         validate:"required,code,max=50"`
	Description string    `json:"description"  validate:"required,min=5,max=4000"`
}
