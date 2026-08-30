package policies

import "github.com/spider/spider/pkg/apis"

type Policy interface {
	Name() string
	Threshold() float64
	Evaluate(result apis.AggregatedDetectionResult) apis.SecurityDecision
}

type ThresholdPolicy struct {
	ThresholdValue     float64
	ActionOnDetection  string
}

func NewThresholdPolicy(threshold float64, actionOnDetection string) *ThresholdPolicy {
	if actionOnDetection == "" {
		actionOnDetection = "block"
	}
	return &ThresholdPolicy{ThresholdValue: threshold, ActionOnDetection: actionOnDetection}
}

func (p *ThresholdPolicy) Name() string { return "threshold" }

func (p *ThresholdPolicy) Threshold() float64 { return p.ThresholdValue }

func (p *ThresholdPolicy) Evaluate(result apis.AggregatedDetectionResult) apis.SecurityDecision {
	if result.Score >= p.ThresholdValue {
		if p.ActionOnDetection == "review" {
			return apis.DecisionReview
		}
		return apis.DecisionBlock
	}
	return apis.DecisionAllow
}
