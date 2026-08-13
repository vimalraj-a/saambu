package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type HealthHandler struct {
	Mongo *mongo.Client
}

func NewHealthHandler(mongoClient *mongo.Client) *HealthHandler {
	return &HealthHandler{Mongo: mongoClient}
}

func (h *HealthHandler) Check(c *gin.Context) {
	if err := h.Mongo.Ping(c.Request.Context(), nil); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "down",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
