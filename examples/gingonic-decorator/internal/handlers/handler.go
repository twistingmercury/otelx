package handlers

import "github.com/gin-gonic/gin"

const (
	uuidKey = "uuid"
)

type Handler interface {
	GetUUID(c *gin.Context)
}
