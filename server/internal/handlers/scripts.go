package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/vimalraj-a/mongo-hack/server/internal/agent"
	"github.com/vimalraj-a/mongo-hack/server/internal/analysis"
	"github.com/vimalraj-a/mongo-hack/server/internal/capture"
	"github.com/vimalraj-a/mongo-hack/server/internal/models"
)

type ScriptsHandler struct {
	Scripts        *mongo.Collection
	ChangeRequests *mongo.Collection
	Runs           *mongo.Collection
	Captures       *mongo.Collection
	Vision         *analysis.Client
}

func NewScriptsHandler(db *mongo.Database, vision *analysis.Client) *ScriptsHandler {
	return &ScriptsHandler{
		Scripts:        db.Collection("test_scripts"),
		ChangeRequests: db.Collection("change_requests"),
		Runs:           db.Collection("execution_runs"),
		Captures:       db.Collection("captures"),
		Vision:         vision,
	}
}

func (h *ScriptsHandler) fetchCapture(ctx context.Context, id bson.ObjectID) (models.Capture, error) {
	var cap models.Capture
	err := h.Captures.FindOne(ctx, bson.M{"_id": id}).Decode(&cap)
	return cap, err
}

func (h *ScriptsHandler) fetchScript(ctx context.Context, id bson.ObjectID) (models.TestScript, error) {
	var script models.TestScript
	err := h.Scripts.FindOne(ctx, bson.M{"_id": id}).Decode(&script)
	return script, err
}

// GenerateScript is POST /api/captures/:id/script — proposes the initial NL
// test script (v1) from a capture's screenshot + element inventory.
func (h *ScriptsHandler) GenerateScript(c *gin.Context) {
	captureID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid capture id"})
		return
	}

	ctx := c.Request.Context()
	cap, err := h.fetchCapture(ctx, captureID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "capture not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	steps, err := analysis.GenerateNLScript(ctx, h.Vision, cap)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	script := models.TestScript{
		CaptureID: cap.ID,
		Steps:     steps,
		Version:   1,
		Status:    models.ScriptStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}

	insertCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := h.Scripts.InsertOne(insertCtx, script)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	script.ID = res.InsertedID.(bson.ObjectID)

	c.JSON(http.StatusCreated, script)
}

// GetScript is GET /api/scripts/:id — current steps plus the change-request
// chat history.
func (h *ScriptsHandler) GetScript(c *gin.Context) {
	scriptID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid script id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	script, err := h.fetchScript(ctx, scriptID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cursor, err := h.ChangeRequests.Find(ctx, bson.M{"testScriptId": scriptID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var changeRequests []models.ChangeRequest
	if err := cursor.All(ctx, &changeRequests); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"script": script, "changeRequests": changeRequests})
}

type postMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

// PostMessage is POST /api/scripts/:id/messages — the only way a script's
// steps ever change. Grounds the request against the capture's screenshot
// and HTML before either applying it or asking a clarifying question.
func (h *ScriptsHandler) PostMessage(c *gin.Context) {
	scriptID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid script id"})
		return
	}
	var req postMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	script, err := h.fetchScript(ctx, scriptID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cap, err := h.fetchCapture(ctx, script.CaptureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, err := analysis.ReviewChangeRequest(ctx, h.Vision, script.Steps, cap, req.Message)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	changeReq := models.ChangeRequest{
		TestScriptID:  scriptID,
		Message:       req.Message,
		ResponseType:  result.Type,
		ResponseText:  result.Question,
		PreviousSteps: script.Steps,
		CreatedAt:     now,
	}

	if result.Type == models.ChangeResponseUpdated {
		changeReq.NewSteps = result.NewSteps
		script.Steps = result.NewSteps
		script.Version++
		script.UpdatedAt = now
		script.Status = models.ScriptStatusDraft // a spec edit invalidates any prior resolution

		updateCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := h.Scripts.UpdateOne(updateCtx, bson.M{"_id": scriptID}, bson.M{"$set": bson.M{
			"steps":     script.Steps,
			"version":   script.Version,
			"status":    script.Status,
			"updatedAt": script.UpdatedAt,
		}}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	insertCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := h.ChangeRequests.InsertOne(insertCtx, changeReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	changeReq.ID = res.InsertedID.(bson.ObjectID)

	c.JSON(http.StatusOK, gin.H{"changeRequest": changeReq, "script": script})
}

// Verify is POST /api/scripts/:id/verify — opens a fresh browser session,
// replays the capture's prerequisite steps (not judged, not persisted per
// run — it's already visible on the Capture), then runs the current test
// steps with assertion judging on, and stores/returns that as an
// ExecutionRun storyboard.
func (h *ScriptsHandler) Verify(c *gin.Context) {
	scriptID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid script id"})
		return
	}

	ctx := c.Request.Context()

	script, err := h.fetchScript(ctx, scriptID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cap, err := h.fetchCapture(ctx, script.CaptureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	session, err := capture.NewSession(ctx, cap.URL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	if len(cap.PrerequisiteSteps) > 0 {
		agent.RunSteps(ctx, session, h.Vision, cap.PrerequisiteSteps, false)
	}

	executions, passed := agent.RunSteps(ctx, session, h.Vision, script.Steps, true)

	run := models.ExecutionRun{
		TestScriptID:  scriptID,
		ScriptVersion: script.Version,
		Steps:         executions,
		OverallPassed: passed,
		CreatedAt:     time.Now(),
	}

	insertCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := h.Runs.InsertOne(insertCtx, run)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	run.ID = res.InsertedID.(bson.ObjectID)

	c.JSON(http.StatusOK, run)
}
