// Package codegen turns a verified execution trace into a Playwright
// TypeScript spec file. GenerateWithLLM (llm.go) is the primary path;
// BuildSpec here is the deterministic, no-LLM fallback used whenever the
// LLM output fails ValidateGeneratedCode (validate.go) — a bad generation
// should never reach the user silently.
package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vimalraj-a/mongo-hack/server/internal/models"
)

// BuildSpec deterministically renders prereqSteps (setup) followed by
// testRun.Steps (the test itself) into one Playwright test() block.
func BuildSpec(prereqSteps []models.StepExecution, testRun models.ExecutionRun, expectedToFail bool) string {
	var b strings.Builder

	b.WriteString("import { test, expect } from '@playwright/test';\n\n")
	b.WriteString(fmt.Sprintf("test('%s', async ({ page }) => {\n", testName(testRun)))

	if expectedToFail {
		b.WriteString("  // Encodes the user-confirmed-correct expected behavior — the app does not\n")
		b.WriteString("  // currently satisfy it. Remove test.fail() once the underlying bug is fixed.\n")
		b.WriteString("  test.fail();\n\n")
	}

	if len(prereqSteps) > 0 {
		b.WriteString("  // Prerequisite setup\n")
		for _, se := range prereqSteps {
			writeStep(&b, se)
		}
		b.WriteString("\n")
	}

	b.WriteString("  // Test steps\n")
	for _, se := range testRun.Steps {
		writeStep(&b, se)
	}

	b.WriteString("});\n")
	return b.String()
}

func testName(run models.ExecutionRun) string {
	if len(run.Steps) == 0 {
		return "generated test"
	}
	return jsEscape(run.Steps[0].NLText)
}

func writeStep(b *strings.Builder, se models.StepExecution) {
	b.WriteString(fmt.Sprintf("  // %s\n", se.NLText))
	switch se.Action.Type {
	case "click":
		b.WriteString(fmt.Sprintf("  await page.locator('%s').click();\n", jsEscape(se.Action.Selector)))
	case "fill":
		b.WriteString(fmt.Sprintf("  await page.locator('%s').fill('%s');\n", jsEscape(se.Action.Selector), jsEscape(se.Action.Value)))
	case "select":
		b.WriteString(fmt.Sprintf("  await page.locator('%s').selectOption('%s');\n", jsEscape(se.Action.Selector), jsEscape(se.Action.Value)))
	case "press":
		key := se.Action.Value
		if key == "" {
			key = "Enter"
		}
		b.WriteString(fmt.Sprintf("  await page.keyboard.press('%s');\n", jsEscape(key)))
	case "navigate":
		b.WriteString(fmt.Sprintf("  await page.goto('%s');\n", jsEscape(se.Action.Value)))
	case "wait":
		ms := 1500
		if se.Action.Value != "" {
			if secs, err := strconv.ParseFloat(se.Action.Value, 64); err == nil {
				ms = int(secs * 1000)
			}
		}
		b.WriteString(fmt.Sprintf("  await page.waitForTimeout(%d);\n", ms))
	case "expect":
		writeAssertion(b, se.Assertion)
	}
}

func writeAssertion(b *strings.Builder, assertion *models.Assertion) {
	if assertion == nil {
		return
	}
	check := assertion.PlaywrightCheck
	if check == nil {
		b.WriteString(fmt.Sprintf("  // Expect: %s (no concrete Playwright check identified)\n", assertion.Description))
		return
	}
	switch check.Type {
	case "visible":
		b.WriteString(fmt.Sprintf("  await expect(page.locator('%s')).toBeVisible();\n", jsEscape(check.Selector)))
	case "text":
		if check.Selector != "" {
			b.WriteString(fmt.Sprintf("  await expect(page.locator('%s')).toContainText('%s');\n", jsEscape(check.Selector), jsEscape(check.ExpectedText)))
		} else {
			b.WriteString(fmt.Sprintf("  await expect(page.locator('body')).toContainText('%s');\n", jsEscape(check.ExpectedText)))
		}
	case "url":
		b.WriteString(fmt.Sprintf("  expect(page.url()).toContain('%s');\n", jsEscape(check.ExpectedText)))
	case "count":
		count := "0"
		if n, err := strconv.Atoi(check.ExpectedText); err == nil {
			count = strconv.Itoa(n)
		}
		b.WriteString(fmt.Sprintf("  await expect(page.locator('%s')).toHaveCount(%s);\n", jsEscape(check.Selector), count))
	default:
		b.WriteString(fmt.Sprintf("  // Expect: %s\n", assertion.Description))
	}
}

func jsEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
