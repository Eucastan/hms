package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Eucastan/hms/payment/internal/models"
	"github.com/Eucastan/hms/payment/internal/services"
)

type BillingHandler struct {
	svc services.BillingService
}

func NewBillingHandler(svc services.BillingService) *BillingHandler {
	return &BillingHandler{svc: svc}
}

func (h *BillingHandler) CreateBillCharge(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	createdBy, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input models.BillChargeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json", "details": err.Error()})
		return
	}

	input.CreatedBy = createdBy.String()
	input.ReferenceType = strings.ToLower(input.ReferenceType)

	charge, invoice, total, err := h.svc.CreateBillCharge(ctx, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Bill charge created",
		"charge_id":  charge.ID,
		"invoice_id": invoice.ID,
		"new_total":  total,
	})
}

func (h *BillingHandler) RefundBillCharge(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	refundedBy, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chargeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid charge id"})
		return
	}

	err = h.svc.RefundBillCharge(ctx, chargeID, refundedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Charge refunded successfully"})
}

func (s *BillingHandler) GetInvoice(c *gin.Context) {
	ctx := c.Request.Context()

	InvoiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	Invoice, err := s.svc.GetInvoiceByID(ctx, InvoiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, Invoice)
}

func (h *BillingHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "Healthy",
		"service":   "Billing Service",
		"version":   "v1.0.0",
		"uptime":    time.Since(time.Now().Add(-time.Hour)).String(),
		"timestamp": time.Now().UTC(),
		"message":   "Billing Service is running smoothly",
	})
}

func (h *BillingHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "Alive",
		"service":   "Billing Service",
		"version":   "v1.0.0",
		"uptime":    time.Since(time.Now().Add(-time.Hour)).String(),
		"timestamp": time.Now().UTC(),
		"message":   "Billing Service is running smoothly",
	})
}

func (h *BillingHandler) Readiness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "Ready",
		"service":   "Billing Service",
		"version":   "v1.0.0",
		"uptime":    time.Since(time.Now().Add(-time.Hour)).String(),
		"timestamp": time.Now().UTC(),
		"message":   "Billing Service is running smoothly",
	})
}
