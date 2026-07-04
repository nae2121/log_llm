package service

import (
	"context"

	"log-llm/backend/internal/model"
	"log-llm/backend/internal/repository"
)

type DashboardService struct {
	repo *repository.Repository
}

type RiskSummary struct {
	SeverityCounts        []model.SeverityCount        `json:"severityCounts"`
	CategoryCounts        []model.LabelCount           `json:"categoryCounts"`
	SourceCounts          []model.LabelCount           `json:"sourceCounts"`
	HighRiskConversations []model.HighRiskConversation `json:"highRiskConversations"`
}

type ConversationDetail struct {
	Conversation model.Conversation  `json:"conversation"`
	Messages     []model.Message     `json:"messages"`
	RiskEvents   []model.RiskEvent   `json:"riskEvents"`
	DetectorRuns []model.DetectorRun `json:"detectorRuns"`
}

func NewDashboardService(repo *repository.Repository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) ListConversations(ctx context.Context) ([]model.Conversation, error) {
	return s.repo.ListConversations(ctx, 100)
}

func (s *DashboardService) ListMessages(ctx context.Context, conversationID string) ([]model.Message, error) {
	return s.repo.ListMessages(ctx, conversationID)
}

func (s *DashboardService) RiskEvents(ctx context.Context, conversationID string, limit int) ([]model.RiskEvent, error) {
	return s.repo.ListRiskEvents(ctx, conversationID, limit)
}

func (s *DashboardService) RiskSummary(ctx context.Context) (*RiskSummary, error) {
	severityCounts, err := s.repo.SeverityCounts(ctx)
	if err != nil {
		return nil, err
	}
	categoryCounts, err := s.repo.TypeCounts(ctx)
	if err != nil {
		return nil, err
	}
	sourceCounts, err := s.repo.SourceCounts(ctx)
	if err != nil {
		return nil, err
	}
	highRiskConversations, err := s.repo.HighRiskConversations(ctx, 10)
	if err != nil {
		return nil, err
	}
	return &RiskSummary{
		SeverityCounts:        normalizeSeverityCounts(severityCounts),
		CategoryCounts:        categoryCounts,
		SourceCounts:          sourceCounts,
		HighRiskConversations: highRiskConversations,
	}, nil
}

func (s *DashboardService) ConversationDetail(ctx context.Context, conversationID string) (*ConversationDetail, error) {
	conversation, err := s.repo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	messages, err := s.repo.ListMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	events, err := s.repo.ListRiskEvents(ctx, conversationID, 500)
	if err != nil {
		return nil, err
	}
	runs, err := s.repo.ListDetectorRunsByConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	return &ConversationDetail{
		Conversation: *conversation,
		Messages:     messages,
		RiskEvents:   events,
		DetectorRuns: runs,
	}, nil
}

func (s *DashboardService) DetectorRunsByRequest(ctx context.Context, requestID string) ([]model.DetectorRun, error) {
	return s.repo.ListDetectorRunsByRequest(ctx, requestID)
}

func normalizeSeverityCounts(counts []model.SeverityCount) []model.SeverityCount {
	known := map[string]int64{
		model.SeverityHigh:     0,
		model.SeverityMedium:   0,
		model.SeverityLow:      0,
		model.SeverityCritical: 0,
	}
	for _, count := range counts {
		known[count.Severity] = count.Count
	}
	return []model.SeverityCount{
		{Severity: model.SeverityCritical, Count: known[model.SeverityCritical]},
		{Severity: model.SeverityHigh, Count: known[model.SeverityHigh]},
		{Severity: model.SeverityMedium, Count: known[model.SeverityMedium]},
		{Severity: model.SeverityLow, Count: known[model.SeverityLow]},
	}
}
