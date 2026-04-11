package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/Eucastan/hms/api-gateway/configs"
	"github.com/Eucastan/hms/shared/pkg/auth"
	"github.com/Eucastan/hms/shared/pkg/healthcheck"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	r := gin.Default()
	r.Use(auth.RateLimiter())

	// Routes
	api := r.Group("/api/v1")
	{
		api.Any("/patients/*proxyPath", proxyHandler(cfg.AuthServiceURL))
		api.Any("/patients/*proxyPath", proxyHandler(cfg.PatientServiceURL))
		api.Any("/lab/*proxyPath", proxyHandler(cfg.LabServiceURL))
		api.Any("/pharmacy/*proxyPath", proxyHandler(cfg.PharmacyServiceURL))
		api.Any("/billing/*proxyPath", proxyHandler(cfg.BillingServiceURL))
		api.Any("/clinical/*proxyPath", proxyHandler(cfg.ClinicalServiceURL))
	}

	// Health check
	healthHandler := healthcheck.NewHealthCheckHandlers(cfg.ServiceName, cfg.Version)
	r.GET("/health", healthHandler.Health)
	r.GET("/liveness", healthHandler.Liveness)
	r.GET("/readiness", healthHandler.Readiness)

	log.Printf("API Gateway running on %s", cfg.Port)
	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("Gateway failed: %v", err)
	}
}

func proxyHandler(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		remote, err := url.Parse(target)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "bad gateway config"})
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(remote)

		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.URL.Path = remote.Path + c.Param("proxyPath")

			// Forward Authorization header (JWT) - services will validate themselves
			if token := c.GetHeader("Authorization"); token != "" {
				req.Header.Set("Authorization", token)
			}

			if userID, exists := c.Get("user_id"); exists {
				req.Header.Set("X-User-ID", userID.(string))
			}
			if role, exists := c.Get("role"); exists {
				req.Header.Set("X-Role", role.(string))
			}

			req.Header.Set("X-Request-ID", c.GetHeader("X-Request-ID"))
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
