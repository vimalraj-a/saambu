package router

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/vimalraj-a/mongo-hack/server/internal/analysis"
	"github.com/vimalraj-a/mongo-hack/server/internal/config"
	"github.com/vimalraj-a/mongo-hack/server/internal/handlers"
)

func New(mongoClient *mongo.Client, cfg config.Config) *gin.Engine {
	r := gin.Default()

	db := mongoClient.Database(cfg.MongoDBName)
	vision := analysis.NewClient(cfg.VisionLLMBaseURL, cfg.VisionLLMAPIKey, cfg.VisionLLMModel)
	coder := analysis.NewClient(cfg.CoderLLMBaseURL, cfg.CoderLLMAPIKey, cfg.CoderLLMModel)

	health := handlers.NewHealthHandler(mongoClient)
	r.GET("/health", health.Check)

	captures := handlers.NewCapturesHandler(db, vision)
	r.POST("/api/captures", captures.Create)
	r.GET("/api/captures/:id", captures.Get)

	scripts := handlers.NewScriptsHandler(db, vision)
	r.POST("/api/captures/:id/script", scripts.GenerateScript)
	r.GET("/api/scripts/:id", scripts.GetScript)
	r.POST("/api/scripts/:id/messages", scripts.PostMessage)
	r.POST("/api/scripts/:id/verify", scripts.Verify)

	generatedTests := handlers.NewGeneratedTestsHandler(db, coder)
	r.POST("/api/scripts/:id/finalize", generatedTests.Finalize)
	r.POST("/api/scripts/:id/lock-known-failure", generatedTests.LockKnownFailure)
	r.GET("/api/generated-tests/:id/download", generatedTests.Download)

	return r
}
