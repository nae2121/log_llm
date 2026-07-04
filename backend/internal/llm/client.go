package llm

import (
	"context"
	"errors"

	"log-llm/backend/internal/config"
)

type ChatRequest struct {
	Provider     string
	Model        string
	SystemPrompt string
	UserInput    string
}

type ChatResponse struct {
	Content          string
	PromptTokens     int
	CompletionTokens int
}

type JudgeRequest struct {
	Provider string
	Model    string
	Prompt   string
}

type ChatLLMClient interface {
	Chat(ctx context.Context, request ChatRequest) (ChatResponse, error)
}

type JudgeLLMClient interface {
	Judge(ctx context.Context, request JudgeRequest) (string, error)
}

type Router struct {
	cfg    config.Config
	ollama *OllamaClient
}

func NewRouter(cfg config.Config) *Router {
	return &Router{
		cfg:    cfg,
		ollama: NewOllamaClient(cfg.OllamaBaseURL),
	}
}

func (r *Router) Chat(ctx context.Context, request ChatRequest) (ChatResponse, error) {
	provider := providerOrDefault(request.Provider, r.cfg.DefaultChatProvider)
	switch provider {
	case "ollama":
		if request.Model == "" {
			request.Model = r.cfg.OllamaChatModel
		}
		return r.ollama.Chat(ctx, request)
	case "gemini":
		return ChatResponse{}, errors.New("gemini chat provider is not implemented yet")
	case "openai":
		return ChatResponse{}, errors.New("openai chat provider is not implemented yet")
	default:
		return ChatResponse{}, errors.New("unsupported chat provider")
	}
}

func (r *Router) Judge(ctx context.Context, request JudgeRequest) (string, error) {
	provider := providerOrDefault(request.Provider, r.cfg.DefaultJudgeProvider)
	switch provider {
	case "ollama":
		if request.Model == "" {
			request.Model = r.cfg.OllamaJudgeModel
		}
		return r.ollama.Judge(ctx, request)
	case "gemini":
		return "", errors.New("gemini judge provider is not implemented yet")
	case "openai":
		return "", errors.New("openai judge provider is not implemented yet")
	default:
		return "", errors.New("unsupported judge provider")
	}
}

func providerOrDefault(provider string, fallback string) string {
	if provider == "" {
		return fallback
	}
	return provider
}
