package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/Eucastan/hms/patient/internal/models"
	"github.com/Eucastan/hms/patient/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PatientHandler struct {
	svc services.PatientService
}

func NewPatientHandler(svc services.PatientService) *PatientHandler {
	return &PatientHandler{svc: svc}
}

func (h *PatientHandler) CreatePatientRecord(c *gin.Context) {
	ctx := c.Request.Context()

	var req models.PatientCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	pts, err := h.svc.CreatePatient(ctx, req)

	if err != nil {

		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		if errors.Is(err, services.ErrFailedValidation) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create patient",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Patient record created successfully",
		"patient": pts,
	})
}

func (h *PatientHandler) GetPatientRecord(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string"})
		return
	}

	pts, err := h.svc.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured fetching records " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"records": pts,
	})
}

func (h *PatientHandler) UpdatePatientRecord(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	patientID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string"})
		return
	}

	var req models.PatientUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request " + err.Error()})
		return
	}

	if err := h.svc.UpdatePatient(ctx, patientID, req); err != nil {
		if errors.Is(err, services.ErrFailedValidation) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Update successful"})
}

func (h *PatientHandler) GetAdmissionRecords(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string"})
		return
	}

	admsn, err := h.svc.FindAdmissionByID(ctx, id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured fetching records " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"records": admsn,
	})
}

func (h *PatientHandler) UpdateAdmissionRecord(c *gin.Context) {
	ctx := c.Request.Context()

	admsnID := c.Param("id")
	id, err := uuid.Parse(admsnID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string"})
		return
	}

	var req models.AdmissionUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request " + err.Error()})
		return
	}

	if err := h.svc.UpdateAdmission(ctx, id, &req); err != nil {
		if errors.Is(err, services.ErrFailedValidation) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Update successful"})
}

func (h *PatientHandler) DeletePatientRecord(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	patientID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string"})
		return
	}

	if err := h.svc.DeletePatient(ctx, patientID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Record deleted"})
}

func (h *PatientHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "Healthy",
		"service":   "Patient Service",
		"version":   "v1.0.0",
		"uptime":    time.Since(time.Now().Add(-time.Hour)).String(),
		"timestamp": time.Now().UTC(),
		"message":   "Patient Service is running smoothly",
	})
}

func (h *PatientHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "Alive",
		"service":   "Patient Service",
		"version":   "v1.0.0",
		"uptime":    time.Since(time.Now().Add(-time.Hour)).String(),
		"timestamp": time.Now().UTC(),
		"message":   "Patient Service is running smoothly",
	})
}

func (h *PatientHandler) Readiness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "Ready",
		"service":   "Patient Service",
		"version":   "v1.0.0",
		"uptime":    time.Since(time.Now().Add(-time.Hour)).String(),
		"timestamp": time.Now().UTC(),
		"message":   "Patient Service is running smoothly",
	})
}
