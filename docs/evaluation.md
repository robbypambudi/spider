# Evaluation

`pkg/security/evaluation` computes TPR, FPR, TNR, FNR, precision, recall, F1, AUC, and **TPR @ target FPR** from labeled scores.

Target operating points:

```text
TPR @ FPR 0.05% / 0.1% / 0.5% / 1%
```

`POST /api/v1/security/evaluate` accepts labeled samples. Thresholds are request/config parameters.

Security overhead:

```text
security_overhead_ms = protected_e2e - baseline_e2e
security_overhead_percent = overhead / baseline * 100
```

See `experiments/configs/baseline-low-fpr.yaml`.
