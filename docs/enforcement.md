# Enforcement

`Enforcer.resolve(decision)` maps:

| Decision | Action |
| --- | --- |
| ALLOW | forward to LLM serving |
| BLOCK | reject; do not call the provider |
| REVIEW | hold for secondary inspection |
| ERROR | `SPIDER_FAIL_MODE=closed` reject, or `open` forward |

Fail-open is never hardcoded. Production default is fail-closed.
