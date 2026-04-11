package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Eucastan/hms/lab/internal/models"
	"github.com/Eucastan/hms/lab/internal/services"

	"net/http"
)

type LabResultHandler struct {
	svc services.LabResultService
}

func NewLabResultHandler(svc services.LabResultService) *LabResultHandler {
	return &LabResultHandler{svc: svc}
}

func (h *LabResultHandler) CreateLabResult(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	performedBy, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user identity: " + err.Error()})
		return
	}

	var req models.LabResultCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json requests: " + err.Error()})
		return
	}

	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string: " + err.Error()})
		return
	}

	if err := models.Validate.Struct(req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	result, err := h.svc.CreateLabResult(
		ctx,
		performedBy,
		patientID,
		req.TestType,
		req.ResultValue,
		req.Unit,
		req.ReferenceRange,
		req.Comments,
		req.Verified,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lab result: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"result":  result,
		"message": "Lab result created",
	})
}

func (h *LabResultHandler) GetLabResult(c *gin.Context) {
	ctx := c.Request.Context()

	patientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error occured converting string: " + err.Error()})
		return
	}

	labResult, err := h.svc.GetByPatientID(ctx, patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Lab result: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": labResult})
}

func (h *LabResultHandler) UpdateLabResult(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	performedBy, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user identity: " + err.Error()})
		return
	}

	labID := c.Param("id")

	labResultID, err := uuid.Parse(labID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error: " + err.Error()})
		return
	}

	var req models.LabResultUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json requests: " + err.Error()})
		return
	}

	if err := models.Validate.Struct(req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	err = h.svc.UpdateLabResult(ctx, labResultID, performedBy, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lab result: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Update successful"})
}

func (h *LabResultHandler) DeleteLabResult(c *gin.Context) {
	ctx := c.Request.Context()

	labResultID := c.Param("id")

	labID, err := uuid.Parse(labResultID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error: " + err.Error()})
		return
	}

	if err := h.svc.DeleteLabResult(ctx, labID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Record deleted"})
}
