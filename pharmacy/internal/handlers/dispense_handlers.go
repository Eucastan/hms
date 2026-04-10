package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Eucastan/hms/pharmacy/internal/grpc/client"
	"github.com/Eucastan/hms/pharmacy/internal/models"
	"github.com/Eucastan/hms/pharmacy/internal/services"
)

type DispenseHandler struct {
	svc     services.DispenseService
	clients *client.Clients
}

func NewDispenseHandler(svc services.DispenseService, c *client.Clients) *DispenseHandler {
	return &DispenseHandler{svc: svc, clients: c}
}

func (h *DispenseHandler) CreateDispense(c *gin.Context) {
	ctx := c.Request.Context()

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user in context"})
		return
	}

	dispenseBy, err := uuid.Parse(userId.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id format: " + err.Error()})
		return
	}

	var req models.CreateDispenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json request: " + err.Error()})
		return
	}

	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string: " + err.Error()})
		return
	}

	drugID, err := uuid.Parse(req.DrugID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string: " + err.Error()})
		return
	}

	var prescriptionUUID uuid.UUID
	if req.PrescriptionID != "" {
		prescriptionUUID = uuid.MustParse(req.PrescriptionID)
	}

	if err := models.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalidation failed",
			"details": err.Error(),
		})
		return
	}

	if err := h.clients.ValidatePatientID(ctx, req.PatientID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Patient not found: " + err.Error()})
		return
	}

	dispense, err := h.svc.CreateDispense(
		ctx,
		patientID,
		dispenseBy,
		prescriptionUUID,
		drugID,
		req.Quantity,
		req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create: " + err.Error()})
		return
	}

	err = h.clients.CreateBillCharge(
		ctx,
		req.PatientID,
		dispense.ID.String(), // source_ref_id = dispense ID
		fmt.Sprintf("Dispensed %s - %dx", dispense.Drug.Name, dispense.Quantity),
		dispense.Quantity,
		dispense.Total,
	)
	if err != nil {
		log.Printf("WARNING: Billing charge failed for dispense %s: %v", dispense.ID, err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Created",
		"dispense": dispense,
	})
}

func (h *DispenseHandler) GetDispense(c *gin.Context) {
	ctx := c.Request.Context()

	dispenseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string: " + err.Error()})
		return
	}

	dispense, err := h.svc.FindByID(ctx, dispenseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dispense: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dispense": dispense})

}

func (h *DispenseHandler) GetPrescription(c *gin.Context) {
	ctx := c.Request.Context()

	prescriptionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string: " + err.Error()})
		return
	}

	prescription, err := h.svc.FindPrescriptionByID(ctx, prescriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prescription: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"prescription": prescription})

}

func (h *DispenseHandler) UpdateDispense(c *gin.Context) {
	ctx := c.Request.Context()

	userId, exists := c.MustGet("user_id").(string)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You're not logged in"})
		return
	}

	dispenseBy, err := uuid.Parse(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id format: " + err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string: " + err.Error()})
		return
	}

	var req models.UpdateDispenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if err := models.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalidation failed",
			"details": err.Error(),
		})
		return
	}

	if err = h.svc.UpdateDispense(ctx, id, dispenseBy, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Update successful"})
}

func (h *DispenseHandler) GetDispenseByDrugID(c *gin.Context) {
	ctx := c.Request.Context()

	drugID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string: " + err.Error()})
		return
	}

	dispense, err := h.svc.FindDispenseByDrugID(ctx, drugID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dispense": dispense})
}
