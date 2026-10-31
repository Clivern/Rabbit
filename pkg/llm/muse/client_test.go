// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package muse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitMuseClient(t *testing.T) {
	t.Run("New defaults", func(t *testing.T) {
		client := New(Config{APIKey: "key"})
		assert.Equal(t, DefaultBaseURL, client.config.BaseURL)
		assert.Equal(t, DefaultModel, client.config.Model)
	})

	t.Run("BuildRequest", func(t *testing.T) {
		client := New(Config{APIKey: "key", Effort: "low"})
		request := client.BuildRequest([]Message{
			{Role: "system", Content: "Be terse."},
			{Role: "user", Content: "hi"},
			{Role: "assistant", Content: "hello"},
		})

		assert.Equal(t, "Be terse.", request.Instructions)
		assert.Equal(t, "low", request.Reasoning.Effort)
		assert.Len(t, request.Input, 2)
		assert.Equal(t, "input_text", request.Input[0].Content[0].Type)
		assert.Equal(t, "output_text", request.Input[1].Content[0].Type)
	})

	t.Run("Complete", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/responses", r.URL.Path)
				assert.Equal(t, "Bearer key", r.Header.Get("Authorization"))

				var request Request
				json.NewDecoder(r.Body).Decode(&request)
				assert.Equal(t, DefaultModel, request.Model)

				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"id": "resp_1",
					"object": "response",
					"status": "completed",
					"reasoning": {"effort": "max"},
					"instructions": null,
					"error": null,
					"incomplete_details": null,
					"output": [
						{"type": "reasoning", "summary": []},
						{"type": "message", "role": "assistant", "content": [{"type": "output_text", "text": "Hi"}]}
					],
					"usage": {"input_tokens": 18, "input_tokens_details": {"cached_tokens": 4}, "output_tokens": 70, "output_tokens_details": {"reasoning_tokens": 60}, "total_tokens": 88}
				}`))
			}),
		)
		defer server.Close()

		client := New(Config{APIKey: "key", BaseURL: server.URL})
		text, usage, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "Say hi"}})

		assert.NoError(t, err)
		assert.Equal(t, "Hi", text)
		assert.Equal(t, int64(60), usage.ReasoningTokens)
		assert.Equal(t, int64(88), usage.TotalTokens)
	})
	t.Run("Stream", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request Request
				json.NewDecoder(r.Body).Decode(&request)
				assert.True(t, request.Stream)

				w.Header().Set("Content-Type", "text/event-stream")
				w.Write([]byte("event: response.output_text.delta\n" +
					`data: {"type":"response.output_text.delta","delta":"One"}` + "\n\n" +
					"event: response.output_text.delta\n" +
					`data: {"type":"response.output_text.delta","delta":", two"}` + "\n\n" +
					"event: response.completed\n" +
					`data: {"type":"response.completed","response":{"usage":{"input_tokens":5,"output_tokens":7,"total_tokens":12}}}` + "\n\n" +
					"data: [DONE]\n\n"))
			}),
		)
		defer server.Close()

		var deltas []string
		client := New(Config{APIKey: "key", BaseURL: server.URL})
		text, usage, err := client.Stream(context.Background(), []Message{{Role: "user", Content: "Count"}}, func(delta string) {
			deltas = append(deltas, delta)
		})

		assert.NoError(t, err)
		assert.Equal(t, "One, two", text)
		assert.Equal(t, []string{"One", ", two"}, deltas)
		assert.Equal(t, int64(12), usage.TotalTokens)
	})

	t.Run("Stream failed", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.Write([]byte(`data: {"type":"error","message":"rate limited"}` + "\n\n"))
			}),
		)
		defer server.Close()

		client := New(Config{APIKey: "key", BaseURL: server.URL})
		_, _, err := client.Stream(context.Background(), []Message{{Role: "user", Content: "Count"}}, func(string) {})

		assert.ErrorContains(t, err, "rate limited")
	})
}
