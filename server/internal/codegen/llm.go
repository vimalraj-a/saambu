package codegen

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vimalraj-a/mongo-hack/server/internal/analysis"
	"github.com/vimalraj-a/mongo-hack/server/internal/models"
)

const generateCodeToolName = "generate_playwright_spec"

var generateCodeParams = json.RawMessage(`{
	"type": "object",
	"properties": {
		"code": {
			"type": "string",
			"description": "The complete contents of the Playwright TypeScript spec file."
		}
	},
	"required": ["code"],
	"additionalProperties": false
}`)

type generateCodeArgs struct {
	Code string `json:"code"`
}

// traceStep is a screenshot-free projection of models.StepExecution — the
// coder model is text-only and doesn't need (or benefit from) the
// after-screenshot; leaving it out keeps the prompt small.
type traceStep struct {
	NLText    string            `json:"nlText"`
	Action    models.Action     `json:"action"`
	Assertion *models.Assertion `json:"assertion,omitempty"`
}

func toTraceSteps(steps []models.StepExecution) []traceStep {
	out := make([]traceStep, len(steps))
	for i, se := range steps {
		out[i] = traceStep{NLText: se.NLText, Action: se.Action, Assertion: se.Assertion}
	}
	return out
}

// GenerateWithLLM asks the coder-role client to transcribe an already-
// verified trace into a Playwright spec file. Its job is transcription, not
// invention: the prompt hands it the exact selectors/values/assertions and
// tells it not to alter them. Callers should run ValidateGeneratedCode on
// the result and fall back to BuildSpec if it doesn't check out.
func GenerateWithLLM(ctx context.Context, coder *analysis.Client, prereqSteps []models.StepExecution, testRun models.ExecutionRun, expectedToFail bool) (string, error) {
	payload := struct {
		PrerequisiteSteps []traceStep `json:"prerequisiteSteps"`
		TestSteps         []traceStep `json:"testSteps"`
		ExpectedToFail    bool        `json:"expectedToFail"`
	}{
		PrerequisiteSteps: toTraceSteps(prereqSteps),
		TestSteps:         toTraceSteps(testRun.Steps),
		ExpectedToFail:    expectedToFail,
	}
	traceJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("codegen: marshal trace: %w", err)
	}

	prompt := fmt.Sprintf(`Transcribe this already-verified browser action trace into a single Playwright TypeScript test using @playwright/test. Use the selectors and values EXACTLY as given below — do not invent, "improve", or modify any of them; your job is faithful transcription, not test design.

One test() block. Map actions like this:
- click -> page.locator(selector).click()
- fill -> page.locator(selector).fill(value)
- select -> page.locator(selector).selectOption(value)
- press -> page.keyboard.press(value || 'Enter')
- navigate -> page.goto(value)
- wait -> page.waitForTimeout(<value in ms, default 1500>)
- expect with assertion.playwrightCheck.type visible/text/url/count -> a real expect(...) assertion of that kind against assertion.playwrightCheck.selector / expectedText. If playwrightCheck is absent, add a comment with assertion.description instead of a check.

prerequisiteSteps run first (setup), then testSteps. If expectedToFail is true, call test.fail(); as the very first line inside the test body, with a short comment noting this encodes user-confirmed-correct behavior that the app does not currently satisfy.

Trace:
%s

Report only the file contents via the tool call.`, traceJSON)

	tool := analysis.Tool{
		Name:        generateCodeToolName,
		Description: "Report the generated Playwright spec file contents.",
		Parameters:  generateCodeParams,
	}

	raw, err := coder.CallWithTool(ctx, []analysis.Message{analysis.TextMessage("user", prompt)}, tool)
	if err != nil {
		return "", fmt.Errorf("codegen: generate with LLM: %w", err)
	}

	var args generateCodeArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "", fmt.Errorf("codegen: parse generate_playwright_spec arguments (%s): %w", raw, err)
	}
	return args.Code, nil
}
