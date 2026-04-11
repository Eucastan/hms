package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChargeStatus string

const (
	ChargePending  ChargeStatus = "pending"
	ChargeInvoiced ChargeStatus = "invoiced"
	ChargeRefunded ChargeStatus = "refunded"
)

type BillCharge struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PatientID     uuid.UUID `gorm:"type:uuid;index"`
	InvoiceID     uuid.UUID `gorm:"type:uuid;index"`
	SourceRefID   uuid.UUID `gorm:"type:uuid;uniqueIndex"`
	ReferenceType string    `gorm:"size:50"` // drug, lab, consultation, ward
	Description   string
	Quantity      int32
	UnitPrice     float64
	TotalAmount   float64
	Status        ChargeStatus `gorm:"default:'pending'"`
	CreatedBy     uuid.UUID    `gorm:"type:uuid;index"` // ownership!
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"` // soft delete
}
