package router

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/vimalraj-a/mongo-hack/server/internal/handlers"
)

func New(mongoClient *mongo.Client) *gin.Engine {
	r := gin.Default()

	health := handlers.NewHealthHandler(mongoClient)
	r.GET("/health", health.Check)

	return r
}
