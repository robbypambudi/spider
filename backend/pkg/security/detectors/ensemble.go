package detectors

import (
	"context"
	"fmt"

	"github.com/spider/spider/pkg/apis"
)

// EnsembleDetector runs multiple detectors and returns the highest score.
// Used while Prompt-Shield ML weights are validated against the HF card.
type EnsembleDetector struct {
	Name_     string
	Version_  string
	Detectors []Detector
}

func NewEnsembleDetector(name string, detectors ...Detector) *EnsembleDetector {
	return &EnsembleDetector{Name_: name, Version_: "0.1.0", Detectors: detectors}
}

func (d *EnsembleDetector) Name() string    { return d.Name_ }
func (d *EnsembleDetector) Version() string { return d.Version_ }

func (d *EnsembleDetector) Detect(ctx context.Context, text string) (apis.DetectionResult, error) {
	if len(d.Detectors) == 0 {
		return apis.DetectionResult{}, fmt.Errorf("ensemble has no detectors")
	}

	best := apis.DetectionResult{Detector: d.Name_, Score: 0}
	var combined []apis.DetectionResult
	for _, det := range d.Detectors {
		result, err := det.Detect(ctx, text)
		if err != nil {
			return apis.DetectionResult{}, err
		}
		combined = append(combined, result)
		if result.Score > best.Score {
			best = result
		}
		if result.IsInjection {
			best.IsInjection = true
		}
	}
	if best.Score == 0 && len(combined) > 0 {
		best = combined[0]
	}
	best.Detector = d.Name_
	best.Metadata = map[string]interface{}{
		"version":  d.Version_,
		"kind":     "ensemble",
		"strategy": "max_score",
		"sources":  combined,
	}
	return best, nil
}

func RegisterPromptShieldEnsemble(endpoint, model string) {
	Registry["prompt-shield+rules"] = func() Detector {
		return NewEnsembleDetector("prompt-shield+rules",
			NewPromptShieldDetector(endpoint, model),
			&RuleBasedDetector{},
		)
	}
}
