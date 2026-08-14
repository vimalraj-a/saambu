// Package agent contains the one execution loop shared by prerequisite
// setup and test verification.
package agent

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/vimalraj-a/mongo-hack/server/internal/analysis"
	"github.com/vimalraj-a/mongo-hack/server/internal/capture"
	"github.com/vimalraj-a/mongo-hack/server/internal/models"
)

// RunSteps drives session through steps one at a time: for each step it asks
// visionClient to decide a concrete action from the live element inventory
// and current screenshot, executes that action, and — only when
// judgeAssertions is true and the step resolved to an "expect" action —
// judges whether the resulting screenshot satisfies the described
// expectation. It stops at the first unresolvable action or failed
// assertion.
//
// Used for both the prerequisite setup run (judgeAssertions=false — setup
// isn't a test, so nothing there can "fail") and test verification
// (judgeAssertions=true).
func RunSteps(ctx context.Context, session *capture.Session, visionClient *analysis.Client, steps []string, judgeAssertions bool) ([]models.StepExecution, bool) {
	var executions []models.StepExecution
	overallPassed := true

	for i, stepText := range steps {
		elements, err := session.ExtractElements()
		if err != nil {
			executions = append(executions, models.StepExecution{Index: i, NLText: stepText, Error: fmt.Sprintf("extract elements: %v", err)})
			overallPassed = false
			break
		}

		beforeShot, err := session.Screenshot()
		if err != nil {
			executions = append(executions, models.StepExecution{Index: i, NLText: stepText, Error: fmt.Sprintf("screenshot: %v", err)})
			overallPassed = false
			break
		}

		action, assertion, err := analysis.DecideAction(ctx, visionClient, stepText, elements, beforeShot)
		if err != nil {
			executions = append(executions, models.StepExecution{Index: i, NLText: stepText, Error: fmt.Sprintf("decide action: %v", err)})
			overallPassed = false
			break
		}

		if err := executeAction(session, action); err != nil {
			executions = append(executions, models.StepExecution{Index: i, NLText: stepText, Action: action, Error: fmt.Sprintf("execute action: %v", err)})
			overallPassed = false
			break
		}

		// Best-effort — an empty after-screenshot still leaves a usable
		// (if incomplete) storyboard rather than aborting the whole run.
		afterShot, _ := session.Screenshot()

		se := models.StepExecution{Index: i, NLText: stepText, Action: action, ScreenshotAfterBase64: afterShot}

		if action.Type == "expect" && assertion != nil {
			if judgeAssertions {
				held, explanation, jerr := analysis.JudgeAssertion(ctx, visionClient, assertion.Description, afterShot)
				if jerr != nil {
					se.Error = fmt.Sprintf("judge assertion: %v", jerr)
					executions = append(executions, se)
					overallPassed = false
					break
				}
				assertion.Held = held
				assertion.Explanation = explanation
				se.Assertion = assertion
				executions = append(executions, se)
				if !held {
					overallPassed = false
					break
				}
				continue
			}
			se.Assertion = assertion
		}

		executions = append(executions, se)
	}

	return executions, overallPassed
}

func executeAction(session *capture.Session, action models.Action) error {
	switch action.Type {
	case "click":
		if action.Selector == "" {
			return fmt.Errorf("click action missing a selector")
		}
		return session.Click(action.Selector)
	case "fill":
		if action.Selector == "" {
			return fmt.Errorf("fill action missing a selector")
		}
		return session.Fill(action.Selector, action.Value)
	case "select":
		if action.Selector == "" {
			return fmt.Errorf("select action missing a selector")
		}
		return session.Select(action.Selector, action.Value)
	case "press":
		key := action.Value
		if key == "" {
			key = "Enter"
		}
		return session.Press(key)
	case "navigate":
		if action.Value == "" {
			return fmt.Errorf("navigate action missing a URL")
		}
		return session.Navigate(action.Value)
	case "wait":
		d := 1500 * time.Millisecond
		if action.Value != "" {
			if secs, err := strconv.ParseFloat(action.Value, 64); err == nil {
				d = time.Duration(secs * float64(time.Second))
			}
		}
		return session.Wait(d)
	case "expect":
		return nil // handled by assertion judging in RunSteps, not a DOM action
	default:
		return fmt.Errorf("unknown action type %q", action.Type)
	}
}
