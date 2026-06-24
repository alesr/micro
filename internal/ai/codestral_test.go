package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodestralComplete(t *testing.T) {
	expected := "return x + 1"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req codestralRequest
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "func foo()", req.Prompt)
		assert.Equal(t, "}", req.Suffix)
		assert.Equal(t, "codestral-latest", req.Model)

		resp := codestralResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: expected}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("MISTRAL_API_KEY", "test-key")

	p, err := NewCodestralProvider("", server.URL)
	assert.NoError(t, err)

	resp, err := p.Complete(context.Background(), Request{
		BeforeCursor: "func foo()",
		AfterCursor:  "}",
	})
	assert.NoError(t, err)
	assert.Equal(t, expected, resp.Text)
}

func TestCodestralEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := codestralResponse{Choices: []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("MISTRAL_API_KEY", "test-key")

	p, err := NewCodestralProvider("", server.URL)
	assert.NoError(t, err)

	_, err = p.Complete(context.Background(), Request{
		BeforeCursor: "func foo()",
		AfterCursor:  "}",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no choices")
}

func TestCodestralNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid token"}`))
	}))
	defer server.Close()

	t.Setenv("MISTRAL_API_KEY", "test-key")

	p, err := NewCodestralProvider("", server.URL)
	assert.NoError(t, err)

	_, err = p.Complete(context.Background(), Request{
		BeforeCursor: "func foo()",
		AfterCursor:  "}",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 401")
}

func TestCodestralMissingKey(t *testing.T) {
	t.Setenv("MISTRAL_API_KEY", "")

	_, err := NewCodestralProvider("", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MISTRAL_API_KEY")
}

func TestCodestralCustomModel(t *testing.T) {
	t.Setenv("MISTRAL_API_KEY", "test-key")

	p, err := NewCodestralProvider("my-custom-model", "")
	assert.NoError(t, err)
	assert.Equal(t, "my-custom-model", p.model)
}

func TestCodestralCustomBaseURL(t *testing.T) {
	t.Setenv("MISTRAL_API_KEY", "test-key")

	p, err := NewCodestralProvider("", "https://example.com/v1/fim")
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/v1/fim", p.baseURL)
}

func TestCodestralContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer server.Close()

	t.Setenv("MISTRAL_API_KEY", "test-key")

	p, err := NewCodestralProvider("", server.URL)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = p.Complete(ctx, Request{BeforeCursor: "func"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestCodestralName(t *testing.T) {
	t.Setenv("MISTRAL_API_KEY", "test-key")

	p, err := NewCodestralProvider("", "")
	assert.NoError(t, err)
	assert.Equal(t, "codestral", p.Name())
}
