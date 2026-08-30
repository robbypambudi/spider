# Policies

`ThresholdPolicy` compares the aggregated score to a configured threshold:

```text
score >= threshold → BLOCK (or REVIEW if action_on_detection=review)
otherwise → ALLOW
```

Runtime source of truth: the **default policy** row in the `policies` table (`is_default=true`). It stores:

- `threshold`, `action_on_detection` (columns)
- `chunker`, `chunk_size`, `chunk_overlap` (`config_json`)

`SPIDER_DEFAULT_THRESHOLD` and chunk env vars bootstrap the first default policy on startup. After that, use the Settings UI or `/api/v1/security/policies` CRUD (ADMIN) to change the active operating point without redeploying.

Prepared: `AdaptiveThresholdPolicy`, `DetectorSpecificPolicy`, `TenantPolicy`, `RiskBasedPolicy`.
