package analysis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vimalraj-a/mongo-hack/server/internal/models"
)

const decideActionToolName = "decide_action"

// Flat schema, every field required (empty string/enum member for "n/a")
// rather than nested optional objects or null unions — broader tool-calling
// compatibility across the assorted OpenRouter-hosted models this is meant
// to be swappable across.
var decideActionParams = json.RawMessage(`{
	"type": "object",
	"properties": {
		"actionType": {
			"type": "string",
			"enum": ["click", "fill", "select", "press", "navigate", "wait", "expect"],
			"description": "The kind of step this is."
		},
		"selector": {
			"type": "string",
			"description": "CSS selector copied verbatim from the given element inventory. Empty string if not applicable (e.g. for wait/expect/navigate)."
		},
		"value": {
			"type": "string",
			"description": "Text to type (fill), option to choose (select), key to press (press), or URL (navigate). Empty string if not applicable."
		},
		"assertionDescription": {
			"type": "string",
			"description": "For actionType=expect only: plain-English statement of what should be true. Empty string otherwise."
		},
		"playwrightCheckType": {
			"type": "string",
			"enum": ["visible", "text", "url", "count", ""],
			"description": "For actionType=expect only, if you can identify a concrete Playwright-checkable form of the assertion. Empty string otherwise."
		},
		"playwrightCheckSelector": {
			"type": "string",
			"description": "Selector for the Playwright check, from the element inventory. Empty string if not applicable."
		},
		"playwrightCheckExpectedText": {
			"type": "string",
			"description": "Expected text for a 'text' playwrightCheckType. Empty string otherwise."
		},
		"reasoning": {
			"type": "string",
			"description": "One sentence explaining the choice."
		}
	},
	"required": ["actionType", "selector", "value", "assertionDescription", "playwrightCheckType", "playwrightCheckSelector", "playwrightCheckExpectedText", "reasoning"],
	"additionalProperties": false
}`)

type decideActionArgs struct {
	ActionType                  string `json:"actionType"`
	Selector                    string `json:"selector"`
	Value                       string `json:"value"`
	AssertionDescription        string `json:"assertionDescription"`
	PlaywrightCheckType         string `json:"playwrightCheckType"`
	PlaywrightCheckSelector     string `json:"playwrightCheckSelector"`
	PlaywrightCheckExpectedText string `json:"playwrightCheckExpectedText"`
	Reasoning                   string `json:"reasoning"`
}

// DecideAction asks the vision client what concrete browser action (or,
// for a check/expectation step, what assertion) satisfies one NL step,
// grounded in the live element inventory and current screenshot. For
// fill/select steps it's explicitly told to invent a plausible sample value
// when the step doesn't specify one — that's the "agent comes up with
// sample data" requirement, not a separate phase.
func DecideAction(ctx context.Context, client *Client, stepText string, elements []models.Element, screenshotBase64 string) (models.Action, *models.Assertion, error) {
	elementsJSON, err := json.Marshal(elements)
	if err != nil {
		return models.Action{}, nil, fmt.Errorf("analysis: marshal elements: %w", err)
	}

	prompt := fmt.Sprintf(`You are driving a real browser through one step of a test flow.

Step to perform: %q

Current page's interactive-element inventory (JSON array of {tag, text, selector}) — use a selector from this list verbatim, never invent one:
%s

Decide the single concrete action that carries out this step. If the step asks you to type or select something and doesn't say exactly what value (e.g. "enter an email address"), invent a plausible sample value yourself — that is expected, not an error.

If the step is instead a check or expectation ("the page should show...", "confirm that...", "verify..."), use actionType "expect", fill in assertionDescription, and — only if you can identify a concrete way to check it — the playwrightCheck* fields. Leave every field that doesn't apply as an empty string.

The attached screenshot shows exactly what the page looks like right now.`, stepText, elementsJSON)

	tool := Tool{
		Name:        decideActionToolName,
		Description: "Report the concrete browser action (or assertion check) that carries out the given NL step.",
		Parameters:  decideActionParams,
	}

	raw, err := client.CallWithTool(ctx, []Message{ImageMessage(prompt, screenshotBase64)}, tool)
	if err != nil {
		return models.Action{}, nil, fmt.Errorf("analysis: decide action: %w", err)
	}

	var args decideActionArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return models.Action{}, nil, fmt.Errorf("analysis: parse decide_action arguments (%s): %w", raw, err)
	}

	action := models.Action{
		Type:      args.ActionType,
		Selector:  args.Selector,
		Value:     args.Value,
		Reasoning: args.Reasoning,
	}

	var assertion *models.Assertion
	if args.ActionType == "expect" {
		assertion = &models.Assertion{Description: args.AssertionDescription}
		if args.PlaywrightCheckType != "" {
			assertion.PlaywrightCheck = &models.PlaywrightCheck{
				Type:         args.PlaywrightCheckType,
				Selector:     args.PlaywrightCheckSelector,
				ExpectedText: args.PlaywrightCheckExpectedText,
			}
		}
	}

	return action, assertion, nil
}
