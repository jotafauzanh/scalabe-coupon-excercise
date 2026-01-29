package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jotafauzanh/scalabe-coupon-excercise/pkg/db"
	"github.com/jotafauzanh/scalabe-coupon-excercise/pkg/redis"
)

type DevController struct{}

func NewDevController() *DevController {
	return &DevController{}
}

type HealthStatus struct {
	Status    string                   `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Services  map[string]ServiceHealth `json:"services"`
}

type ServiceHealth struct {
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency,omitempty"`
}

func (c *DevController) HealthCheck(ctx *gin.Context) {
	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  make(map[string]ServiceHealth),
	}

	// Check database health
	dbStart := time.Now()
	sqlDB, err := db.DB.DB()
	if err != nil {
		health.Services["database"] = ServiceHealth{
			Status:  "unhealthy",
			Message: "failed to get database instance: " + err.Error(),
		}
		health.Status = "unhealthy"
	} else {
		err = sqlDB.Ping()
		latency := time.Since(dbStart)
		if err != nil {
			health.Services["database"] = ServiceHealth{
				Status:  "unhealthy",
				Message: "database ping failed: " + err.Error(),
				Latency: latency,
			}
			health.Status = "unhealthy"
		} else {
			health.Services["database"] = ServiceHealth{
				Status:  "healthy",
				Latency: latency,
			}
		}
	}

	// Check Redis health
	redisStart := time.Now()
	err = redis.CheckHealth()
	latency := time.Since(redisStart)
	if err != nil {
		health.Services["redis"] = ServiceHealth{
			Status:  "unhealthy",
			Message: err.Error(),
			Latency: latency,
		}
		health.Status = "unhealthy"
	} else {
		health.Services["redis"] = ServiceHealth{
			Status:  "healthy",
			Latency: latency,
		}
	}

	// Check service health (always healthy if server is running)
	health.Services["service"] = ServiceHealth{
		Status: "healthy",
	}

	// Determine HTTP status code
	statusCode := http.StatusOK
	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	ctx.JSON(statusCode, health)
}
