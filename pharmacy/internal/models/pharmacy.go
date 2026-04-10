package models

import (
	"github.com/google/uuid"
	"time"
)

type Drug struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string    `gorm:"size:200;uniqueIndex"`
	GenericName  string    `gorm:"size:200"`
	Form         string    `gorm:"size:50"` // tablet, syrup, injection
	Strength     string    `gorm:"size:50"`
	PackSize     int32
	Stock        int32 `gorm:"default:0"`
	MinStock     int32 `gorm:"default:10"`
	PricePerUnit float64
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Dispense []Dispense `gorm:"foreignKey:DrugID"`
}

type Dispense struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PatientID      uuid.UUID `gorm:"type:uuid;index;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	PrescriptionID uuid.UUID `gorm:"type:uuid;index,omitempty"` // link if from prescription
	DrugID         uuid.UUID `gorm:"type:uuid;index"`
	Drug           Drug      `gorm:"constrait:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Quantity       int32     `gorm:"not null;default:1"`
	DispensedBy    uuid.UUID `gorm:"type:uuid;index"` // pharmacist ID
	Notes          string    `gorm:"type:text"`
	Total          float64   `gorm:"default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Prescription struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PatientID uuid.UUID `gorm:"type:uuid;index;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	DoctorID  uuid.UUID `gorm:"type:uuid;index;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	DrugName  string    `gorm:"not null;size:200"`
	Dosage    string    `gorm:"not null;size:100"`
	Status    string    `gorm:"not null;default:'pending'"` // pending, sent_to_pharmacy, dispensed
}
