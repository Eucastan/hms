package models

import (
	"github.com/google/uuid"
	"time"
)

type LabTestRequest struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PatientID uuid.UUID `gorm:"type:uuid;index;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	RequestBy uuid.UUID `gorm:"type:uuid;index"` // clinician ID
	TestType  string    `gorm:"size:100;index"`
	Priority  string    `gorm:"size:20"` // routine, urgent, stat
	Status    string    `gorm:"default:'requested'"`
	Notes     string    `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type LabResult struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	PerformedBy    uuid.UUID `gorm:"type:uuid;index"`
	PatientID      uuid.UUID `gorm:"type:uuid;index;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	TestType       string    `gorm:"size:100"`
	ResultValue    string    `gorm:"type:text"`
	Unit           string    `gorm:"size:50"`
	ReferenceRange string    `gorm:"size:100"`
	Comments       string    `gorm:"type:text"`
	Verified       bool      `gorm:"default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
