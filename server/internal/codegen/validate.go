package codegen

import (
	"fmt"
	"strings"

	"github.com/vimalraj-a/mongo-hack/server/internal/models"
)

// ValidateGeneratedCode is the guardrail against a hallucinated transcription:
// every non-empty selector or fill/select value that appears in the verified
// trace must appear somewhere in the generated code, or the output isn't
// trusted (the caller should fall back to BuildSpec instead).
func ValidateGeneratedCode(code string, prereqSteps []models.StepExecution, testRun models.ExecutionRun) error {
	all := make([]models.StepExecution, 0, len(prereqSteps)+len(testRun.Steps))
	all = append(all, prereqSteps...)
	all = append(all, testRun.Steps...)

	for _, se := range all {
		if se.Action.Selector != "" && !strings.Contains(code, se.Action.Selector) {
			return fmt.Errorf("generated code is missing selector %q from the verified trace (step %q)", se.Action.Selector, se.NLText)
		}
		if (se.Action.Type == "fill" || se.Action.Type == "select") && se.Action.Value != "" && !strings.Contains(code, se.Action.Value) {
			return fmt.Errorf("generated code is missing value %q from the verified trace (step %q)", se.Action.Value, se.NLText)
		}
		if se.Assertion != nil && se.Assertion.PlaywrightCheck != nil && se.Assertion.PlaywrightCheck.Selector != "" {
			if !strings.Contains(code, se.Assertion.PlaywrightCheck.Selector) {
				return fmt.Errorf("generated code is missing assertion selector %q from the verified trace (step %q)", se.Assertion.PlaywrightCheck.Selector, se.NLText)
			}
		}
	}
	return nil
}
