package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ginHandler handles HTTP requests with telemetry.
type ginHandler struct{}

// NewGinHandler creates a new Handler.
func NewGinHandler() (Handler, error) {
	return &ginHandler{}, nil // returning an error just for consistency of the examples
}

// GetUUID generates and returns a UUID with full telemetry.
func (gh *ginHandler) GetUUID(c *gin.Context) {
	newUUID := uuid.New().String()
	c.Set(uuidKey, newUUID)
	c.JSON(http.StatusOK, gin.H{"uuid": newUUID})
}
