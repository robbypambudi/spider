# Aggregation

Chunk scores are not independent final answers. `DetectionAggregator` produces one document-level score.

Implemented: `MaxScoreAggregator`.

Prepared: `MeanScoreAggregator`, `TopKAggregator`, `VotingAggregator`, `WeightedAggregator`.

Max-score is the conservative research default: one high-scoring chunk can block the request.
