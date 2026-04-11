package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Eucastan/hms/lab/internal/grpc/client"
	"github.com/Eucastan/hms/lab/internal/models"
	"github.com/Eucastan/hms/lab/internal/services"

	"net/http"
)

type LabRequestHandler struct {
	svc services.LabRequestService
	c   *client.Clients
}

func NewLabRequestHandler(svc services.LabRequestService) *LabRequestHandler {
	return &LabRequestHandler{svc: svc}
}

func (h *LabRequestHandler) CreateLabRequest(c *gin.Context) {
	ctx := c.Request.Context()

	requestByStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	requestBy, err := uuid.Parse(requestByStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user identity"})
		return
	}

	var req models.LabRequestCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json requests: " + err.Error()})
		return
	}

	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user identity"})
		return
	}

	if err := models.Validate.Struct(req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	/*if err := h.c.ValidatePatientID(ctx, req.PatientID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Can not verify patient " + err.Error()})
		return
	}*/

	labRequest, err := h.svc.CreateLabRequest(
		ctx,
		patientID,
		requestBy,
		req.TestType,
		req.Priority,
		"requested",
		req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lab request: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"labreq":  labRequest,
		"message": "Lab request created",
	})
}

func (h *LabRequestHandler) GetLabRequest(c *gin.Context) {
	ctx := c.Request.Context()

	patientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string: " + err.Error()})
		return
	}

	/*if err := h.c.ValidatePatientID(ctx, patientID.String()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Can not verify patient " + err.Error()})
		return
	}*/

	labReq, err := h.svc.GetLabRequestsByPatient(ctx, patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Lab request: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"labreq": labReq})
}

func (h *LabRequestHandler) UpdateLabRequest(c *gin.Context) {
	ctx := c.Request.Context()

	requestByStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	requestBy, err := uuid.Parse(requestByStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user identity: " + err.Error()})
		return
	}

	labReqID := c.Param("id")

	labRequestID, err := uuid.Parse(labReqID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error: " + err.Error()})
		return
	}

	var req models.LabRequestUpdate
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

	err = h.svc.UpdateLabRequest(ctx, labRequestID, requestBy, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lab request: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Update successful"})
}

func (h *LabRequestHandler) DeleteLabRequest(c *gin.Context) {
	ctx := c.Request.Context()

	labReqID := c.Param("id")

	labRequestID, err := uuid.Parse(labReqID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error: " + err.Error()})
		return
	}

	if err := h.svc.DeleteLabRequest(ctx, labRequestID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Record deleted"})
}
