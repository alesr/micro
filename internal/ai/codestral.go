package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	codestralDefaultBaseURL = "https://api.mistral.ai/v1/fim/completions"
	codestralDefaultModel   = "codestral-latest"
	codestralEnvKey         = "MISTRAL_API_KEY"
)

type CodestralProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewCodestralProvider(model, baseURL string) (*CodestralProvider, error) {
	apiKey := os.Getenv(codestralEnvKey)
	if apiKey == "" {
		return nil, fmt.Errorf("codestral: %s environment variable not set", codestralEnvKey)
	}
	if model == "" {
		model = codestralDefaultModel
	}
	if baseURL == "" {
		baseURL = codestralDefaultBaseURL
	}
	return &CodestralProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (p *CodestralProvider) Name() string {
	return "codestral"
}

type codestralRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Suffix      string   `json:"suffix"`
	MaxTokens   int      `json:"max_tokens"`
	Temperature float64  `json:"temperature"`
	Stop        []string `json:"stop,omitempty"`
}

type codestralResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p *CodestralProvider) Complete(ctx context.Context, req Request) (*Response, error) {
	body := codestralRequest{
		Model:       p.model,
		Prompt:      req.BeforeCursor,
		Suffix:      req.AfterCursor,
		MaxTokens:   256,
		Temperature: 0,
		Stop:        []string{"\n\n"},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("codestral: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("codestral: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("codestral: http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("codestral: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("codestral: status %d: %s", resp.StatusCode, string(respBody))
	}

	var cresp codestralResponse
	if err := json.Unmarshal(respBody, &cresp); err != nil {
		return nil, fmt.Errorf("codestral: parse response: %w", err)
	}

	if len(cresp.Choices) == 0 {
		return nil, fmt.Errorf("codestral: no choices in response")
	}

	return &Response{Text: cresp.Choices[0].Message.Content}, nil
}
