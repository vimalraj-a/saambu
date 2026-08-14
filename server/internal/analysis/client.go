// Package analysis holds every LLM call in the app, all going through the
// same OpenAI-compatible chat/completions wire format so the provider and
// model for each role (vision vs. coder) is a config change, not a code
// change — see server/.env.example.
package analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to one OpenAI-compatible endpoint for one role (vision or
// coder). Each role gets its own Client instance.
type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}
}

// Message mirrors the OpenAI chat/completions message shape. Content is
// either a plain string or a []ContentPart for multimodal (vision) turns.
type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type ContentPart struct {
	Type     string    `json:"type"` // "text" | "image_url"
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

func TextMessage(role, text string) Message {
	return Message{Role: role, Content: text}
}

// ImageMessage builds a user turn carrying text plus a screenshot, as a
// data: URL content part.
func ImageMessage(text string, screenshotPNGBase64 string) Message {
	return Message{
		Role: "user",
		Content: []ContentPart{
			{Type: "text", Text: text},
			{Type: "image_url", ImageURL: &ImageURL{URL: "data:image/png;base64," + screenshotPNGBase64}},
		},
	}
}

// Tool is one OpenAI-compatible function-tool definition. Parameters must be
// a JSON Schema object.
type Tool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

type toolWire struct {
	Type     string       `json:"type"`
	Function toolFuncWire `json:"function"`
}

type toolFuncWire struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type chatRequest struct {
	Model      string     `json:"model"`
	Messages   []Message  `json:"messages"`
	Tools      []toolWire `json:"tools"`
	ToolChoice any        `json:"tool_choice"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// CallWithTool sends messages with a single forced tool call and returns the
// tool call's parsed arguments. Structured output goes through tool calling
// rather than response_format:json_schema, since tool calling is far more
// broadly supported across OpenRouter-hosted open models.
func (c *Client) CallWithTool(ctx context.Context, messages []Message, tool Tool) (json.RawMessage, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("analysis: no API key configured for model %q — set it in server/.env", c.model)
	}
	if c.model == "" {
		return nil, fmt.Errorf("analysis: no model configured — set the corresponding *_LLM_MODEL in server/.env")
	}

	reqBody := chatRequest{
		Model:    c.model,
		Messages: messages,
		Tools: []toolWire{{
			Type: "function",
			Function: toolFuncWire{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		}},
		ToolChoice: map[string]any{
			"type":     "function",
			"function": map[string]string{"name": tool.Name},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("analysis: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("analysis: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("analysis: request to %s failed: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("analysis: read response: %w", err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("analysis: unmarshal response (status %d): %w; body: %s", resp.StatusCode, err, truncate(respBody, 500))
	}

	if parsed.Error != nil {
		return nil, fmt.Errorf("analysis: %s returned an error for model %q: %s", c.baseURL, c.model, parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("analysis: %s returned status %d for model %q: %s", c.baseURL, resp.StatusCode, c.model, truncate(respBody, 500))
	}
	if len(parsed.Choices) == 0 || len(parsed.Choices[0].Message.ToolCalls) == 0 {
		content := ""
		if len(parsed.Choices) > 0 {
			content = parsed.Choices[0].Message.Content
		}
		return nil, fmt.Errorf("analysis: model %q did not return a tool call (it may not support forced tool calling); response text: %s", c.model, truncate([]byte(content), 500))
	}

	return json.RawMessage(parsed.Choices[0].Message.ToolCalls[0].Function.Arguments), nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
