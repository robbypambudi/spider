import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";

export function WorkerDetailPage() {
  const { workerId = "" } = useParams();
  const { data } = useQuery({
    queryKey: ["workers", workerId],
    queryFn: () => api.worker(workerId),
    enabled: Boolean(workerId),
    staleTime: 10_000,
  });

  if (!data) return <div className="text-zinc-400">Loading worker…</div>;

  return (
    <div className="space-y-4">
      <h1 className="font-mono text-2xl">{data.worker_id}</h1>
      <Card className="space-y-2 font-mono text-sm">
        <div>
          status <Badge value={data.status} />
        </div>
        <div>hostname={data.hostname}</div>
        <div>site={data.site ?? "—"}</div>
        <div>version={data.version}</div>
        <div>
          cpu={data.resources.cpu_total} ram={data.resources.memory_total_mb}MB
        </div>
        <div>running_requests={data.running_requests}</div>
        <div>models={data.models.map((m) => `${m.name}:${m.status}`).join(", ") || "—"}</div>
      </Card>
      <Card>
        <h2 className="mb-3 text-sm uppercase text-zinc-500">GPUs</h2>
        {data.resources.gpus.length === 0 ? (
          <p className="text-sm text-zinc-400">CPU-only worker. GPU hardware is optional.</p>
        ) : (
          <ul className="space-y-2 font-mono text-sm">
            {data.resources.gpus.map((gpu) => (
              <li key={gpu.index}>
                [{gpu.index}] {gpu.name} {gpu.memory_used_mb}/{gpu.memory_total_mb}MB util={gpu.utilization}%
              </li>
            ))}
          </ul>
        )}
      </Card>
    </div>
  );
}
