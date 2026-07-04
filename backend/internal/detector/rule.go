package detector

import (
	"context"
	"regexp"
	"strings"
)

type RuleDetector struct{}

type keywordRule struct {
	Type     string
	Severity string
	Score    int
	Reason   string
	Keywords []string
}

func NewRuleDetector() *RuleDetector {
	return &RuleDetector{}
}

func (d *RuleDetector) Detect(_ context.Context, input DetectionInput) ([]RiskEvent, error) {
	source := sourceFor("rule", input.Phase)
	text := input.UserInput
	if input.Phase == PhasePostcheck {
		text = input.UserInput + "\n" + input.ModelOutput
	}

	events := make([]RiskEvent, 0)
	for _, rule := range keywordRules {
		for _, keyword := range rule.Keywords {
			if evidence, ok := findEvidence(text, keyword); ok {
				events = append(events, RiskEvent{
					Type:     rule.Type,
					Severity: rule.Severity,
					Score:    rule.Score,
					Evidence: evidence,
					Reason:   rule.Reason,
					Source:   source,
				})
			}
		}
	}

	if input.Phase == PhasePostcheck {
		events = append(events, outputLeakEvents(input.ModelOutput, source)...)
	}

	return events, nil
}

var keywordRules = []keywordRule{
	{
		Type:     "prompt_injection",
		Severity: "high",
		Score:    85,
		Reason:   "既存の指示やシステムの優先順位を上書きしようとしているため",
		Keywords: []string{
			"前の命令を無視",
			"以前の指示を無視",
			"システムプロンプトを上書き",
			"ignore previous instructions",
			"ignore all instructions",
			"system prompt",
			"developer message",
		},
	},
	{
		Type:     "system_prompt_leak_attempt",
		Severity: "high",
		Score:    90,
		Reason:   "内部指示や隠しルールの開示を求めているため",
		Keywords: []string{
			"システムプロンプトを表示",
			"内部指示を教えて",
			"隠しルールを教えて",
			"reveal your prompt",
			"show me your system prompt",
		},
	},
	{
		Type:     "jailbreak",
		Severity: "high",
		Score:    82,
		Reason:   "安全制約や検閲を無効化する脱獄指示が含まれているため",
		Keywords: []string{
			"DAN",
			"制限を解除",
			"安全ルールを無視",
			"検閲なし",
			"unrestricted mode",
		},
	},
	{
		Type:     "sensitive_info_request",
		Severity: "medium",
		Score:    65,
		Reason:   "認証情報や秘密情報に関連する語句が含まれているため",
		Keywords: []string{
			"APIキー",
			"token",
			"password",
			"secret",
			".env",
			"環境変数",
			"秘密鍵",
		},
	},
	{
		Type:     "tool_abuse",
		Severity: "medium",
		Score:    60,
		Reason:   "外部ツールやコマンドの不正利用につながる可能性があるため",
		Keywords: []string{
			"execute command",
			"run shell",
			"curl http",
			"rm -rf",
			"外部ツールを実行",
			"コマンドを実行",
		},
	},
	{
		Type:     "rag_injection",
		Severity: "medium",
		Score:    58,
		Reason:   "検索・RAG文書内の指示として振る舞わせる注入の可能性があるため",
		Keywords: []string{
			"retrieved document",
			"ignore this document",
			"この文書の指示に従え",
			"検索結果の命令",
		},
	},
	{
		Type:     "policy_bypass",
		Severity: "high",
		Score:    76,
		Reason:   "ポリシーや制約の回避を明示的に求めているため",
		Keywords: []string{
			"bypass policy",
			"policy bypass",
			"ポリシーを回避",
			"安全ポリシーを無視",
		},
	},
}

var leakRegexes = []struct {
	eventType string
	severity  string
	score     int
	reason    string
	pattern   *regexp.Regexp
}{
	{
		eventType: "system_prompt_leak_success",
		severity:  "critical",
		score:     95,
		reason:    "モデル出力にシステムプロンプトや開発者メッセージらしき内容が含まれているため",
		pattern:   regexp.MustCompile(`(?i)(system prompt|developer message|hidden instruction|内部指示|隠しルール)`),
	},
	{
		eventType: "sensitive_info_leak",
		severity:  "critical",
		score:     96,
		reason:    "モデル出力に認証情報や秘密鍵らしき文字列が含まれているため",
		pattern:   regexp.MustCompile(`(?i)(sk-[A-Za-z0-9_-]{16,}|AIza[0-9A-Za-z_-]{20,}|BEGIN [A-Z ]*PRIVATE KEY|DATABASE_URL=|password\s*=|token\s*=|secret\s*=|\.env)`),
	},
}

func outputLeakEvents(output string, source string) []RiskEvent {
	events := make([]RiskEvent, 0)
	for _, rule := range leakRegexes {
		match := rule.pattern.FindString(output)
		if match == "" {
			continue
		}
		events = append(events, RiskEvent{
			Type:     rule.eventType,
			Severity: rule.severity,
			Score:    rule.score,
			Evidence: truncate(match, 240),
			Reason:   rule.reason,
			Source:   source,
		})
	}
	return events
}

func findEvidence(text string, keyword string) (string, bool) {
	lowerText := strings.ToLower(text)
	lowerKeyword := strings.ToLower(keyword)
	index := strings.Index(lowerText, lowerKeyword)
	if index == -1 {
		return "", false
	}
	end := index + len(keyword)
	if end > len(text) {
		end = len(text)
	}
	return truncate(text[index:end], 240), true
}

func truncate(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit]) + "..."
}
