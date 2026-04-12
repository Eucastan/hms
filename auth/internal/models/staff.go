package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	Doctor     Role = "doctor"
	Admin      Role = "admin"
	Nurse      Role = "nurse"
	Accountant Role = "accountant"
)

type Staff struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"uniqueIndex;size:255;not null"`
	Password  string    `gorm:"size:255;not null"`
	FirstName string    `gorm:"size:100"`
	LastName  string    `gorm:"size:100"`
	Role      Role      `gorm:"type:varchar(50);not null;index;default:'admin'"`
	Active    bool      `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
