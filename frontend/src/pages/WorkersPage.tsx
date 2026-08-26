import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";

export function WorkersPage() {
  const { data } = useQuery({ queryKey: ["workers"], queryFn: api.workers, staleTime: 10_000 });

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-medium">Workers</h1>
      <Card className="overflow-x-auto p-0">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-line text-xs uppercase text-zinc-500">
            <tr>
              <th className="px-4 py-3">Worker</th>
              <th>Status</th>
              <th>Site</th>
              <th>GPU</th>
              <th>VRAM</th>
              <th>Util</th>
              <th>Models</th>
              <th>In-flight</th>
            </tr>
          </thead>
          <tbody>
            {(data ?? []).map((worker) => {
              const gpu = worker.resources.gpus[0];
              return (
                <tr key={worker.worker_id} className="border-b border-line/70">
                  <td className="px-4 py-3 font-mono">
                    <Link className="text-orange-400" to={`/workers/${worker.worker_id}`}>
                      {worker.worker_id}
                    </Link>
                  </td>
                  <td>
                    <Badge value={worker.status} />
                  </td>
                  <td>{worker.site ?? "—"}</td>
                  <td>{gpu?.name ?? "cpu-only"}</td>
                  <td>{gpu ? `${gpu.memory_used_mb}/${gpu.memory_total_mb} MB` : "—"}</td>
                  <td>{gpu ? `${gpu.utilization}%` : "—"}</td>
                  <td className="font-mono text-xs">
                    {worker.models.map((m) => m.name).join(", ") || "—"}
                  </td>
                  <td>{worker.running_requests}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
