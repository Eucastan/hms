package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Eucastan/hms/pharmacy/internal/models"
)

type DrugRepository interface {
	Create(ctx context.Context, drug *models.Drug) (*models.Drug, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Drug, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]any) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type DrugRepo struct {
	DB *gorm.DB
}

func NewDrugRepository(db *gorm.DB) DrugRepository {
	return &DrugRepo{DB: db}
}

func (r *DrugRepo) Create(ctx context.Context, drug *models.Drug) (*models.Drug, error) {

	if err := r.DB.WithContext(ctx).Create(drug).Error; err != nil {
		return nil, err
	}

	return drug, nil
}

func (r *DrugRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Drug, error) {
	var drug models.Drug
	err := r.DB.WithContext(ctx).Preload("Dispense").Where("id = ?", id).First(&drug).Error
	if err != nil {
		return nil, err
	}

	return &drug, nil
}

func (r *DrugRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]any) error {

	delete(updates, "created_at")

	return r.DB.WithContext(ctx).
		Model(&models.Drug{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *DrugRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.DB.WithContext(ctx).Where("id = ?", id).Delete(&models.Drug{}).Error
}
