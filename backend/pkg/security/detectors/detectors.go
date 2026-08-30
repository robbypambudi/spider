package detectors

import (
	"context"
	"regexp"
	"time"

	"github.com/spider/spider/pkg/apis"
)

type Detector interface {
	Name() string
	Version() string
	Detect(ctx context.Context, text string) (apis.DetectionResult, error)
}

var Registry = map[string]func() Detector{
	"rule-based": func() Detector { return &RuleBasedDetector{} },
	"noop":       func() Detector { return &NoOpDetector{} },
}

func RegisterPromptShield(endpoint, model string) {
	Registry["prompt-shield"] = func() Detector {
		return NewPromptShieldDetector(endpoint, model)
	}
}

var injectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ignore\s+(all\s+)?(previous|prior|above)\s+instructions`),
	regexp.MustCompile(`(?i)disregard\s+(your|all|the)\s+(previous\s+)?instructions`),
	regexp.MustCompile(`(?i)reveal\s+(the\s+)?system\s+prompt`),
	regexp.MustCompile(`(?i)show\s+(me\s+)?(your\s+)?(hidden\s+)?system\s+prompt`),
	regexp.MustCompile(`(?i)you\s+are\s+now\s+dan`),
	regexp.MustCompile(`(?i)jailbreak`),
	regexp.MustCompile(`(?i)pretend\s+you\s+have\s+no\s+(restrictions|rules|guidelines)`),
	regexp.MustCompile(`(?i)do\s+not\s+follow\s+(your\s+)?(safety\s+)?(policy|policies|rules)`),
	regexp.MustCompile(`(?i)override\s+(the\s+)?(safety|content)\s+(policy|filters?)`),
	regexp.MustCompile(`(?i)new\s+instructions?\s*:\s*you\s+are`),
	regexp.MustCompile(`(?i)from\s+now\s+on\s+you\s+(will|must)\s+ignore`),
}

type RuleBasedDetector struct{}

func (d *RuleBasedDetector) Name() string    { return "rule-based" }
func (d *RuleBasedDetector) Version() string { return "0.1.0-dev" }

func (d *RuleBasedDetector) Detect(ctx context.Context, text string) (apis.DetectionResult, error) {
	_ = ctx
	start := time.Now()
	var matched []string
	for _, pattern := range injectionPatterns {
		if pattern.MatchString(text) {
			matched = append(matched, pattern.String())
		}
	}
	isInjection := len(matched) > 0
	score := 0.0
	if isInjection {
		score = 1.0
	}
	latency := float64(time.Since(start).Microseconds()) / 1000.0
	return apis.DetectionResult{
		Detector:    d.Name(),
		Score:       score,
		IsInjection: isInjection,
		LatencyMs:   latency,
		Metadata: map[string]interface{}{
			"version":          d.Version(),
			"kind":             "rule-based",
			"warning":          "development/testing only",
			"matched_patterns": matched,
		},
	}, nil
}

type NoOpDetector struct{}

func (d *NoOpDetector) Name() string    { return "noop" }
func (d *NoOpDetector) Version() string { return "0.1.0" }

func (d *NoOpDetector) Detect(ctx context.Context, text string) (apis.DetectionResult, error) {
	_ = ctx
	_ = text
	start := time.Now()
	latency := float64(time.Since(start).Microseconds()) / 1000.0
	return apis.DetectionResult{
		Detector:    d.Name(),
		Score:       0,
		IsInjection: false,
		LatencyMs:   latency,
		Metadata: map[string]interface{}{
			"version": d.Version(),
			"kind":    "noop",
		},
	}, nil
}

func Names() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[j] < names[i] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	return names
}
