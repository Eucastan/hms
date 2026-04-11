package dto

import (
	"time"
)

type DiagnosisResponse struct {
	ID          string    `json:"id"`
	PatientID   string    `json:"patient_id"`
	AdmissionID *string   `json:"admission_id,omitempty"`
	DoctorID    string    `json:"doctor_id"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
