package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"log-llm/backend/internal/config"
	"log-llm/backend/internal/detector"
	"log-llm/backend/internal/llm"
	"log-llm/backend/internal/model"
	"log-llm/backend/internal/repository"

	"github.com/google/uuid"
)

type ChatService struct {
	cfg       config.Config
	repo      *repository.Repository
	llm       llm.ChatLLMClient
	detection *detector.Service
}

type ChatRequest struct {
	ConversationID *string `json:"conversationId"`
	Message        string  `json:"message"`
	Model          string  `json:"model"`
}

type ChatResponse struct {
	ConversationID   string                 `json:"conversationId"`
	AssistantMessage model.Message          `json:"assistantMessage"`
	Risk             detector.AggregateRisk `json:"risk"`
}

func NewChatService(cfg config.Config, repo *repository.Repository, llmClient llm.ChatLLMClient, detection *detector.Service) *ChatService {
	return &ChatService{
		cfg:       cfg,
		repo:      repo,
		llm:       llmClient,
		detection: detection,
	}
}

func (s *ChatService) Chat(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	userInput := strings.TrimSpace(request.Message)
	if userInput == "" {
		return nil, errors.New("message is required")
	}

	provider := strings.TrimSpace(request.Model)
	if provider == "" {
		provider = s.cfg.DefaultChatProvider
	}
	modelLabel := s.modelLabel(provider)

	conversationID, err := s.ensureConversation(ctx, request.ConversationID, userInput)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	userMessage := &model.Message{
		ID:             uuid.NewString(),
		ConversationID: conversationID,
		Role:           "user",
		Content:        userInput,
		Model:          modelLabel,
		CreatedAt:      now,
	}
	if err := s.repo.CreateMessage(ctx, userMessage); err != nil {
		return nil, err
	}

	requestID := uuid.NewString()
	llmRequest := &model.LLMRequest{
		ID:             requestID,
		ConversationID: conversationID,
		UserMessageID:  userMessage.ID,
		SystemPrompt:   s.cfg.SystemPrompt,
		UserInput:      userInput,
		ChatModel:      modelLabel,
		CreatedAt:      now,
	}
	if err := s.repo.CreateLLMRequest(ctx, llmRequest); err != nil {
		return nil, err
	}

	preEvents, _ := s.detection.Detect(ctx, detector.DetectionInput{
		ConversationID: conversationID,
		RequestID:      requestID,
		UserInput:      userInput,
		SystemPrompt:   s.cfg.SystemPrompt,
		Model:          modelLabel,
		Phase:          detector.PhasePrecheck,
	})
	if err := s.repo.CreateRiskEvents(ctx, toModelRiskEvents(preEvents, conversationID, userMessage.ID, requestID)); err != nil {
		return nil, err
	}

	llmStart := time.Now()
	chatResponse, err := s.llm.Chat(ctx, llm.ChatRequest{
		Provider:     provider,
		Model:        s.chatModel(provider),
		SystemPrompt: s.cfg.SystemPrompt,
		UserInput:    userInput,
	})
	if err != nil {
		if !s.cfg.AllowLLMFallback {
			return nil, err
		}
		chatResponse = llm.ChatResponse{
			Content: "LLM backend is unavailable. The message was logged and analyzed, but no model response could be generated.",
		}
	}
	latencyMS := int(time.Since(llmStart).Milliseconds())

	assistantMessage := &model.Message{
		ID:             uuid.NewString(),
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        strings.TrimSpace(chatResponse.Content),
		Model:          modelLabel,
		CreatedAt:      time.Now().UTC(),
	}
	if assistantMessage.Content == "" {
		assistantMessage.Content = "The model returned an empty response."
	}
	if err := s.repo.CreateMessage(ctx, assistantMessage); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateLLMRequestResponse(ctx, requestID, assistantMessage.ID, assistantMessage.Content, latencyMS, chatResponse.PromptTokens, chatResponse.CompletionTokens); err != nil {
		return nil, err
	}

	postEvents, _ := s.detection.Detect(ctx, detector.DetectionInput{
		ConversationID: conversationID,
		RequestID:      requestID,
		UserInput:      userInput,
		ModelOutput:    assistantMessage.Content,
		SystemPrompt:   s.cfg.SystemPrompt,
		Model:          modelLabel,
		Phase:          detector.PhasePostcheck,
	})
	if err := s.repo.CreateRiskEvents(ctx, toModelRiskEventsWithOutput(postEvents, conversationID, userMessage.ID, assistantMessage.ID, requestID, assistantMessage.Content)); err != nil {
		return nil, err
	}

	_ = s.repo.TouchConversation(ctx, conversationID)

	allEvents := append(preEvents, postEvents...)
	return &ChatResponse{
		ConversationID:   conversationID,
		AssistantMessage: *assistantMessage,
		Risk:             detector.Aggregate(allEvents),
	}, nil
}

func (s *ChatService) ensureConversation(ctx context.Context, requestedID *string, firstMessage string) (string, error) {
	if requestedID != nil && strings.TrimSpace(*requestedID) != "" {
		conversation, err := s.repo.GetConversation(ctx, strings.TrimSpace(*requestedID))
		if err != nil {
			return "", err
		}
		return conversation.ID, nil
	}

	conversation, err := s.repo.CreateConversation(ctx, titleFromMessage(firstMessage))
	if err != nil {
		return "", err
	}
	return conversation.ID, nil
}

func (s *ChatService) modelLabel(provider string) string {
	if provider == "ollama" {
		return "ollama:" + s.cfg.OllamaChatModel
	}
	return provider
}

func (s *ChatService) chatModel(provider string) string {
	if provider == "ollama" {
		return s.cfg.OllamaChatModel
	}
	return ""
}

func toModelRiskEvents(events []detector.RiskEvent, conversationID string, messageID string, requestID string) []model.RiskEvent {
	converted := make([]model.RiskEvent, 0, len(events))
	for _, event := range events {
		id := messageID
		converted = append(converted, model.RiskEvent{
			ID:             uuid.NewString(),
			ConversationID: conversationID,
			MessageID:      &id,
			RequestID:      requestID,
			Type:           event.Type,
			Severity:       event.Severity,
			Score:          event.Score,
			Evidence:       event.Evidence,
			Reason:         event.Reason,
			Source:         event.Source,
			CreatedAt:      time.Now().UTC(),
		})
	}
	return converted
}

func toModelRiskEventsWithOutput(events []detector.RiskEvent, conversationID string, userMessageID string, assistantMessageID string, requestID string, modelOutput string) []model.RiskEvent {
	converted := make([]model.RiskEvent, 0, len(events))
	for _, event := range events {
		id := userMessageID
		if event.Evidence != "" && strings.Contains(modelOutput, event.Evidence) {
			id = assistantMessageID
		}
		converted = append(converted, model.RiskEvent{
			ID:             uuid.NewString(),
			ConversationID: conversationID,
			MessageID:      &id,
			RequestID:      requestID,
			Type:           event.Type,
			Severity:       event.Severity,
			Score:          event.Score,
			Evidence:       event.Evidence,
			Reason:         event.Reason,
			Source:         event.Source,
			CreatedAt:      time.Now().UTC(),
		})
	}
	return converted
}

func titleFromMessage(message string) string {
	trimmed := strings.Join(strings.Fields(message), " ")
	runes := []rune(trimmed)
	if len(runes) > 42 {
		return string(runes[:42]) + "..."
	}
	if trimmed == "" {
		return "New conversation"
	}
	return trimmed
}
