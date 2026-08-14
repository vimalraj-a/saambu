package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/vimalraj-a/mongo-hack/server/internal/agent"
	"github.com/vimalraj-a/mongo-hack/server/internal/analysis"
	"github.com/vimalraj-a/mongo-hack/server/internal/capture"
	"github.com/vimalraj-a/mongo-hack/server/internal/models"
)

type CapturesHandler struct {
	Collection *mongo.Collection
	Vision     *analysis.Client
}

func NewCapturesHandler(db *mongo.Database, vision *analysis.Client) *CapturesHandler {
	return &CapturesHandler{Collection: db.Collection("captures"), Vision: vision}
}

type createCaptureRequest struct {
	URL              string `json:"url" binding:"required"`
	PrerequisiteText string `json:"prerequisiteText"`
}

// Create drives the browser through the prerequisite instruction (if any),
// then captures a screenshot + HTML + element inventory of wherever the
// session lands — that becomes the starting point for everything
// downstream. Synchronous: this can take a while if the prerequisite
// involves several steps, each needing an LLM call.
func (h *CapturesHandler) Create(c *gin.Context) {
	var req createCaptureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	session, err := capture.NewSession(ctx, req.URL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	doc := models.Capture{
		URL:              req.URL,
		PrerequisiteText: req.PrerequisiteText,
		CreatedAt:        time.Now(),
	}

	if req.PrerequisiteText != "" {
		steps, err := analysis.SplitInstructions(ctx, h.Vision, req.PrerequisiteText)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		doc.PrerequisiteSteps = steps
		executions, _ := agent.RunSteps(ctx, session, h.Vision, steps, false)
		doc.PrerequisiteRun = executions
	}

	title, err := session.Title()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	doc.Title = title

	screenshot, err := session.Screenshot()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	doc.ScreenshotBase64 = screenshot

	html, err := session.ExtractHTML()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	doc.HTMLSnapshot = html

	elements, err := session.ExtractElements()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	doc.Elements = elements

	insertCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := h.Collection.InsertOne(insertCtx, doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	doc.ID = res.InsertedID.(bson.ObjectID)

	c.JSON(http.StatusCreated, doc)
}

func (h *CapturesHandler) Get(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid capture id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var doc models.Capture
	if err := h.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "capture not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}
