package repositories

import (
	"context"
	"errors"

	"github.com/Eucastan/hms/pharmacy/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DispenseRepository interface {
	Create(ctx context.Context, req *models.Dispense) (*models.Dispense, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Dispense, error)
	CreatePrescription(ctx context.Context, req *models.Prescription) error
	FindPrescriptionByID(ctx context.Context, id uuid.UUID) (*models.Prescription, error)
	Update(ctx context.Context, id, dispensedBy uuid.UUID, updates map[string]any) error
	FindDispenseByDrugID(ctx context.Context, drugID uuid.UUID) ([]*models.Dispense, error)
}

type DispenseRepo struct {
	DB *gorm.DB
}

func NewDispenseRepository(db *gorm.DB) DispenseRepository {
	return &DispenseRepo{DB: db}
}

func (r *DispenseRepo) Create(ctx context.Context, req *models.Dispense) (*models.Dispense, error) {

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var drug models.Drug
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&drug, req.DrugID).Error; err != nil {
			return err
		}

		if drug.Stock < req.Quantity {
			return errors.New("insuficient stock")
		}

		req.Total = float64(req.Quantity) * drug.PricePerUnit

		drug.Stock -= req.Quantity
		if err := tx.Save(&drug).Error; err != nil {
			return err
		}

		return tx.Create(req).Error
	})

	if err != nil {
		return nil, err
	}

	return req, nil

}

func (r *DispenseRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Dispense, error) {
	var dispense models.Dispense

	err := r.DB.WithContext(ctx).Preload("Drug").Where("id = ?", id).First(&dispense).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("dispense not found")
		}

		return nil, err
	}

	return &dispense, nil
}

func (r *DispenseRepo) CreatePrescription(ctx context.Context, req *models.Prescription) error {
	return r.DB.WithContext(ctx).Create(req).Error
}

func (r *DispenseRepo) FindPrescriptionByID(ctx context.Context, id uuid.UUID) (*models.Prescription, error) {
	var prescription models.Prescription
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&prescription).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("prescription not found")
		}
		return nil, err
	}
	return &prescription, nil
}

func (r *DispenseRepo) Update(ctx context.Context, id, dispensedBy uuid.UUID, updates map[string]any) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.Dispense
		if err := tx.Preload("Drug").First(&current, id).Error; err != nil {
			return err
		}

		quantityDelta := int32(0)
		if q, ok := updates["quantity"]; ok {
			newQty := q.(int32)
			quantityDelta = newQty - current.Quantity
		}

		if quantityDelta != 0 {
			if current.Drug.Stock < quantityDelta {
				return errors.New("insufficient stock for update")
			}
			current.Drug.Stock -= quantityDelta
			if err := tx.Save(&current.Drug).Error; err != nil {
				return err
			}

			// Recalculate total
			updates["total"] = float64(updates["quantity"].(int32)) * current.Drug.PricePerUnit
		}

		return tx.Model(&models.Dispense{}).
			Where("id = ? AND dispensedBy = ?", id, dispensedBy).
			Updates(updates).Error
	})
}

func (r *DispenseRepo) FindDispenseByDrugID(ctx context.Context, drugID uuid.UUID) ([]*models.Dispense, error) {
	var dispense []*models.Dispense
	if err := r.DB.WithContext(ctx).Preload("Drug").Where("drug_id = ?", drugID).Find(&dispense).Error; err != nil {
		return nil, err
	}

	return dispense, nil
}
