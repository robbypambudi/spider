# Security pipeline

`pkg/security/pipeline/SecurityPipeline.inspect()` is the inspection stage:

1. **Preprocessor** — unicode/whitespace normalization
2. **Chunker** — default `FixedSizeChunker`
3. **Detector** — `PromptInjectionDetector.detect(text) → DetectionResult`
4. **Aggregator** — default `MaxScoreAggregator` (long documents are as risky as their worst chunk)
5. **Policy** — default `ThresholdPolicy` (`score >= threshold → BLOCK`)
6. **Enforcer** — converts the decision into `forward | reject | hold`

The pipeline does not call LLM providers. `InferenceService` asks the enforcer whether to forward **after** inspection. A `BLOCK` result never reaches `LLMProvider.infer()`.
