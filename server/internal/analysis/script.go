package analysis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vimalraj-a/mongo-hack/server/internal/models"
)

const generateScriptToolName = "generate_test_script"

var generateScriptParams = json.RawMessage(`{
	"type": "object",
	"properties": {
		"steps": {
			"type": "array",
			"items": {"type": "string"},
			"description": "Ordered, discrete NL steps describing a test flow for this page — each one action or one check, no compound steps."
		}
	},
	"required": ["steps"],
	"additionalProperties": false
}`)

type generateScriptArgs struct {
	Steps []string `json:"steps"`
}

// GenerateNLScript looks at a capture's screenshot and element inventory and
// proposes a natural-language test script — the editable "spec" the user
// subsequently refines only through mediated change requests.
func GenerateNLScript(ctx context.Context, client *Client, cap models.Capture) ([]string, error) {
	elementsJSON, err := json.Marshal(cap.Elements)
	if err != nil {
		return nil, fmt.Errorf("analysis: marshal elements: %w", err)
	}

	prompt := fmt.Sprintf(`You are looking at a captured web page (title: %q, url: %s). Propose a natural-language test script: an ordered list of steps that exercises one clear, worthwhile user flow on this page (e.g. filling in and submitting a visible form, following a primary call-to-action, or checking that key content is present).

Write each step as one discrete action or check ("Click the 'Sign in' button", "Enter a valid email in the Email field", "Expect a welcome message to appear") — no compound steps. Keep it focused: 3-8 steps is typical.

Interactive elements on the page (JSON array of {tag, text, selector}):
%s

The attached screenshot shows exactly what the page looks like.`, cap.Title, cap.URL, elementsJSON)

	tool := Tool{
		Name:        generateScriptToolName,
		Description: "Report the proposed natural-language test script for this page.",
		Parameters:  generateScriptParams,
	}

	raw, err := client.CallWithTool(ctx, []Message{ImageMessage(prompt, cap.ScreenshotBase64)}, tool)
	if err != nil {
		return nil, fmt.Errorf("analysis: generate script: %w", err)
	}

	var args generateScriptArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, fmt.Errorf("analysis: parse generate_test_script arguments (%s): %w", raw, err)
	}
	return args.Steps, nil
}
