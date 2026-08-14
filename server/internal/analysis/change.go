package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vimalraj-a/mongo-hack/server/internal/models"
)

const reviewChangeToolName = "review_change_request"

var reviewChangeParams = json.RawMessage(`{
	"type": "object",
	"properties": {
		"responseType": {
			"type": "string",
			"enum": ["clarification", "updated"],
			"description": "'clarification' if the request doesn't clearly match anything on the actual page, or is ambiguous. 'updated' if you can justify the change against the page and are applying it."
		},
		"question": {
			"type": "string",
			"description": "One clarifying question. Only used when responseType=clarification; empty string otherwise."
		},
		"steps": {
			"type": "array",
			"items": {"type": "string"},
			"description": "The FULL updated ordered list of steps (not a diff), reflecting the requested change. Only used when responseType=updated; empty array otherwise."
		}
	},
	"required": ["responseType", "question", "steps"],
	"additionalProperties": false
}`)

type reviewChangeArgs struct {
	ResponseType string   `json:"responseType"`
	Question     string   `json:"question"`
	Steps        []string `json:"steps"`
}

// ChangeResult is the outcome of reviewing one change-request message.
type ChangeResult struct {
	Type     models.ChangeResponseType
	Question string
	NewSteps []string
}

// ReviewChangeRequest is the only path by which a TestScript's steps ever
// change. It grounds the user's freeform request against the capture's
// actual screenshot and HTML — not just the current steps — and either
// applies the change (returning the full new step list) or asks a
// clarifying question instead of guessing.
func ReviewChangeRequest(ctx context.Context, client *Client, currentSteps []string, cap models.Capture, message string) (ChangeResult, error) {
	numbered := make([]string, len(currentSteps))
	for i, s := range currentSteps {
		numbered[i] = fmt.Sprintf("%d. %s", i+1, s)
	}

	elementsJSON, err := json.Marshal(cap.Elements)
	if err != nil {
		return ChangeResult{}, fmt.Errorf("analysis: marshal elements: %w", err)
	}

	prompt := fmt.Sprintf(`You maintain the natural-language test script below for the page at %s (title: %q). The user is asking for a change in plain language — you must only apply it if you can actually justify it against the real page (the screenshot, the HTML, and the element inventory), not just because it sounds plausible.

Current script:
%s

User's requested change: %q

If the request references something that isn't actually on the page, or is too ambiguous to act on confidently, respond with responseType "clarification" and ask exactly one specific question that would let you proceed. Otherwise respond with responseType "updated" and return the complete updated step list (every step, not just the changed one) — keep it in the same one-action-per-step style as the current script.

HTML snapshot of the page:
%s

Interactive elements (JSON array of {tag, text, selector}):
%s

The attached screenshot shows exactly what the page looks like.`, cap.URL, cap.Title, strings.Join(numbered, "\n"), message, cap.HTMLSnapshot, elementsJSON)

	tool := Tool{
		Name:        reviewChangeToolName,
		Description: "Report either a clarifying question or the fully updated step list for the requested change.",
		Parameters:  reviewChangeParams,
	}

	raw, err := client.CallWithTool(ctx, []Message{ImageMessage(prompt, cap.ScreenshotBase64)}, tool)
	if err != nil {
		return ChangeResult{}, fmt.Errorf("analysis: review change request: %w", err)
	}

	var args reviewChangeArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return ChangeResult{}, fmt.Errorf("analysis: parse review_change_request arguments (%s): %w", raw, err)
	}

	result := ChangeResult{Type: models.ChangeResponseType(args.ResponseType), Question: args.Question, NewSteps: args.Steps}
	if result.Type != models.ChangeResponseClarification && result.Type != models.ChangeResponseUpdated {
		return ChangeResult{}, fmt.Errorf("analysis: model returned unexpected responseType %q", args.ResponseType)
	}
	return result, nil
}
