package handlers

import (
	"net/http"

	"github.com/Eucastan/hms/clinical/internal/grpc/client"
	"github.com/Eucastan/hms/clinical/internal/models"
	"github.com/Eucastan/hms/shared/pkg/grpcserver"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GRPCClientHandler struct {
	grpcClient *client.SendToAllClients
}

func NewGRPCClientHandler(grpcClient client.SendToAllClients) *GRPCClientHandler {
	return &GRPCClientHandler{grpcClient: &grpcClient}
}

func (h *GRPCClientHandler) CreateLabRequest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	token := c.GetString("token")

	ctx := c.Request.Context()

	ctx = grpcserver.AppendJWTToContext(ctx, token)

	var req models.LabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.RequestBy = userID.(string)

	// Send to Lab service via gRPC
	if err := h.grpcClient.SendLabTestRequest(ctx, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send lab request: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Lab request sent successfully"})
}

func (h *GRPCClientHandler) CreatePrescription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	token := c.GetString("token")

	ctx := c.Request.Context()

	ctx = grpcserver.AppendJWTToContext(ctx, token)

	var req models.CreatePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.DoctorID = userID.(string)

	// Send to Pharmacy service via gRPC
	if err := h.grpcClient.SendPrescriptionToPharmacy(ctx, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send prescription: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Prescription sent successfully"})
}

func (h *GRPCClientHandler) UpdateLabRequest(c *gin.Context) {
	ctx := c.Request.Context()

	token := c.GetString("token")

	ctx = grpcserver.AppendJWTToContext(ctx, token)

	requestByStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	requestBy := requestByStr.(string)

	labReqID := c.Param("id")

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

	err := h.grpcClient.UpdateLabRequest(ctx, labReqID, requestBy, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lab request: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Update successful"})
}

func (h *GRPCClientHandler) GetPatient(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error converting uuid"})
		return
	}

	token := c.GetString("token")

	ctx := c.Request.Context()

	ctx = grpcserver.AppendJWTToContext(ctx, token)

	patient, err := h.grpcClient.GetPatient(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found: " + err.Error()})
		return
	}

	c.JSON(200, patient)
}

func (h *GRPCClientHandler) SearchPatient(c *gin.Context) {
	name := c.Query("name")
	hospitalNo := c.Query("hospitalNo")

	if name == "" && hospitalNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "at least one of 'name' or 'hospitalNo' query parameter is required",
		})
		return
	}

	token := c.GetString("token")

	ctx := c.Request.Context()
	ctx = grpcserver.AppendJWTToContext(ctx, token)

	patients, err := h.grpcClient.SearchPatients(ctx, name, hospitalNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search patient: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"patients": patients,
		"count":    len(patients),
	})

}
