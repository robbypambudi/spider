# Detectors

Detectors return a **score**. They do not own ALLOW/BLOCK. Thresholds live in policy/config so TPR @ target-FPR sweeps can vary the threshold without changing detector code.

Implemented:

* `NoOpDetector` — always `0.0` (baseline)
* `RuleBasedDetector` — **development/testing only**; regex/keyword matches score `1.0`

Prepared (not implemented; do not fake model scores):

* `PromptShieldDetector`
* `FlanT5Detector`
* `TransformerDetector`
* `RemoteDetector`
* `EnsembleDetector`
