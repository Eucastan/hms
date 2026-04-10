package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Eucastan/hms/pharmacy/internal/models"
	"github.com/Eucastan/hms/pharmacy/internal/services"
)

type DrugHandler struct {
	svc services.DrugService
}

func NewDrugHandler(svc services.DrugService) *DrugHandler {
	return &DrugHandler{svc: svc}
}

func (h *DrugHandler) CreateDrug(c *gin.Context) {
	ctx := c.Request.Context()

	var req models.DrugCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if err := models.Validate.Struct(req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	drug, err := h.svc.Create(
		ctx,
		req.Name,
		req.GenericName,
		req.Form,
		req.Strength,
		req.PackSize,
		req.Stock,
		req.PricePerUnit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create drug: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Drug created",
		"drug":    drug,
	})
}

func (h *DrugHandler) GetDrug(c *gin.Context) {
	ctx := c.Request.Context()

	drugID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string: " + err.Error()})
		return
	}

	drug, err := h.svc.FindByID(ctx, drugID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get drug: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"drug": drug})

}

func (h *DrugHandler) UpdateDrug(c *gin.Context) {
	ctx := c.Request.Context()

	drugID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string: " + err.Error()})
		return
	}

	var req models.DrugUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if err := models.Validate.Struct(req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	if err = h.svc.Update(ctx, drugID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update drug: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Update successful"})
}

func (h *DrugHandler) DeleteDrug(c *gin.Context) {
	ctx := c.Request.Context()

	drugID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting string: " + err.Error()})
		return
	}

	err = h.svc.Delete(ctx, drugID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete drug: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Drug deleted"})
}
