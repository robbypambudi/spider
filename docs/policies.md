# Policies

`ThresholdPolicy` compares the aggregated score to a configured threshold:

```text
score >= threshold → BLOCK (or REVIEW if action_on_detection=review)
otherwise → ALLOW
```

`SPIDER_DEFAULT_THRESHOLD` is the source of truth for the MVP. Thresholds must stay outside detectors so evaluation can sweep operating points (TPR @ FPR 0.05%, 0.1%, 0.5%, 1%).

Prepared: `AdaptiveThresholdPolicy`, `DetectorSpecificPolicy`, `TenantPolicy`, `RiskBasedPolicy`.
