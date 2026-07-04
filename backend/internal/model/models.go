package model

import (
	"time"

	"gorm.io/datatypes"
)

const (
	SeverityNone     = "none"
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

type Conversation struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Message struct {
	ID             string    `gorm:"type:uuid;primaryKey" json:"id"`
	ConversationID string    `gorm:"type:uuid;index" json:"conversationId"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Model          string    `json:"model"`
	CreatedAt      time.Time `json:"createdAt"`
}

type LLMRequest struct {
	ID                 string    `gorm:"type:uuid;primaryKey" json:"id"`
	ConversationID     string    `gorm:"type:uuid;index" json:"conversationId"`
	UserMessageID      string    `gorm:"type:uuid;index" json:"userMessageId"`
	AssistantMessageID *string   `gorm:"type:uuid;index" json:"assistantMessageId,omitempty"`
	SystemPrompt       string    `json:"-"`
	UserInput          string    `json:"userInput"`
	ModelOutput        string    `json:"modelOutput"`
	ChatModel          string    `json:"chatModel"`
	LatencyMS          int       `json:"latencyMs"`
	PromptTokens       int       `json:"promptTokens"`
	CompletionTokens   int       `json:"completionTokens"`
	CreatedAt          time.Time `json:"createdAt"`
}

type RiskEvent struct {
	ID             string    `gorm:"type:uuid;primaryKey" json:"id"`
	ConversationID string    `gorm:"type:uuid;index" json:"conversationId"`
	MessageID      *string   `gorm:"type:uuid;index" json:"messageId,omitempty"`
	RequestID      string    `gorm:"type:uuid;index" json:"requestId"`
	Type           string    `json:"type"`
	Severity       string    `json:"severity"`
	Score          int       `json:"score"`
	Evidence       string    `json:"evidence"`
	Reason         string    `json:"reason"`
	Source         string    `json:"source"`
	CreatedAt      time.Time `json:"createdAt"`
}

type DetectorRun struct {
	ID            string         `gorm:"type:uuid;primaryKey" json:"id"`
	RequestID     string         `gorm:"type:uuid;index" json:"requestId"`
	DetectorName  string         `json:"detectorName"`
	DetectorModel string         `json:"detectorModel"`
	Phase         string         `json:"phase"`
	InputJSON     datatypes.JSON `gorm:"type:jsonb" json:"inputJson"`
	OutputJSON    datatypes.JSON `gorm:"type:jsonb" json:"outputJson"`
	RawOutput     string         `json:"rawOutput"`
	LatencyMS     int            `json:"latencyMs"`
	Success       bool           `json:"success"`
	ErrorMessage  string         `json:"errorMessage"`
	CreatedAt     time.Time      `json:"createdAt"`
}

type SeverityCount struct {
	Severity string `json:"severity"`
	Count    int64  `json:"count"`
}

type LabelCount struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type HighRiskConversation struct {
	ConversationID string    `json:"conversationId"`
	Title          string    `json:"title"`
	MaxScore       int       `json:"maxScore"`
	EventCount     int64     `json:"eventCount"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
