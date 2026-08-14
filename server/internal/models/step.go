package models

// Action is the concrete browser action the analysis LLM decided on for one
// NL step, grounded in the live element inventory at decision time.
type Action struct {
	Type      string `bson:"type" json:"type"`
	Selector  string `bson:"selector,omitempty" json:"selector,omitempty"`
	Value     string `bson:"value,omitempty" json:"value,omitempty"`
	Reasoning string `bson:"reasoning,omitempty" json:"reasoning,omitempty"`
}

// PlaywrightCheck is the concrete Playwright-checkable shape the LLM proposed
// for an assertion, so codegen can render a real `expect(...)` call instead
// of just a comment.
type PlaywrightCheck struct {
	Type         string `bson:"type" json:"type"`
	Selector     string `bson:"selector,omitempty" json:"selector,omitempty"`
	ExpectedText string `bson:"expectedText,omitempty" json:"expectedText,omitempty"`
}

// Assertion is the recorded outcome of judging an "expect"-type step against
// the after-screenshot.
type Assertion struct {
	Description     string           `bson:"description" json:"description"`
	Held            bool             `bson:"held" json:"held"`
	Explanation     string           `bson:"explanation,omitempty" json:"explanation,omitempty"`
	PlaywrightCheck *PlaywrightCheck `bson:"playwrightCheck,omitempty" json:"playwrightCheck,omitempty"`
}

// StepExecution is one step's outcome inside a run of the orchestrator's
// RunSteps loop — used both for the prerequisite setup run (embedded on
// Capture) and the test-verification run (embedded on ExecutionRun).
type StepExecution struct {
	Index                 int        `bson:"index" json:"index"`
	NLText                string     `bson:"nlText" json:"nlText"`
	Action                Action     `bson:"action" json:"action"`
	ScreenshotAfterBase64 string     `bson:"screenshotAfterBase64,omitempty" json:"screenshotAfterBase64,omitempty"`
	Assertion             *Assertion `bson:"assertion,omitempty" json:"assertion,omitempty"`
	Error                 string     `bson:"error,omitempty" json:"error,omitempty"`
}
