package handlers

import (
	"errors"
	"net/http"

	"github.com/Eucastan/hms/auth/internal/models"
	"github.com/Eucastan/hms/auth/internal/services"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc services.AuthService
}

func NewAuthHandler(svc services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	var req models.StaffCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID, err := h.svc.Register(ctx, &req)
	if err != nil {
		if errors.Is(err, services.ErrValidation) || errors.Is(err, services.ErrWeakPassword) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"user_id": userID,
	})

}

func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, userID, err := h.svc.Login(ctx, &req)
	if err != nil {

		if errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user_id": userID,
	})
}

func (h *AuthHandler) UpdateStaff(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	updaterRole, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user role"})
		return
	}

	var input models.StaffUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	if err := h.svc.UpdateStaff(ctx, id, &input, updaterRole.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (h *AuthHandler) Profile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user ID"})
		return
	}

	getUserID, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":  getUserID,
		"message": "Login successful",
	})
}

func (h *AuthHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.Param("id")
	id, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	roleStr, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user role"})
		return
	}

	if err := h.svc.DeleteStaff(ctx, id, roleStr.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed: " + userIDStr})

}
