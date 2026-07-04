package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OllamaClient struct {
	baseURL string
	client  *http.Client
}

func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 90 * time.Second},
	}
}

func (c *OllamaClient) Chat(ctx context.Context, request ChatRequest) (ChatResponse, error) {
	payload := map[string]any{
		"model":  request.Model,
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": request.SystemPrompt},
			{"role": "user", "content": request.UserInput},
		},
	}

	var response struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		PromptEvalCount int    `json:"prompt_eval_count"`
		EvalCount       int    `json:"eval_count"`
		Error           string `json:"error"`
	}
	if err := c.postJSON(ctx, "/api/chat", payload, &response); err != nil {
		return ChatResponse{}, err
	}
	if response.Error != "" {
		return ChatResponse{}, errors.New(response.Error)
	}
	if strings.TrimSpace(response.Message.Content) == "" {
		return ChatResponse{}, errors.New("ollama returned an empty chat response")
	}
	return ChatResponse{
		Content:          response.Message.Content,
		PromptTokens:     response.PromptEvalCount,
		CompletionTokens: response.EvalCount,
	}, nil
}

func (c *OllamaClient) Judge(ctx context.Context, request JudgeRequest) (string, error) {
	payload := map[string]any{
		"model":  request.Model,
		"prompt": request.Prompt,
		"stream": false,
		"format": "json",
	}

	var response struct {
		Response string `json:"response"`
		Error    string `json:"error"`
	}
	if err := c.postJSON(ctx, "/api/generate", payload, &response); err != nil {
		return "", err
	}
	if response.Error != "" {
		return "", errors.New(response.Error)
	}
	if strings.TrimSpace(response.Response) == "" {
		return "", errors.New("ollama returned an empty judge response")
	}
	return response.Response, nil
}

func (c *OllamaClient) postJSON(ctx context.Context, path string, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}
	if err := json.Unmarshal(content, target); err != nil {
		return fmt.Errorf("decode ollama response: %w", err)
	}
	return nil
}
