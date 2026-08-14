package analysis

import (
	"context"
	"encoding/json"
	"fmt"
)

const judgeAssertionToolName = "judge_assertion"

var judgeAssertionParams = json.RawMessage(`{
	"type": "object",
	"properties": {
		"held": {
			"type": "boolean",
			"description": "Whether the screenshot satisfies the expectation."
		},
		"explanation": {
			"type": "string",
			"description": "One sentence: what you actually see and why it does or doesn't satisfy the expectation."
		}
	},
	"required": ["held", "explanation"],
	"additionalProperties": false
}`)

type judgeAssertionArgs struct {
	Held        bool   `json:"held"`
	Explanation string `json:"explanation"`
}

// JudgeAssertion asks the vision client whether the given after-screenshot
// satisfies a plain-English expectation. Only used during test verification
// (judgeAssertions=true in the orchestrator) — prerequisite setup steps
// don't carry assertions.
func JudgeAssertion(ctx context.Context, client *Client, description string, screenshotBase64 string) (bool, string, error) {
	prompt := fmt.Sprintf(`Look at the attached screenshot, taken right after an action on a web page. Does it satisfy this expectation?

Expectation: %q

Answer strictly based on what is visible in the screenshot — don't assume anything you can't actually see.`, description)

	tool := Tool{
		Name:        judgeAssertionToolName,
		Description: "Report whether the screenshot satisfies the given expectation.",
		Parameters:  judgeAssertionParams,
	}

	raw, err := client.CallWithTool(ctx, []Message{ImageMessage(prompt, screenshotBase64)}, tool)
	if err != nil {
		return false, "", fmt.Errorf("analysis: judge assertion: %w", err)
	}

	var args judgeAssertionArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return false, "", fmt.Errorf("analysis: parse judge_assertion arguments (%s): %w", raw, err)
	}
	return args.Held, args.Explanation, nil
}
