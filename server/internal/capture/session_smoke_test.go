package capture

import (
	"context"
	"testing"
)

// Live smoke test — actually launches Chrome. Skippable via -short if it
// ever gets slow in CI, but there's no CI here yet so it always runs.
func TestSmokeSession(t *testing.T) {
	s, err := NewSession(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer s.Close()

	title, err := s.Title()
	if err != nil {
		t.Fatalf("Title: %v", err)
	}
	t.Logf("title: %q", title)
	if title == "" {
		t.Fatal("expected a non-empty title")
	}

	shot, err := s.Screenshot()
	if err != nil {
		t.Fatalf("Screenshot: %v", err)
	}
	if len(shot) < 100 {
		t.Fatalf("screenshot base64 looks too short: %d chars", len(shot))
	}
	t.Logf("screenshot: %d base64 chars", len(shot))

	html, err := s.ExtractHTML()
	if err != nil {
		t.Fatalf("ExtractHTML: %v", err)
	}
	if html == "" {
		t.Fatal("expected non-empty HTML")
	}
	t.Logf("html: %d chars", len(html))

	elements, err := s.ExtractElements()
	if err != nil {
		t.Fatalf("ExtractElements: %v", err)
	}
	t.Logf("found %d interactive elements", len(elements))
	for _, el := range elements {
		t.Logf("  %s %q -> %s", el.Tag, el.Text, el.Selector)
	}
}
