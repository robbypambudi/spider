from __future__ import annotations


class ServingNodeReconciler:
    name = "serving-node"

    async def reconcile(self) -> None:
        """Sync serving-node desired vs actual state. Inventory is worker-driven in MVP."""
        return None


class ModelReconciler:
    name = "model"

    async def reconcile(self) -> None:
        raise NotImplementedError("ModelReconciler requires desired-model CRDs. See docs/cluster.md.")


class DeploymentReconciler:
    name = "deployment"

    async def reconcile(self) -> None:
        raise NotImplementedError(
            "DeploymentReconciler requires serving deployment objects. See docs/cluster.md."
        )
