from __future__ import annotations

from enum import StrEnum

from pkg.apis.security.models import SecurityDecision


class FailMode(StrEnum):
    OPEN = "open"
    CLOSED = "closed"


class EnforcementAction(StrEnum):
    FORWARD = "forward"
    REJECT = "reject"
    HOLD = "hold"


class Enforcer:
    """Convert a security decision into a runtime action.

    Fail-open vs fail-closed is configuration, never hardcoded.
    """

    def __init__(self, fail_mode: FailMode | str = FailMode.CLOSED) -> None:
        self.fail_mode = FailMode(fail_mode)

    def resolve(self, decision: SecurityDecision) -> EnforcementAction:
        if decision is SecurityDecision.ALLOW:
            return EnforcementAction.FORWARD
        if decision is SecurityDecision.BLOCK:
            return EnforcementAction.REJECT
        if decision is SecurityDecision.REVIEW:
            return EnforcementAction.HOLD
        if decision is SecurityDecision.ERROR:
            if self.fail_mode is FailMode.OPEN:
                return EnforcementAction.FORWARD
            return EnforcementAction.REJECT
        return EnforcementAction.REJECT

    def should_forward(self, decision: SecurityDecision) -> bool:
        return self.resolve(decision) is EnforcementAction.FORWARD
