package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/vimalraj-a/mongo-hack/server/internal/analysis"
	"github.com/vimalraj-a/mongo-hack/server/internal/codegen"
	"github.com/vimalraj-a/mongo-hack/server/internal/models"
)

type GeneratedTestsHandler struct {
	GeneratedTests *mongo.Collection
	Scripts        *mongo.Collection
	Runs           *mongo.Collection
	Captures       *mongo.Collection
	Coder          *analysis.Client
}

func NewGeneratedTestsHandler(db *mongo.Database, coder *analysis.Client) *GeneratedTestsHandler {
	return &GeneratedTestsHandler{
		GeneratedTests: db.Collection("generated_tests"),
		Scripts:        db.Collection("test_scripts"),
		Runs:           db.Collection("execution_runs"),
		Captures:       db.Collection("captures"),
		Coder:          coder,
	}
}

type resolveRequest struct {
	RunID string `json:"runId" binding:"required"`
}

// Finalize is POST /api/scripts/:id/finalize — only valid when the named run
// fully passed. Builds the real (not expected-to-fail) generated test.
func (h *GeneratedTestsHandler) Finalize(c *gin.Context) {
	h.resolve(c, false, models.ScriptStatusResolvedPass)
}

// LockKnownFailure is POST /api/scripts/:id/lock-known-failure — used when
// the user confirms their description was right and the app is wrong.
// Builds a generated test that still encodes the correct expected
// behavior, annotated as an expected failure.
func (h *GeneratedTestsHandler) LockKnownFailure(c *gin.Context) {
	h.resolve(c, true, models.ScriptStatusResolvedKnownFailure)
}

func (h *GeneratedTestsHandler) resolve(c *gin.Context, expectedToFail bool, resolvedStatus models.ScriptStatus) {
	scriptID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid script id"})
		return
	}
	var req resolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	runID, err := bson.ObjectIDFromHex(req.RunID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid runId"})
		return
	}

	ctx := c.Request.Context()

	var script models.TestScript
	if err := h.Scripts.FindOne(ctx, bson.M{"_id": scriptID}).Decode(&script); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var run models.ExecutionRun
	if err := h.Runs.FindOne(ctx, bson.M{"_id": runID, "testScriptId": scriptID}).Decode(&run); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "execution run not found for this script"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !expectedToFail && !run.OverallPassed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot finalize a run that did not pass — use lock-known-failure instead, or send a change request and re-verify"})
		return
	}
	if expectedToFail && run.OverallPassed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this run already passed — there's nothing to lock as a known failure"})
		return
	}

	var cap models.Capture
	if err := h.Captures.FindOne(ctx, bson.M{"_id": script.CaptureID}).Decode(&cap); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	code, err := codegen.GenerateWithLLM(ctx, h.Coder, cap.PrerequisiteRun, run, expectedToFail)
	if err != nil || codegen.ValidateGeneratedCode(code, cap.PrerequisiteRun, run) != nil {
		// LLM output missing (error) or failed the selector/value guardrail —
		// fall back to the deterministic template rather than risk a
		// hallucinated spec reaching the user silently.
		code = codegen.BuildSpec(cap.PrerequisiteRun, run, expectedToFail)
	}

	gt := models.GeneratedTest{
		TestScriptID:   scriptID,
		SourceRunID:    runID,
		Code:           code,
		ExpectedToFail: expectedToFail,
		CreatedAt:      time.Now(),
	}
	res, err := h.GeneratedTests.InsertOne(ctx, gt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	gt.ID = res.InsertedID.(bson.ObjectID)

	if _, err := h.Scripts.UpdateOne(ctx, bson.M{"_id": scriptID}, bson.M{"$set": bson.M{
		"status":    resolvedStatus,
		"updatedAt": time.Now(),
	}}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gt)
}

// Download is GET /api/generated-tests/:id/download — serves the stored
// code as a downloadable .spec.ts file.
func (h *GeneratedTestsHandler) Download(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid generated test id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var gt models.GeneratedTest
	if err := h.GeneratedTests.FindOne(ctx, bson.M{"_id": id}).Decode(&gt); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "generated test not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("test-%s.spec.ts", gt.ID.Hex())
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/typescript; charset=utf-8", []byte(gt.Code))
}
