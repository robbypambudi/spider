package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/security/aggregation"
	"github.com/spider/spider/pkg/security/chunking"
	"github.com/spider/spider/pkg/security/detectors"
	"github.com/spider/spider/pkg/security/enforcement"
	"github.com/spider/spider/pkg/security/policies"
	"github.com/spider/spider/pkg/security/preprocessing"
)

type Pipeline struct {
	Preprocessor preprocessing.Preprocessor
	Chunker      chunking.Chunker
	Detector     detectors.Detector
	Aggregator   aggregation.Aggregator
	Policy       policies.Policy
	Enforcer     *enforcement.Enforcer
}

type BuildOptions struct {
	DetectorName         string
	Threshold            float64
	Chunker              string
	ChunkSize            int
	ChunkOverlap         int
	FailMode             string
	PromptShieldEndpoint string
}

func NewChunker(chunkerName, endpoint string, chunkSize, chunkOverlap int) chunking.Chunker {
	if chunkSize <= 0 {
		chunkSize = 256
	}
	if chunkOverlap < 0 {
		chunkOverlap = 0
	}
	switch chunkerName {
	case "token":
		if endpoint != "" {
			return chunking.NewSidecarTokenChunker(endpoint, chunkSize, chunkOverlap)
		}
	}
	if chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 8
	}
	if chunkSize <= 0 {
		chunkSize = 2048
	}
	return chunking.NewFixedSizeChunker(chunkSize, chunkOverlap)
}

func BuildDefault(opts BuildOptions) (*Pipeline, error) {
	factory, ok := detectors.Registry[opts.DetectorName]
	if !ok {
		supported := detectors.Names()
		return nil, fmt.Errorf("unknown detector %q; implemented: %v", opts.DetectorName, supported)
	}
	chunkSize := opts.ChunkSize
	if chunkSize <= 0 {
		if opts.Chunker == "token" {
			chunkSize = 256
		} else {
			chunkSize = 2048
		}
	}
	chunkOverlap := opts.ChunkOverlap
	if chunkOverlap < 0 {
		chunkOverlap = 0
	}
	policy := policies.NewThresholdPolicy(opts.Threshold, "block")
	return &Pipeline{
		Preprocessor: &preprocessing.DefaultPreprocessor{},
		Chunker:      NewChunker(opts.Chunker, opts.PromptShieldEndpoint, chunkSize, chunkOverlap),
		Detector:     factory(),
		Aggregator:   &aggregation.MaxScoreAggregator{},
		Policy:       policy,
		Enforcer:     enforcement.NewEnforcer(opts.FailMode),
	}, nil
}

func (p *Pipeline) SetPolicy(threshold float64, actionOnDetection string) {
	p.Policy = policies.NewThresholdPolicy(threshold, actionOnDetection)
}

func (p *Pipeline) SetChunker(c chunking.Chunker) {
	if c != nil {
		p.Chunker = c
	}
}

func (p *Pipeline) Inspect(ctx context.Context, request apis.SecurityRequest) (result apis.SecurityResult) {
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			latency := float64(time.Since(start).Microseconds()) / 1000.0
			slog.Error("security_pipeline_panic",
				"request_id", request.RequestID,
				"panic", r,
			)
			result = apis.SecurityResult{
				RequestID:      request.RequestID,
				Decision:       apis.DecisionError,
				Score:          0,
				TotalLatencyMs: latency,
				Policy:         p.Policy.Name(),
				Metadata:       map[string]interface{}{"error": fmt.Sprintf("pipeline panic: %v", r)},
			}
		}
	}()

	processed := p.Preprocessor.Process(request.Text)
	chunks := p.Chunker.Chunk(processed.Text)

	var detectorResults []apis.DetectionResult
	for _, chunk := range chunks {
		result, err := p.Detector.Detect(ctx, chunk.Text)
		if err != nil {
			latency := float64(time.Since(start).Microseconds()) / 1000.0
			return apis.SecurityResult{
				RequestID:      request.RequestID,
				Decision:       apis.DecisionError,
				Score:          0,
				TotalLatencyMs: latency,
				Policy:         p.Policy.Name(),
				Metadata:       map[string]interface{}{"error": err.Error()},
			}
		}
		detectorResults = append(detectorResults, result)
	}

	aggregated := p.Aggregator.Aggregate(detectorResults)
	decision := p.Policy.Evaluate(aggregated)
	latency := float64(time.Since(start).Microseconds()) / 1000.0

	threshold := p.Policy.Threshold()
	metadata := map[string]interface{}{
		"detector":            p.Detector.Name(),
		"detector_version":    p.Detector.Version(),
		"chunker":             p.Chunker.Name(),
		"aggregator":          p.Aggregator.Name(),
		"threshold":           threshold,
		"source":              request.Source,
		"model":               request.Model,
		"original_length":     processed.OriginalLength,
		"preprocessed_length": len(processed.Text),
	}
	for k, v := range aggregated.Metadata {
		metadata[k] = v
	}

	return apis.SecurityResult{
		RequestID:       request.RequestID,
		Decision:        decision,
		Score:           aggregated.Score,
		DetectorResults: aggregated.DetectorResults,
		ChunksScanned:   aggregated.ChunksScanned,
		TotalLatencyMs:  latency,
		Policy:          p.Policy.Name(),
		Metadata:        metadata,
	}
}
