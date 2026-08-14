package capture

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// Click clicks the element matching the given CSS selector.
func (s *Session) Click(selector string) error {
	if err := chromedp.Run(s.ctx, chromedp.Click(selector, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("capture: click %q: %w", selector, err)
	}
	return nil
}

// Fill clears the element matching selector and types value into it,
// dispatching real key events so framework-bound input handlers fire.
func (s *Session) Fill(selector, value string) error {
	if err := chromedp.Run(s.ctx,
		chromedp.Clear(selector, chromedp.ByQuery),
		chromedp.SendKeys(selector, value, chromedp.ByQuery),
	); err != nil {
		return fmt.Errorf("capture: fill %q: %w", selector, err)
	}
	return nil
}

// Select sets a <select> element's value.
func (s *Session) Select(selector, value string) error {
	if err := chromedp.Run(s.ctx, chromedp.SetValue(selector, value, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("capture: select %q=%q: %w", selector, value, err)
	}
	return nil
}

// Press dispatches a key event (e.g. "Enter", "Escape") to whatever element
// currently has focus.
func (s *Session) Press(key string) error {
	if err := chromedp.Run(s.ctx, chromedp.KeyEvent(key)); err != nil {
		return fmt.Errorf("capture: press %q: %w", key, err)
	}
	return nil
}

// Navigate re-navigates the current tab to url, for "navigate"-type steps
// mid-flow (as opposed to the initial navigation done in NewSession).
func (s *Session) Navigate(url string) error {
	if err := chromedp.Run(s.ctx, chromedp.Navigate(url), chromedp.Sleep(500*time.Millisecond)); err != nil {
		return fmt.Errorf("capture: navigate to %s: %w", url, err)
	}
	return nil
}
