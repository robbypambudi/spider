from __future__ import annotations

from dao.repositories.policy import PolicyRepository
from pkg.config.settings import Settings


class PolicyService:
    def __init__(self, policies: PolicyRepository, settings: Settings) -> None:
        self.policies = policies
        self.settings = settings

    async def list_policies(self) -> list[dict[str, object]]:
        rows = await self.policies.list_policies()
        prepared: list[dict[str, object]] = [
            {"name": "adaptive", "status": "prepared"},
            {"name": "detector-specific", "status": "prepared"},
            {"name": "tenant", "status": "prepared"},
            {"name": "risk-based", "status": "prepared"},
        ]
        implemented = [
            {
                "name": row.name,
                "kind": row.kind,
                "threshold": row.threshold,
                "action_on_detection": row.action_on_detection,
                "is_default": row.is_default,
                "status": "implemented",
            }
            for row in rows
        ]
        if not implemented:
            implemented.append(
                {
                    "name": self.settings.default_security_policy,
                    "kind": "threshold",
                    "threshold": self.settings.default_threshold,
                    "action_on_detection": "block",
                    "is_default": True,
                    "status": "implemented",
                }
            )
        return implemented + prepared
