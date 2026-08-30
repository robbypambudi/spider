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

## Known false positive: Windows Defender flags release binaries

Prebuilt Windows binaries from [Releases](https://github.com/robbypambudi/spider/releases)
(`spider.exe`, `spider-worker.exe`) may be flagged by Microsoft Defender as
`Trojan:Win32/Wacatac.B!ml`. The `!ml` suffix means this is a **cloud ML
heuristic classification**, not a signature match — SPIDER does not touch the
registry, Group Policy, or any of the other behavior described in that
detection family. It's a known false-positive pattern for unsigned Go binaries
that self-relaunch as a hidden background process (`spider worker join
--detach`, see [`backend/pkg/workerctl/daemon_windows.go`](backend/pkg/workerctl/daemon_windows.go))
— that specific combination (hidden window, self re-exec, file persistence,
outbound network calls) resembles heuristics used to catch backdoors, even
though every part of it here is legitimate.

Mitigations in place:

- Windows release builds keep debug symbols (no `-s -w`) — stripped binaries
  are more likely to trip AV heuristics; see [`.github/workflows/release.yml`](.github/workflows/release.yml).
- Every release includes `SHA256SUMS.txt` — verify the binary you downloaded
  matches before trusting it.

Not yet done: code signing (requires a paid certificate) and submission to
[Microsoft's file submission portal](https://www.microsoft.com/en-us/wdsi/filesubmission)
for false-positive review. If you hit this detection, verify the checksum,
and/or build from source instead (`cd backend && go build ./cmd/spider`).
