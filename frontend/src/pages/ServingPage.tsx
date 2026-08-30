import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";
import { Boxes, Server } from "lucide-react";

export function ServingPage() {
  const nodes = useQuery({
    queryKey: ["serving", "nodes"],
    queryFn: api.servingNodes,
    staleTime: 15_000,
  });

  const models = useQuery({
    queryKey: ["serving", "models"],
    queryFn: api.servingModels,
    staleTime: 15_000,
  });

  const nodeList = nodes.data ?? [];
  const modelList = models.data ?? [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Serving Cluster Nodes</h1>
        <p className="text-sm text-slate-500">
          Cluster serving nodes evaluate runtime defenses under realistic multi-node LLM serving workloads.
        </p>
      </div>

      {/* Serving Nodes List */}
      <Card className="space-y-4 p-5">
        <div className="flex items-center justify-between border-b border-line-subtle pb-3">
          <div className="flex items-center gap-2">
            <Server className="h-5 w-5 text-accent" />
            <h2 className="text-sm font-bold text-slate-900">Registered Serving Nodes</h2>
          </div>
          <span className="font-mono text-xs text-slate-500">{nodeList.length} node(s)</span>
        </div>

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {nodeList.map((node) => {
            const workerId = String(node.worker_id ?? "");
            const status = String(node.status ?? "ONLINE");
            const isOnline = status === "ONLINE" || status === "READY";

            return (
              <div
                key={workerId}
                className="flex flex-col justify-between rounded-xl border border-line bg-slate-50/60 p-4 space-y-3"
              >
                <div className="flex items-start justify-between">
                  <div className="space-y-0.5">
                    <Link
                      to={`/workers/${workerId}`}
                      className="font-mono text-xs font-bold text-slate-900 hover:text-accent"
                    >
                      {workerId}
                    </Link>
                    <div className="text-[11px] text-slate-400">Serving Router Node</div>
                  </div>
                  <Badge value={status} showDot pulse={isOnline} />
                </div>
              </div>
            );
          })}
          {nodeList.length === 0 && (
            <div className="col-span-full py-8 text-center text-xs text-slate-400 font-sans">
              No serving nodes registered. Join a worker to start serving.
            </div>
          )}
        </div>
      </Card>

      {/* Loaded Models */}
      <Card className="space-y-4 p-5">
        <div className="flex items-center justify-between border-b border-line-subtle pb-3">
          <div className="flex items-center gap-2">
            <Boxes className="h-5 w-5 text-indigo-500" />
            <h2 className="text-sm font-bold text-slate-900">Loaded Models on Cluster</h2>
          </div>
          <span className="font-mono text-xs text-slate-500">{modelList.length} loaded</span>
        </div>

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {modelList.map((model, idx) => (
            <div
              key={`${model.worker_id}-${idx}`}
              className="flex items-center justify-between rounded-xl border border-line bg-white p-4 shadow-xs"
            >
              <div className="space-y-0.5">
                <div className="font-mono text-xs font-bold text-slate-900">{model.name}</div>
                <div className="text-[11px] text-slate-400 font-mono">Node: {model.worker_id}</div>
              </div>
              <Badge value={model.status || "READY"} showDot />
            </div>
          ))}
          {modelList.length === 0 && (
            <div className="col-span-full py-8 text-center text-xs text-slate-400 font-sans">
              No active models deployed across cluster nodes.
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}
