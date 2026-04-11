// internal/services/billing_service.go
package services

import (
	"context"
	"fmt"

	"github.com/Eucastan/hms/payment/internal/models"
	"github.com/Eucastan/hms/payment/internal/repositories"
	"go.uber.org/zap"

	"github.com/google/uuid"
)

type BillingService interface {
	CreateBillCharge(ctx context.Context, input models.BillChargeRequest) (*models.BillCharge, *models.Invoice, float64, error)
	RefundBillCharge(ctx context.Context, chargeID uuid.UUID, refundedBy uuid.UUID) error
	GetInvoiceByID(ctx context.Context, id uuid.UUID) (*models.Invoice, error)
}

type BillingSvc struct {
	repo   repositories.BillingRepository
	logger *zap.Logger
}

func NewBillingService(repo repositories.BillingRepository, logger *zap.Logger) BillingService {
	return &BillingSvc{repo: repo, logger: logger}
}

func (s *BillingSvc) CreateBillCharge(ctx context.Context, input models.BillChargeRequest) (*models.BillCharge, *models.Invoice, float64, error) {
	if err := models.Validate.Struct(input); err != nil {
		s.logger.Warn("Bill charge validation failed", zap.Error(err))
		return nil, nil, 0, fmt.Errorf("validation failed: %w", err)
	}

	createdBy, err := uuid.Parse(input.CreatedBy)
	if err != nil {
		s.logger.Warn("Invalid createdBy format", zap.Error(err))
		return nil, nil, 0, err
	}

	sourceRefID, err := uuid.Parse(input.SourceRefID)
	if err != nil {
		s.logger.Warn("Invalid sourceRefID format", zap.Error(err))
		return nil, nil, 0, err
	}

	patientID, err := uuid.Parse(input.PatientID)
	if err != nil {
		s.logger.Warn("Invalid patientID format", zap.Error(err))
		return nil, nil, 0, err
	}

	charge := &models.BillCharge{
		ID:            uuid.New(),
		PatientID:     patientID,
		SourceRefID:   sourceRefID,
		ReferenceType: input.ReferenceType,
		Description:   input.Description,
		Quantity:      int32(input.Quantity),
		UnitPrice:     input.UnitPrice,
		TotalAmount:   float64(input.Quantity) * input.UnitPrice,
		CreatedBy:     createdBy,
	}

	invoice, total, err := s.repo.CreateBillCharge(ctx, charge)
	if err != nil {
		s.logger.Error("Failed to create bill charge in repository",
			zap.Error(err),
			zap.String("patient_id", patientID.String()),
		)

		return nil, nil, 0, err
	}

	s.logger.Info("Bill charge created successfully",
		zap.String("charge_id", charge.ID.String()),
		zap.String("invoice_id", invoice.ID.String()),
		zap.Float64("total", total),
	)

	return charge, invoice, total, nil
}

func (s *BillingSvc) RefundBillCharge(ctx context.Context, chargeID uuid.UUID, refundedBy uuid.UUID) error {
	if err := s.repo.RefundBillCharge(ctx, chargeID, refundedBy); err != nil {
		s.logger.Error("Failed to refund bill charge",
			zap.Error(err),
			zap.String("charge_id", chargeID.String()),
			zap.String("refunded_by", refundedBy.String()),
		)
		return err
	}

	s.logger.Info("Charge refunded successfully",
		zap.String("charge_id", chargeID.String()),
		zap.String("refunded_by", refundedBy.String()),
	)

	return nil
}

func (s *BillingSvc) GetInvoiceByID(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	invoice, err := s.repo.GetInvoiceByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get invoice by ID",
			zap.Error(err),
			zap.String("invoice_id", id.String()),
		)
		return nil, err
	}

	return invoice, nil
}
