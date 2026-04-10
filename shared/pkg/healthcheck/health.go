package healthcheck

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthCheckHandlers struct {
	ServiceName string
	Version     string
	StartTime   time.Time
}

func NewHealthCheckHandlers(serviceName, version string) *HealthCheckHandlers {
	return &HealthCheckHandlers{
		ServiceName: serviceName,
		Version:     version,
		StartTime:   time.Now().UTC(),
	}
}

func (h *HealthCheckHandlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "Healthy",
		"service":   h.ServiceName,
		"version":   h.Version,
		"uptime":    time.Since(h.StartTime).String(),
		"timestamp": h.StartTime,
	})
}

func (h *HealthCheckHandlers) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "Alive",
		"service": h.ServiceName,
		"version": h.Version,
	})
}

func (h *HealthCheckHandlers) Readiness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "Ready",
		"service": h.ServiceName,
		"version": h.Version,
	})
}
