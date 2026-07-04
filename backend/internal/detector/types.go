package detector

import (
	"context"
	"strings"
)

const (
	PhasePrecheck  = "precheck"
	PhasePostcheck = "postcheck"

	SourceRulePrecheck      = "rule_precheck"
	SourceRulePostcheck     = "rule_postcheck"
	SourceLLMJudgePrecheck  = "llm_judge_precheck"
	SourceLLMJudgePostcheck = "llm_judge_postcheck"
)

var allowedTypes = map[string]bool{
	"prompt_injection":           true,
	"jailbreak":                  true,
	"system_prompt_leak_attempt": true,
	"system_prompt_leak_success": true,
	"sensitive_info_request":     true,
	"sensitive_info_leak":        true,
	"tool_abuse":                 true,
	"rag_injection":              true,
	"policy_bypass":              true,
	"benign":                     true,
}

var allowedSeverities = map[string]bool{
	"none":     true,
	"low":      true,
	"medium":   true,
	"high":     true,
	"critical": true,
}

type DetectionInput struct {
	ConversationID string `json:"conversationId"`
	RequestID      string `json:"requestId"`
	UserInput      string `json:"userInput"`
	ModelOutput    string `json:"modelOutput"`
	SystemPrompt   string `json:"-"`
	Model          string `json:"model"`
	Phase          string `json:"phase"`
}

type RiskEvent struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Score    int    `json:"score"`
	Evidence string `json:"evidence"`
	Reason   string `json:"reason"`
	Source   string `json:"source"`
}

type Detector interface {
	Detect(ctx context.Context, input DetectionInput) ([]RiskEvent, error)
}

func sourceFor(prefix string, phase string) string {
	if strings.EqualFold(phase, PhasePostcheck) {
		return prefix + "_postcheck"
	}
	return prefix + "_precheck"
}

func severityFromScore(score int) string {
	switch {
	case score >= 90:
		return "critical"
	case score >= 70:
		return "high"
	case score >= 40:
		return "medium"
	case score > 0:
		return "low"
	default:
		return "none"
	}
}

func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}
