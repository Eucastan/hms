package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Patient struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	HospitalNo    string         `gorm:"uniqueIndex;size:50" json:"hospital_no"` // file/hospital number
	FirstName     string         `gorm:"size:100;not null" json:"first_name"`
	LastName      string         `gorm:"size:100;not null" json:"last_name"`
	DateOfBirth   string         `gorm:"index" json:"date_of_birth"`
	Age           int32          `json:"age"` // computed or stored
	Gender        string         `gorm:"type:varchar(20)" json:"gender" `
	Address       string         `gorm:"size:500" json:"address"`
	Phone         string         `gorm:"size:30" json:"phone"`
	NextOfKinName string         `gorm:"size:100" json:"next_of_kin_name"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Admission []Admission `gorm:"foreignKey:PatientID" json:"admissions"`
}

type Admission struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	PatientID  uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Patient    Patient   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	AdmittedAt time.Time `gorm:"index"`
	Ward       string    `gorm:"size:50"`
	BedNumber  string    `gorm:"size:20"`
	Status     string    `gorm:"size:30;default:'active'"` // active, transferred, discharged
	Reason     string    `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
