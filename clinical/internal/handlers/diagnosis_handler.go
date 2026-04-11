package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Eucastan/hms/clinical/internal/dto"
	"github.com/Eucastan/hms/clinical/internal/repositories"
	"github.com/Eucastan/hms/clinical/internal/services"
)

type DiagnosisHandler struct {
	svc      services.DiagnosisService
	validate *validator.Validate
}

func NewDiagnosisHandler(svc services.DiagnosisService) *DiagnosisHandler {
	return &DiagnosisHandler{
		svc:      svc,
		validate: dto.Validate,
	}
}

func (h *DiagnosisHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	doctorIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	doctorID, err := uuid.Parse(doctorIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user identity"})
		return
	}

	var input dto.DiagnosisCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	if err := h.validate.Struct(input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	diag, err := h.svc.CreateDiagnosis(ctx, doctorID, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create diagnosis", "details": err.Error()})
		return
	}

	resp := dto.DiagnosisResponse{
		ID:          diag.ID.String(),
		PatientID:   input.PatientID.String(),
		DoctorID:    doctorID.String(),
		Code:        input.Code,
		Description: input.Description,
	}

	if input.AdmissionID != uuid.Nil {
		admissionStr := input.AdmissionID.String()
		resp.AdmissionID = &admissionStr
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *DiagnosisHandler) GetDiagnosis(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error: " + err.Error()})
		return
	}

	diagnosis, err := h.svc.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": repositories.ErrNotFound})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Diagnosis: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"diagnosis": diagnosis})
}

func (h *DiagnosisHandler) UpdateDiagnosis(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found in context"})
		return
	}

	doctorID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error: " + err.Error()})
		return
	}

	diagID := c.Param("id")

	diagnosID, err := uuid.Parse(diagID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error: " + err.Error()})
		return
	}

	var req dto.DiagnosisUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid requests: " + err.Error()})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Code != nil {
		updates["code"] = *req.Code
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	err = h.svc.UpdateDiagnosis(ctx, diagnosID, doctorID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Diagnosis: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Update successful"})
}

func (h *DiagnosisHandler) DeleteDiagnosis(c *gin.Context) {
	ctx := c.Request.Context()
	diagID := c.Param("id")

	diagnosID, err := uuid.Parse(diagID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error: " + err.Error()})
		return
	}

	if err := h.svc.DeleteDiagnosis(ctx, diagnosID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Record deleted"})
}
