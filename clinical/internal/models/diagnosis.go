package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Diagnosis struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PatientID   uuid.UUID `gorm:"type:uuid;index;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	AdmissionID uuid.UUID `gorm:"type:uuid;index;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	DoctorID    uuid.UUID `gorm:"type:uuid;index;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"` // from Auth
	Code        string    `gorm:"size:50"`                                                        // ICD-10 code
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
