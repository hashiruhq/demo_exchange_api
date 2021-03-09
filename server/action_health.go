package server

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

// HealthGoroutineThreshold -- the total number of goroutines that signal a heavy load on the container
const HealthGoroutineThreshold = 10000

// AddHealthRoutes godoc
func (srv *server) AddHealthRoutes(r *gin.Engine) {
	if !srv.Config.Server.API.Health {
		return
	}

	health := r.Group("/health")
	{
		health.GET("/live", LivenessCheckHandler)
		health.HEAD("/live", LivenessCheckHandler)

		health.GET("/ready", ReadinessCheckHandler)
		health.HEAD("/ready", ReadinessCheckHandler)
	}
}

// LivenessCheckHandler -- checks if server is properly running according to internal checks.
// If not, server is considered to have failed and needs to be restarted.
// Liveness probes are used to detect situations where the application
// has gone into a state where it can not recover except by being restarted.
func LivenessCheckHandler(c *gin.Context) {
	// @todo Create verification methods for this check
	c.Status(http.StatusOK)
}

// ReadinessCheckHandler -- checks if there are more than threshold
// number of goroutines running, returns service unavailable.
//
// Readiness probes are used to detect situations where application
// is under heavy load and temporarily unable to serve. In a orchestrated
// setup like Kubernetes, containers reporting that they are not ready do
// not receive traffic through Kubernetes Services.
func ReadinessCheckHandler(c *gin.Context) {
	if err := goroutineCountCheck(HealthGoroutineThreshold); err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	c.Status(http.StatusOK)
}

// checks threshold against total number of go-routines in the system and
// throws error if more than threshold go-routines are running.
func goroutineCountCheck(threshold int) error {
	count := runtime.NumGoroutine()
	if count > threshold {
		return fmt.Errorf("too many goroutines (%d > %d)", count, threshold)
	}
	return nil
}
