# Cluster

Workers represent serving-capable nodes. The cluster is not a general job platform.

Flow after ALLOW:

```text
ServingRouter → Scheduler.select_worker → serving node
```

Kubernetes is optional. Local compose runs a CPU-only mock worker.
