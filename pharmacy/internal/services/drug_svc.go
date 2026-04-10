package services

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Eucastan/hms/pharmacy/internal/models"
	"github.com/Eucastan/hms/pharmacy/internal/repositories"
)

type DrugService interface {
	Create(
		ctx context.Context,
		name,
		genericName,
		form,
		strength string,
		packSize int32,
		stock int32,
		price float64,
	) (*models.Drug, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Drug, error)
	Update(ctx context.Context, id uuid.UUID, inputs *models.DrugUpdateRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type DrugSvc struct {
	repo   repositories.DrugRepository
	logger *zap.Logger
}

func NewDrugService(repo repositories.DrugRepository, logger *zap.Logger) DrugService {
	return &DrugSvc{repo: repo, logger: logger}
}

func (s *DrugSvc) Create(
	ctx context.Context,
	name,
	genericName,
	form,
	strength string,
	packSize int32,
	stock int32,
	price float64,
) (*models.Drug, error) {

	d := &models.Drug{
		ID:           uuid.New(),
		Name:         name,
		GenericName:  genericName,
		Form:         form,
		Strength:     strength,
		PackSize:     packSize,
		Stock:        stock,
		PricePerUnit: price,
	}

	drug, err := s.repo.Create(ctx, d)
	if err != nil {
		s.logger.Error("Failed to create drug")
		return nil, err
	}

	s.logger.Info("Drug created successfully", zap.String("drug_id", drug.ID.String()))
	return drug, nil
}

func (s *DrugSvc) FindByID(ctx context.Context, id uuid.UUID) (*models.Drug, error) {

	drug, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Warn("Drug not found", zap.String("drug_id", id.String()))
		return nil, err
	}

	return drug, nil
}

func (s *DrugSvc) Update(ctx context.Context, id uuid.UUID, input *models.DrugUpdateRequest) error {
	if err := models.Validate.Struct(input); err != nil {
		s.logger.Warn("Invalid update request", zap.String("drug_id", id.String()))
		return err
	}

	updates := map[string]any{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.GenericName != nil {
		updates["generic_name"] = *input.GenericName
	}
	if input.Form != nil {
		updates["form"] = *input.Form
	}
	if input.Strength != nil {
		updates["strength"] = *input.Strength
	}
	if input.PackSize != nil {
		updates["pack_size"] = *input.PackSize
	}
	if input.Stock != nil {
		updates["stock"] = *input.Stock
	}
	if input.PricePerUnit != nil {
		updates["price_per_unit"] = *input.PricePerUnit
	}
	if input.MinStock != nil {
		updates["min_stock"] = *input.MinStock
	}

	if len(updates) == 0 {
		return nil
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		s.logger.Error("Failed to update drug", zap.Error(err), zap.String("drug_id", id.String()))
		return err
	}

	s.logger.Info("Drug updated successfully", zap.String("drug_id", id.String()))
	return nil

}

func (s *DrugSvc) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete drug", zap.Error(err), zap.String("drug_id", id.String()))
		return err
	}

	s.logger.Info("Drug deleted successfully", zap.String("drug_id", id.String()))
	return nil

}
