package agent

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"

	"github.com/vimalraj-a/mongo-hack/server/internal/analysis"
	"github.com/vimalraj-a/mongo-hack/server/internal/capture"
)

// Live smoke test: launches Chrome and calls the configured vision model.
// Skipped automatically when no vision credentials are set.
func TestSmokeRunSteps(t *testing.T) {
	_ = godotenv.Load("../../.env")

	apiKey := os.Getenv("VISION_LLM_API_KEY")
	model := os.Getenv("VISION_LLM_MODEL")
	if apiKey == "" || model == "" {
		t.Skip("VISION_LLM_API_KEY / VISION_LLM_MODEL not set in server/.env — skipping live smoke test")
	}
	baseURL := os.Getenv("VISION_LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	client := analysis.NewClient(baseURL, apiKey, model)

	session, err := capture.NewSession(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer session.Close()

	steps := []string{
		`Click the "Learn more" link`,
		`Expect the page to no longer be example.com`,
	}

	executions, passed := RunSteps(context.Background(), session, client, steps, true)

	for _, se := range executions {
		t.Logf("step %d %q -> action=%+v assertion=%+v error=%q", se.Index, se.NLText, se.Action, se.Assertion, se.Error)
	}
	if !passed {
		t.Fatalf("expected the run to pass, got overallPassed=false")
	}
	if len(executions) != 2 {
		t.Fatalf("expected 2 step executions, got %d", len(executions))
	}
	if executions[0].Action.Type != "click" {
		t.Fatalf("expected step 0 to be a click action, got %q", executions[0].Action.Type)
	}
}
