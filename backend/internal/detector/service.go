package detector

import (
	"context"
	"sort"
	"strings"
)

type Service struct {
	detectors []Detector
}

type AggregateRisk struct {
	Score      int         `json:"score"`
	Severity   string      `json:"severity"`
	Confidence string      `json:"confidence"`
	Events     []RiskEvent `json:"events"`
}

func NewService(detectors ...Detector) *Service {
	return &Service{detectors: detectors}
}

func (s *Service) Detect(ctx context.Context, input DetectionInput) ([]RiskEvent, []error) {
	events := make([]RiskEvent, 0)
	errs := make([]error, 0)
	for _, detector := range s.detectors {
		detected, err := detector.Detect(ctx, input)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		events = append(events, detected...)
	}
	return events, errs
}

func Aggregate(events []RiskEvent) AggregateRisk {
	representatives := map[string]RiskEvent{}
	maxScore := 0
	sourceByType := map[string]map[string]bool{}

	for _, event := range events {
		if event.Type == "benign" {
			continue
		}
		event.Score = clampScore(event.Score)
		if event.Score > maxScore {
			maxScore = event.Score
		}
		current, exists := representatives[event.Type]
		if !exists || event.Score > current.Score {
			representatives[event.Type] = event
		}
		if sourceByType[event.Type] == nil {
			sourceByType[event.Type] = map[string]bool{}
		}
		sourceByType[event.Type][sourceKind(event.Source)] = true
	}

	summaryEvents := make([]RiskEvent, 0, len(representatives))
	for _, event := range representatives {
		summaryEvents = append(summaryEvents, event)
	}
	sort.Slice(summaryEvents, func(i, j int) bool {
		return summaryEvents[i].Score > summaryEvents[j].Score
	})

	confidence := "low"
	if maxScore > 0 {
		confidence = "medium"
	}
	for _, sources := range sourceByType {
		if sources["rule"] && sources["llm_judge"] {
			confidence = "high"
			break
		}
	}
	if maxScore > 0 && maxScore < 40 && confidence != "high" {
		confidence = "low"
	}

	return AggregateRisk{
		Score:      maxScore,
		Severity:   severityFromScore(maxScore),
		Confidence: confidence,
		Events:     summaryEvents,
	}
}

func sourceKind(source string) string {
	if strings.HasPrefix(source, "llm_judge") {
		return "llm_judge"
	}
	if strings.HasPrefix(source, "rule") {
		return "rule"
	}
	return source
}
