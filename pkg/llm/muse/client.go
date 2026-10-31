// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package muse

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/samber/lo"
)

const (
	DefaultBaseURL = "https://api.meta.ai/v1"
	DefaultModel   = "muse-wonder-1.0"
)

// Client calls the Meta Muse responses API.
type Client struct {
	config Config
	http   *req.Client
}

// Message is a chat message passed to Complete.
type Message struct {
	Role    string
	Content string
}

// Usage is token usage reported by the responses API.
type Usage struct {
	InputTokens     int64
	OutputTokens    int64
	ReasoningTokens int64
	TotalTokens     int64
}

// ContentPart is a single content item of an input or output message.
type ContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// InputItem is a message in the request input.
type InputItem struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
}

// Reasoning configures the model reasoning effort.
type Reasoning struct {
	Effort string `json:"effort"`
}

// Request is the responses API request body.
type Request struct {
	Model        string      `json:"model"`
	Instructions string      `json:"instructions,omitempty"`
	Input        []InputItem `json:"input"`
	Reasoning    *Reasoning  `json:"reasoning,omitempty"`
	Stream       bool        `json:"stream"`
}

// OutputItem is an item of the response output.
type OutputItem struct {
	Type    string        `json:"type"`
	Content []ContentPart `json:"content"`
}

// ResponseUsage is the usage block of the responses API.
type ResponseUsage struct {
	InputTokens         int64 `json:"input_tokens"`
	OutputTokens        int64 `json:"output_tokens"`
	TotalTokens         int64 `json:"total_tokens"`
	OutputTokensDetails struct {
		ReasoningTokens int64 `json:"reasoning_tokens"`
	} `json:"output_tokens_details"`
}

// ResponseError is the error block of a failed response.
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Response is the responses API response body.
type Response struct {
	ID     string         `json:"id"`
	Error  *ResponseError `json:"error"`
	Output []OutputItem   `json:"output"`
	Usage  ResponseUsage  `json:"usage"`
}

// StreamEvent is one server-sent event of a streamed response.
type StreamEvent struct {
	Type     string         `json:"type"`
	Delta    string         `json:"delta"`
	Message  string         `json:"message"`
	Response *Response      `json:"response"`
	Error    *ResponseError `json:"error"`
}

// New returns a Meta Muse client.
func New(cfg Config) *Client {
	cfg.BaseURL = lo.CoalesceOrEmpty(cfg.BaseURL, DefaultBaseURL)
	cfg.Model = lo.CoalesceOrEmpty(cfg.Model, DefaultModel)

	return &Client{
		config: cfg,
		http: req.C().
			SetBaseURL(cfg.BaseURL).
			SetCommonBearerAuthToken(cfg.APIKey),
	}
}

// Complete sends messages and returns the assistant reply and usage.
func (c *Client) Complete(ctx context.Context, messages []Message) (string, Usage, error) {
	response, err := c.Create(ctx, c.BuildRequest(messages))
	if err != nil {
		return "", Usage{}, err
	}

	usage := response.TokenUsage()

	text := response.OutputText()
	if lo.IsEmpty(text) {
		return "", usage, ErrNoCompletion
	}

	return text, usage, nil
}

// Stream sends messages, passes each piece of the reply to onDelta as it arrives and returns the whole reply and usage.
func (c *Client) Stream(ctx context.Context, messages []Message, onDelta func(string)) (string, Usage, error) {
	request := c.BuildRequest(messages)
	request.Stream = true

	resp, err := c.http.R().
		SetContext(ctx).
		SetBody(request).
		DisableAutoReadResponse().
		Post("/responses")
	if err != nil {
		return "", Usage{}, fmt.Errorf("muse request: %w", err)
	}

	defer resp.Body.Close()

	if !resp.IsSuccessState() {
		return "", Usage{}, fmt.Errorf("muse request: status %d", resp.StatusCode)
	}

	var reply strings.Builder
	usage := Usage{}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)

	for scanner.Scan() {
		data, ok := strings.CutPrefix(scanner.Text(), "data: ")
		if !ok || data == "[DONE]" {
			continue
		}

		var event StreamEvent
		if json.Unmarshal([]byte(data), &event) != nil {
			continue
		}

		switch event.Type {
		case "response.output_text.delta":
			reply.WriteString(event.Delta)
			onDelta(event.Delta)
		case "response.completed":
			usage = event.Response.TokenUsage()
		case "response.failed":
			return "", Usage{}, fmt.Errorf("muse response: %s", event.Response.Error.Message)
		case "error":
			return "", Usage{}, fmt.Errorf("muse response: %s", event.Message)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", Usage{}, fmt.Errorf("muse stream: %w", err)
	}

	if reply.Len() == 0 {
		return "", usage, ErrNoCompletion
	}

	return reply.String(), usage, nil
}

// BuildRequest converts chat messages into a responses API request, joining system messages into its instructions.
func (c *Client) BuildRequest(messages []Message) Request {
	request := Request{Model: c.config.Model}
	instructions := []string{}

	for _, message := range messages {
		switch message.Role {
		case "system":
			instructions = append(instructions, message.Content)
		case "assistant":
			request.Input = append(request.Input, InputItem{
				Role:    "assistant",
				Content: []ContentPart{{Type: "output_text", Text: message.Content}},
			})
		default:
			request.Input = append(request.Input, InputItem{
				Role:    "user",
				Content: []ContentPart{{Type: "input_text", Text: message.Content}},
			})
		}
	}

	request.Instructions = strings.Join(instructions, "\n\n")

	if lo.IsNotEmpty(c.config.Effort) {
		request.Reasoning = &Reasoning{Effort: c.config.Effort}
	}

	return request
}

// Create sends a raw request to the responses API.
func (c *Client) Create(ctx context.Context, request Request) (*Response, error) {
	var result Response

	resp, err := c.http.R().
		SetContext(ctx).
		SetBody(request).
		SetSuccessResult(&result).
		Post("/responses")
	if err != nil {
		return nil, fmt.Errorf("muse request: %w", err)
	}

	if !resp.IsSuccessState() {
		return nil, fmt.Errorf("muse request: status %d: %s", resp.StatusCode, resp.Bytes())
	}

	if result.Error != nil {
		return nil, fmt.Errorf("muse response %s: %s: %s", result.ID, result.Error.Code, result.Error.Message)
	}

	return &result, nil
}

// OutputText returns the concatenated text of the assistant output messages.
func (r *Response) OutputText() string {
	var builder strings.Builder

	for _, item := range r.Output {
		if item.Type != "message" {
			continue
		}

		for _, part := range item.Content {
			if part.Type == "output_text" {
				builder.WriteString(part.Text)
			}
		}
	}

	return builder.String()
}

// TokenUsage returns the response's token usage.
func (r *Response) TokenUsage() Usage {
	return Usage{
		InputTokens:     r.Usage.InputTokens,
		OutputTokens:    r.Usage.OutputTokens,
		ReasoningTokens: r.Usage.OutputTokensDetails.ReasoningTokens,
		TotalTokens:     r.Usage.TotalTokens,
	}
}
