from __future__ import annotations

from pkg.apis.security.models import SecurityDecision
from pkg.security.enforcement import EnforcementAction, Enforcer, FailMode


def test_enforcer_block_rejects() -> None:
    enforcer = Enforcer(fail_mode=FailMode.CLOSED)
    assert enforcer.resolve(SecurityDecision.ALLOW) is EnforcementAction.FORWARD
    assert enforcer.resolve(SecurityDecision.BLOCK) is EnforcementAction.REJECT
    assert enforcer.resolve(SecurityDecision.REVIEW) is EnforcementAction.HOLD
    assert enforcer.should_forward(SecurityDecision.BLOCK) is False


def test_fail_closed_on_error() -> None:
    enforcer = Enforcer(fail_mode=FailMode.CLOSED)
    assert enforcer.resolve(SecurityDecision.ERROR) is EnforcementAction.REJECT


def test_fail_open_on_error_is_configurable() -> None:
    enforcer = Enforcer(fail_mode=FailMode.OPEN)
    assert enforcer.resolve(SecurityDecision.ERROR) is EnforcementAction.FORWARD
