package analysis

import (
	"context"
	"encoding/json"
	"fmt"
)

const splitInstructionsToolName = "split_instructions"

var splitInstructionsParams = json.RawMessage(`{
	"type": "object",
	"properties": {
		"steps": {
			"type": "array",
			"items": {"type": "string"},
			"description": "Ordered, discrete, single-action steps that together carry out the instructions."
		}
	},
	"required": ["steps"],
	"additionalProperties": false
}`)

type splitInstructionsArgs struct {
	Steps []string `json:"steps"`
}

// SplitInstructions turns a user's freeform prerequisite text ("log in with
// test@example.com / password123, then open Settings") into an ordered list
// of discrete NL steps, so it can run through the exact same execution loop
// as the test script itself.
func SplitInstructions(ctx context.Context, client *Client, text string) ([]string, error) {
	prompt := fmt.Sprintf(`Break the following instructions into an ordered list of discrete, single-action steps — one action per step (one click, one field fill, one navigation, etc.), in the order they should happen. Don't add steps that weren't implied by the text; don't merge multiple actions into one step.

Instructions: %q`, text)

	tool := Tool{
		Name:        splitInstructionsToolName,
		Description: "Report the ordered list of discrete steps that carry out the given instructions.",
		Parameters:  splitInstructionsParams,
	}

	raw, err := client.CallWithTool(ctx, []Message{TextMessage("user", prompt)}, tool)
	if err != nil {
		return nil, fmt.Errorf("analysis: split instructions: %w", err)
	}

	var args splitInstructionsArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, fmt.Errorf("analysis: parse split_instructions arguments (%s): %w", raw, err)
	}
	return args.Steps, nil
}
