# Security Policy

## Reporting a Vulnerability

Do **not** open a public GitHub issue for security vulnerabilities.

Report suspected vulnerabilities privately to the maintainers. Include:

* a description of the issue
* affected components (`pkg/security`, API routes, worker protocol, etc.)
* reproduction steps
* impact assessment (confidentiality, integrity, availability)

## Scope

SPIDER is a runtime defense framework for prompt injection detection. Security
reports are especially valuable around:

* bypasses of the security pipeline (inspection skipped)
* BLOCK decisions that still reach an LLM provider
* prompt content leaking into logs or persistence when disabled
* authentication / authorization flaws
* worker registration without a valid worker token
* fail-open misconfiguration that cannot be disabled

## Operational Defaults

* Full prompt content is **not** logged or persisted by default.
* `SPIDER_FAIL_MODE=closed` refuses inference on pipeline errors.
* Detectors do **not** own ALLOW/BLOCK decisions; the policy engine does.
* Thresholds are configuration, not hardcoded detector constants.

## Development Detectors

`RuleBasedDetector` is for development and testing only. It is not a
production-quality prompt injection detector. Do not treat rule-based scores as
a security guarantee.
