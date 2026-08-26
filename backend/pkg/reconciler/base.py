from __future__ import annotations

from typing import Protocol


class Reconciler(Protocol):
    name: str

    async def reconcile(self) -> None: ...
