package analysis

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// This is a live smoke test against whatever provider/model is configured in
// server/.env — it is skipped automatically when no API key is set, so it
// never fails a build with no credentials. Run it directly with:
//
//	cd server && go test ./internal/analysis/... -run TestSmoke -v
func TestSmokeVisionClient(t *testing.T) {
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

	client := NewClient(baseURL, apiKey, model)
	tool := Tool{
		Name:        "report_capital",
		Description: "Report the capital city of a country.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": { "capital": { "type": "string" } },
			"required": ["capital"],
			"additionalProperties": false
		}`),
	}

	args, err := client.CallWithTool(context.Background(), []Message{
		TextMessage("user", "What is the capital of France? Call the tool with the answer."),
	}, tool)
	if err != nil {
		t.Fatalf("vision model %q smoke test failed: %v", model, err)
	}

	var out struct {
		Capital string `json:"capital"`
	}
	if err := json.Unmarshal(args, &out); err != nil {
		t.Fatalf("could not parse tool call arguments %q: %v", args, err)
	}
	t.Logf("model %q answered: %q", model, out.Capital)
	if out.Capital == "" {
		t.Fatalf("model %q returned an empty capital field", model)
	}
}

func TestSmokeCoderClient(t *testing.T) {
	_ = godotenv.Load("../../.env")

	apiKey := os.Getenv("CODER_LLM_API_KEY")
	model := os.Getenv("CODER_LLM_MODEL")
	if apiKey == "" || model == "" {
		t.Skip("CODER_LLM_API_KEY / CODER_LLM_MODEL not set in server/.env — skipping live smoke test")
	}
	baseURL := os.Getenv("CODER_LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	client := NewClient(baseURL, apiKey, model)
	tool := Tool{
		Name:        "report_sum",
		Description: "Report the sum of two integers.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": { "sum": { "type": "integer" } },
			"required": ["sum"],
			"additionalProperties": false
		}`),
	}

	args, err := client.CallWithTool(context.Background(), []Message{
		TextMessage("user", "What is 2 + 3? Call the tool with the answer."),
	}, tool)
	if err != nil {
		t.Fatalf("coder model %q smoke test failed: %v", model, err)
	}

	var out struct {
		Sum int `json:"sum"`
	}
	if err := json.Unmarshal(args, &out); err != nil {
		t.Fatalf("could not parse tool call arguments %q: %v", args, err)
	}
	t.Logf("model %q answered: %d", model, out.Sum)
	if out.Sum != 5 {
		t.Fatalf("model %q returned %d, expected 5", model, out.Sum)
	}
}
