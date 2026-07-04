package repository

import (
	"context"
	"errors"
	"time"

	"log-llm/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateConversation(ctx context.Context, title string) (*model.Conversation, error) {
	now := time.Now().UTC()
	conversation := &model.Conversation{
		ID:        uuid.NewString(),
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := r.db.WithContext(ctx).Create(conversation).Error; err != nil {
		return nil, err
	}
	return conversation, nil
}

func (r *Repository) GetConversation(ctx context.Context, id string) (*model.Conversation, error) {
	var conversation model.Conversation
	if err := r.db.WithContext(ctx).First(&conversation, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *Repository) TouchConversation(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&model.Conversation{}).Where("id = ?", id).Update("updated_at", time.Now().UTC()).Error
}

func (r *Repository) ListConversations(ctx context.Context, limit int) ([]model.Conversation, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var conversations []model.Conversation
	err := r.db.WithContext(ctx).Order("updated_at DESC").Limit(limit).Find(&conversations).Error
	return conversations, err
}

func (r *Repository) CreateMessage(ctx context.Context, message *model.Message) error {
	if message.ID == "" {
		message.ID = uuid.NewString()
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *Repository) ListMessages(ctx context.Context, conversationID string) ([]model.Message, error) {
	var messages []model.Message
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (r *Repository) CreateLLMRequest(ctx context.Context, request *model.LLMRequest) error {
	if request.ID == "" {
		request.ID = uuid.NewString()
	}
	if request.CreatedAt.IsZero() {
		request.CreatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Create(request).Error
}

func (r *Repository) UpdateLLMRequestResponse(ctx context.Context, requestID string, assistantMessageID string, output string, latencyMS int, promptTokens int, completionTokens int) error {
	return r.db.WithContext(ctx).
		Model(&model.LLMRequest{}).
		Where("id = ?", requestID).
		Updates(map[string]any{
			"assistant_message_id": assistantMessageID,
			"model_output":         output,
			"latency_ms":           latencyMS,
			"prompt_tokens":        promptTokens,
			"completion_tokens":    completionTokens,
		}).Error
}

func (r *Repository) CreateRiskEvents(ctx context.Context, events []model.RiskEvent) error {
	if len(events) == 0 {
		return nil
	}
	now := time.Now().UTC()
	for i := range events {
		if events[i].ID == "" {
			events[i].ID = uuid.NewString()
		}
		if events[i].CreatedAt.IsZero() {
			events[i].CreatedAt = now
		}
	}
	return r.db.WithContext(ctx).Create(&events).Error
}

func (r *Repository) CreateDetectorRun(ctx context.Context, run *model.DetectorRun) error {
	if run.ID == "" {
		run.ID = uuid.NewString()
	}
	if run.CreatedAt.IsZero() {
		run.CreatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *Repository) ListRiskEvents(ctx context.Context, conversationID string, limit int) ([]model.RiskEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit)
	if conversationID != "" {
		query = query.Where("conversation_id = ?", conversationID)
	}
	var events []model.RiskEvent
	err := query.Find(&events).Error
	return events, err
}

func (r *Repository) ListDetectorRunsByConversation(ctx context.Context, conversationID string) ([]model.DetectorRun, error) {
	var runs []model.DetectorRun
	err := r.db.WithContext(ctx).
		Table("detector_runs").
		Select("detector_runs.*").
		Joins("JOIN llm_requests ON llm_requests.id = detector_runs.request_id").
		Where("llm_requests.conversation_id = ?", conversationID).
		Order("detector_runs.created_at DESC").
		Scan(&runs).Error
	return runs, err
}

func (r *Repository) ListDetectorRunsByRequest(ctx context.Context, requestID string) ([]model.DetectorRun, error) {
	var runs []model.DetectorRun
	err := r.db.WithContext(ctx).
		Where("request_id = ?", requestID).
		Order("created_at DESC").
		Find(&runs).Error
	return runs, err
}

func (r *Repository) SeverityCounts(ctx context.Context) ([]model.SeverityCount, error) {
	var counts []model.SeverityCount
	err := r.db.WithContext(ctx).
		Model(&model.RiskEvent{}).
		Select("severity, count(*) AS count").
		Where("type <> ?", "benign").
		Group("severity").
		Scan(&counts).Error
	return counts, err
}

func (r *Repository) TypeCounts(ctx context.Context) ([]model.LabelCount, error) {
	var counts []model.LabelCount
	err := r.db.WithContext(ctx).
		Model(&model.RiskEvent{}).
		Select("type AS label, count(*) AS count").
		Where("type <> ?", "benign").
		Group("type").
		Order("count DESC").
		Scan(&counts).Error
	return counts, err
}

func (r *Repository) SourceCounts(ctx context.Context) ([]model.LabelCount, error) {
	var counts []model.LabelCount
	err := r.db.WithContext(ctx).
		Model(&model.RiskEvent{}).
		Select("source AS label, count(*) AS count").
		Group("source").
		Order("count DESC").
		Scan(&counts).Error
	return counts, err
}

func (r *Repository) HighRiskConversations(ctx context.Context, limit int) ([]model.HighRiskConversation, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	var conversations []model.HighRiskConversation
	err := r.db.WithContext(ctx).
		Table("risk_events").
		Select("risk_events.conversation_id, conversations.title, MAX(risk_events.score) AS max_score, COUNT(*) AS event_count, conversations.updated_at").
		Joins("JOIN conversations ON conversations.id = risk_events.conversation_id").
		Where("risk_events.type <> ?", "benign").
		Group("risk_events.conversation_id, conversations.title, conversations.updated_at").
		Order("max_score DESC, event_count DESC").
		Limit(limit).
		Scan(&conversations).Error
	return conversations, err
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
