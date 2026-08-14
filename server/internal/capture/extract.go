package capture

import (
	"encoding/json"
	"fmt"

	"github.com/chromedp/chromedp"

	"github.com/vimalraj-a/mongo-hack/server/internal/models"
)

// ExtractHTML returns a trimmed snapshot of the current page's body HTML,
// used (alongside the screenshot) to ground change-request review.
func (s *Session) ExtractHTML() (string, error) {
	var html string
	if err := chromedp.Run(s.ctx, chromedp.Evaluate(extractHTMLJS, &html)); err != nil {
		return "", fmt.Errorf("capture: extract html: %w", err)
	}
	return html, nil
}

// ExtractElements walks the current DOM for interactive elements and
// returns, for each, the best selector candidate — in priority order
// data-testid > id > aria-label > name > a structural nth-of-type CSS path.
// Every selector produced here is plain CSS, valid for both chromedp's
// querySelector-based actions (Click/Fill/Select below) and a Playwright
// page.locator() call in the generated test — the same string works in both
// places, which is exactly what makes a verified trace trustworthy codegen
// input.
func (s *Session) ExtractElements() ([]models.Element, error) {
	var raw string
	if err := chromedp.Run(s.ctx, chromedp.Evaluate(extractElementsJS, &raw)); err != nil {
		return nil, fmt.Errorf("capture: extract elements: %w", err)
	}

	var elements []models.Element
	if err := json.Unmarshal([]byte(raw), &elements); err != nil {
		return nil, fmt.Errorf("capture: parse elements JSON: %w", err)
	}
	return elements, nil
}

const extractHTMLJS = `
(function() {
  var html = document.body ? document.body.outerHTML : document.documentElement.outerHTML;
  var MAX = 20000;
  return html.length > MAX ? html.slice(0, MAX) : html;
})()
`

const extractElementsJS = `
JSON.stringify((function() {
  function esc(v) {
    return String(v).replace(/\\/g, '\\\\').replace(/"/g, '\\"');
  }

  function cssPath(el) {
    var path = [];
    while (el && el.nodeType === Node.ELEMENT_NODE && el !== document.body) {
      var selector = el.tagName.toLowerCase();
      var sibling = el, nth = 1;
      while ((sibling = sibling.previousElementSibling)) {
        if (sibling.tagName === el.tagName) nth++;
      }
      selector += ':nth-of-type(' + nth + ')';
      path.unshift(selector);
      el = el.parentElement;
    }
    path.unshift('body');
    return path.join(' > ');
  }

  function bestSelector(el) {
    var testId = el.getAttribute('data-testid');
    if (testId) return '[data-testid="' + esc(testId) + '"]';
    if (el.id) return '[id="' + esc(el.id) + '"]';
    var ariaLabel = el.getAttribute('aria-label');
    if (ariaLabel) return el.tagName.toLowerCase() + '[aria-label="' + esc(ariaLabel) + '"]';
    var name = el.getAttribute('name');
    if (name) return el.tagName.toLowerCase() + '[name="' + esc(name) + '"]';
    return cssPath(el);
  }

  var selectors = ['a', 'button', 'input', 'textarea', 'select', '[role="button"]'];
  var nodes = document.querySelectorAll(selectors.join(','));
  var out = [];

  nodes.forEach(function(el) {
    var rect = el.getBoundingClientRect();
    if (rect.width === 0 || rect.height === 0) return;
    var style = window.getComputedStyle(el);
    if (style.visibility === 'hidden' || style.display === 'none') return;

    var text = (el.innerText || el.value || el.getAttribute('placeholder') || '').trim().slice(0, 120);

    out.push({
      tag: el.tagName.toLowerCase(),
      text: text,
      selector: bestSelector(el)
    });
  });

  return out.length > 200 ? out.slice(0, 200) : out;
})())
`
