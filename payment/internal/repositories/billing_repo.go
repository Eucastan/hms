package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Eucastan/hms/payment/internal/models"
)

type BillingRepository interface {
	CreateBillCharge(ctx context.Context, charge *models.BillCharge) (*models.Invoice, float64, error)
	RefundBillCharge(ctx context.Context, chargeID, refundedBy uuid.UUID) error
	GetInvoiceByID(ctx context.Context, id uuid.UUID) (*models.Invoice, error)
}

type BillingRepo struct {
	DB *gorm.DB
}

func NewBillingRepository(db *gorm.DB) BillingRepository {
	return &BillingRepo{DB: db}
}

func (r *BillingRepo) CreateBillCharge(ctx context.Context, charge *models.BillCharge) (*models.Invoice, float64, error) {
	var invoice *models.Invoice
	var total float64

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get or create open invoice
		var inv models.Invoice
		err := tx.Where("patient_id = ? AND status != ?", charge.PatientID, models.InvoicePaid).
			FirstOrCreate(&inv, models.Invoice{
				ID:        uuid.New(),
				PatientID: charge.PatientID,
				Status:    models.InvoiceDraft,
			}).Error
		if err != nil {
			return err
		}
		invoice = &inv

		// Attach charge to invoice
		charge.InvoiceID = invoice.ID
		if err := tx.Create(charge).Error; err != nil {
			return err
		}

		// Recalculate invoice total
		var sum float64
		tx.Model(&models.BillCharge{}).
			Where("invoice_id = ? AND deleted_at IS NULL", invoice.ID).
			Select("COALESCE(SUM(total_amount), 0)").
			Scan(&sum)
		total = sum

		// Update invoice
		invoice.TotalAmount = total
		if err := tx.Model(invoice).Update("total_amount", total).Error; err != nil {
			return err
		}

		// Upsert payment record
		var payment models.Payment
		tx.Where("invoice_id = ?", invoice.ID).
			FirstOrCreate(&payment, models.Payment{
				ID:          uuid.New(),
				InvoiceID:   invoice.ID,
				TotalAmount: total,
				PaidAmount:  0,
				Balance:     total,
				Status:      "pending",
			})

		payment.TotalAmount = total
		payment.Balance = total - payment.PaidAmount
		return tx.Save(&payment).Error
	})

	return invoice, total, err
}

func (r *BillingRepo) RefundBillCharge(ctx context.Context, chargeID, refundedBy uuid.UUID) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var charge models.BillCharge
		if err := tx.First(&charge, chargeID).Error; err != nil {
			return err
		}

		// Ownership check
		if charge.CreatedBy != refundedBy {
			return errors.New("you can only refund charges you created")
		}

		// Mark original as refunded
		charge.Status = models.ChargeRefunded
		if err := tx.Save(&charge).Error; err != nil {
			return err
		}

		// Create reversing (negative) charge
		refund := models.BillCharge{
			ID:            uuid.New(),
			PatientID:     charge.PatientID,
			InvoiceID:     charge.InvoiceID,
			SourceRefID:   charge.SourceRefID,
			ReferenceType: charge.ReferenceType,
			Description:   "REFUND: " + charge.Description,
			Quantity:      -charge.Quantity,
			UnitPrice:     charge.UnitPrice,
			TotalAmount:   -charge.TotalAmount,
			Status:        models.ChargeRefunded,
			CreatedBy:     refundedBy,
		}
		if err := tx.Create(&refund).Error; err != nil {
			return err
		}

		// Recalculate invoice total to reflect changes due to refund
		var newTotal float64
		tx.Model(&models.BillCharge{}).
			Where("invoice_id = ? AND deleted_at IS NULL", charge.InvoiceID).
			Select("COALESCE(SUM(total_amount), 0)").
			Scan(&newTotal)

		// Update invoice
		if err := tx.Model(&models.Invoice{}).
			Where("id = ?", charge.InvoiceID).
			Update("total_amount", newTotal).Error; err != nil {
			return err
		}

		// Update payment balance
		var payment models.Payment
		if err := tx.Where("invoice_id = ?", charge.InvoiceID).First(&payment).Error; err == nil {
			payment.Balance = newTotal - payment.PaidAmount
			return tx.Save(&payment).Error
		}

		return nil
	})
}

func (r *BillingRepo) GetInvoiceByID(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.DB.WithContext(ctx).Preload("Item").Where("id = ?", id).First(&invoice).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("not found")
		}

		return nil, err
	}

	return &invoice, nil
}
