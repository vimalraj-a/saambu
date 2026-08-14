// Package capture wraps a headless chromedp browser session. It is used
// both for the initial prerequisite+capture flow and for verify runs — every
// verify run opens a fresh Session and re-navigates from scratch rather than
// reusing a prior one.
package capture

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	// sessionTimeout bounds a session's entire lifetime (navigate + every
	// action run against it). Individual chromedp.Run calls deliberately do
	// NOT get their own derived context: chromedp ties internal listener
	// bookkeeping to whichever context object a Run call receives, and
	// canceling a per-call child context (even after that call returns)
	// leaves later calls on the parent context failing with "context
	// canceled" — a real chromedp quirk, confirmed by isolating it before
	// writing this. One timeout for the whole session avoids it entirely.
	sessionTimeout = 3 * time.Minute
	viewportWidth  = 1440
	viewportHeight = 900
)

// Session is one live headless-browser tab, plus the allocator/context
// plumbing chromedp needs to actually spawn and tear down that browser.
type Session struct {
	ctx         context.Context
	cancelTask  context.CancelFunc
	cancelAlloc context.CancelFunc
	cancelLife  context.CancelFunc
}

// NewSession launches a headless Chrome instance and navigates to url.
// Callers must call Close when done.
func NewSession(ctx context.Context, url string) (*Session, error) {
	lifeCtx, cancelLife := context.WithTimeout(ctx, sessionTimeout)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(lifeCtx, chromedp.DefaultExecAllocatorOptions[:]...)
	taskCtx, cancelTask := chromedp.NewContext(allocCtx)

	s := &Session{ctx: taskCtx, cancelTask: cancelTask, cancelAlloc: cancelAlloc, cancelLife: cancelLife}

	if err := chromedp.Run(s.ctx,
		chromedp.EmulateViewport(viewportWidth, viewportHeight),
		chromedp.Navigate(url),
		chromedp.Sleep(1*time.Second),
	); err != nil {
		s.Close()
		return nil, fmt.Errorf("capture: navigate to %s: %w", url, err)
	}

	return s, nil
}

// Close tears down the browser. Safe to call once; subsequent calls are
// no-ops since all cancel funcs are idempotent.
func (s *Session) Close() {
	s.cancelTask()
	s.cancelAlloc()
	s.cancelLife()
}

// Title returns the current page's document.title.
func (s *Session) Title() (string, error) {
	var title string
	if err := chromedp.Run(s.ctx, chromedp.Title(&title)); err != nil {
		return "", fmt.Errorf("capture: read title: %w", err)
	}
	return title, nil
}

// Screenshot returns a base64-encoded full-page PNG of the current page.
func (s *Session) Screenshot() (string, error) {
	var buf []byte
	if err := chromedp.Run(s.ctx, chromedp.FullScreenshot(&buf, 90)); err != nil {
		return "", fmt.Errorf("capture: screenshot: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

// Wait pauses for the given duration — used for "wait"-type NL steps
// (e.g. "wait for the page to load after clicking submit").
func (s *Session) Wait(d time.Duration) error {
	return chromedp.Run(s.ctx, chromedp.Sleep(d))
}
