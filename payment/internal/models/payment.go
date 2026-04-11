package models

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceDraft     InvoiceStatus = "draft"
	InvoicePartially InvoiceStatus = "partially"
	InvoicePaid      InvoiceStatus = "paid"
	InvoiceCancelled InvoiceStatus = "cancelled"
)

type Invoice struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey"`
	PatientID   uuid.UUID     `gorm:"type:uuid;index;not null"`
	TotalAmount float64       `gorm:"type:decimal(15,2);not null"`
	PaidAmount  float64       `gorm:"type:decimal(15,2);not null"`
	Status      InvoiceStatus // UNPAID | PARTIAL | PAID
	CreatedAt   time.Time
	Item        []Payment `gorm:"foreignKey:InvoiceID"`
}

type Payment struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	InvoiceID   uuid.UUID `gorm:"type:uuid;index;not null"`
	TotalAmount float64   `gorm:"type:decimal(15,2);not null"`
	PaidAmount  float64   `gorm:"type:decimal(15,2);not null"`
	Balance     float64   `gorm:"type:decimal(15,2);not null"`
	Status      string    `gorm:"size:50;default:'pending'"`
	CreatedAt   time.Time
}
