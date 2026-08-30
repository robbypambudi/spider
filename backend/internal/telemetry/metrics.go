package telemetry

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const Version = "0.1.0"
const ServiceName = "spider"

type Metrics struct {
	securityScansTotal           *prometheus.CounterVec
	securityAllowedTotal         prometheus.Counter
	securityBlockedTotal         prometheus.Counter
	securityReviewTotal          prometheus.Counter
	securityPipelineLatency      prometheus.Histogram
	securityDetectionLatency     *prometheus.HistogramVec
	securityChunksTotal          prometheus.Counter
	securityDetectorScore        *prometheus.HistogramVec
	inferenceRequestsTotal       *prometheus.CounterVec
	inferenceLatency             prometheus.Histogram
	inferenceSecurityOverhead    prometheus.Histogram
	registry                     *prometheus.Registry
}

func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{registry: reg}
	m.securityScansTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "spider_security_scans_total", Help: "Security scans by decision",
	}, []string{"decision"})
	m.securityAllowedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "spider_security_allowed_total", Help: "Allowed security decisions",
	})
	m.securityBlockedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "spider_security_blocked_total", Help: "Blocked security decisions",
	})
	m.securityReviewTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "spider_security_review_total", Help: "Review security decisions",
	})
	m.securityPipelineLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "spider_security_pipeline_latency_seconds", Help: "Pipeline latency",
		Buckets: prometheus.DefBuckets,
	})
	m.securityDetectionLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "spider_security_detection_latency_seconds", Help: "Detector latency",
		Buckets: prometheus.DefBuckets,
	}, []string{"detector"})
	m.securityChunksTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "spider_security_chunks_total", Help: "Chunks scanned",
	})
	m.securityDetectorScore = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "spider_security_detector_score", Help: "Detector scores",
		Buckets: []float64{0, 0.25, 0.5, 0.75, 1.0},
	}, []string{"detector"})
	m.inferenceRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "spider_inference_requests_total", Help: "Inference requests",
	}, []string{"status"})
	m.inferenceLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "spider_inference_latency_seconds", Help: "Inference latency",
		Buckets: prometheus.DefBuckets,
	})
	m.inferenceSecurityOverhead = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "spider_inference_security_overhead_seconds", Help: "Security overhead",
		Buckets: prometheus.DefBuckets,
	})
	reg.MustRegister(
		m.securityScansTotal, m.securityAllowedTotal, m.securityBlockedTotal,
		m.securityReviewTotal, m.securityPipelineLatency, m.securityDetectionLatency,
		m.securityChunksTotal, m.securityDetectorScore, m.inferenceRequestsTotal,
		m.inferenceLatency, m.inferenceSecurityOverhead,
	)
	return m
}

func (m *Metrics) ObserveScan(decision string, pipelineLatencyMs float64, chunks int, detectorName string, score, detectorLatencyMs float64) {
	m.securityScansTotal.WithLabelValues(decision).Inc()
	switch decision {
	case "ALLOW":
		m.securityAllowedTotal.Inc()
	case "BLOCK":
		m.securityBlockedTotal.Inc()
	case "REVIEW":
		m.securityReviewTotal.Inc()
	}
	m.securityPipelineLatency.Observe(pipelineLatencyMs / 1000.0)
	m.securityDetectionLatency.WithLabelValues(detectorName).Observe(detectorLatencyMs / 1000.0)
	m.securityChunksTotal.Add(float64(chunks))
	m.securityDetectorScore.WithLabelValues(detectorName).Observe(score)
}

func (m *Metrics) ObserveInference(status string, e2eMs, securityOverheadMs float64) {
	m.inferenceRequestsTotal.WithLabelValues(status).Inc()
	m.inferenceLatency.Observe(e2eMs / 1000.0)
	m.inferenceSecurityOverhead.Observe(securityOverheadMs / 1000.0)
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
