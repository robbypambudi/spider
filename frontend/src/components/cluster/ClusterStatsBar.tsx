import { Card } from "@/components/ui/card";
import type { WorkerView } from "@/types/api";
import { Activity, Cpu, HardDrive, Server } from "lucide-react";

interface ClusterStatsBarProps {
  workers: WorkerView[];
  isLoading?: boolean;
}

export function ClusterStatsBar({ workers, isLoading = false }: ClusterStatsBarProps) {
  const totalWorkers = workers.length;
  const onlineWorkers = workers.filter((w) => w.status === "ONLINE" || w.status === "READY").length;
  const offlineWorkers = workers.filter((w) => w.status === "OFFLINE").length;

  let totalGpus = 0;
  let totalVramMb = 0;
  let usedVramMb = 0;
  let totalCpu = 0;
  let totalRamMb = 0;
  let totalRequests = 0;

  for (const w of workers) {
    totalCpu += w.resources.cpu_total ?? 0;
    totalRamMb += w.resources.memory_total_mb ?? 0;
    totalRequests += w.running_requests ?? 0;
    for (const g of w.resources.gpus ?? []) {
      totalGpus += 1;
      totalVramMb += g.memory_total_mb ?? 0;
      usedVramMb += g.memory_used_mb ?? 0;
    }
  }

  const vramUsedGb = (usedVramMb / 1024).toFixed(1);
  const vramTotalGb = (totalVramMb / 1024).toFixed(1);
  const ramTotalGb = (totalRamMb / 1024).toFixed(1);
  const vramPct = totalVramMb > 0 ? Math.round((usedVramMb / totalVramMb) * 100) : 0;

  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-2 lg:grid-cols-4">
      {/* Workers overview */}
      <Card className="flex items-center gap-3.5 p-4">
        <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-blue-50 text-accent ring-1 ring-blue-100">
          <Server className="h-5 w-5" />
        </div>
        <div className="min-w-0">
          <div className="text-xs font-semibold uppercase tracking-wider text-slate-500">
            Cluster Nodes
          </div>
          <div className="mt-0.5 flex items-baseline gap-2">
            <span className="font-mono text-2xl font-bold text-slate-900">
              {isLoading ? "—" : totalWorkers}
            </span>
            <div className="flex items-center gap-1.5 text-xs">
              <span className="inline-flex items-center gap-1 font-medium text-emerald-600">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                {onlineWorkers} ok
              </span>
              {offlineWorkers > 0 && (
                <span className="inline-flex items-center gap-1 font-medium text-rose-600">
                  <span className="h-1.5 w-1.5 rounded-full bg-rose-500" />
                  {offlineWorkers} off
                </span>
              )}
            </div>
          </div>
        </div>
      </Card>

      {/* Compute Pool */}
      <Card className="flex items-center gap-3.5 p-4">
        <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-50 text-emerald-600 ring-1 ring-emerald-100">
          <Cpu className="h-5 w-5" />
        </div>
        <div className="min-w-0">
          <div className="text-xs font-semibold uppercase tracking-wider text-slate-500">
            Compute Pool
          </div>
          <div className="mt-0.5 flex items-baseline gap-2">
            <span className="font-mono text-2xl font-bold text-slate-900">
              {isLoading ? "—" : totalGpus > 0 ? `${totalGpus} GPUs` : `${totalCpu} vCPU`}
            </span>
            <span className="text-xs text-slate-500 font-mono">
              {totalGpus > 0 ? `${totalCpu} vCPU` : "CPU-only"}
            </span>
          </div>
        </div>
      </Card>

      {/* Memory / VRAM Utilization */}
      <Card className="flex items-center gap-3.5 p-4">
        <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 ring-1 ring-indigo-100">
          <HardDrive className="h-5 w-5" />
        </div>
        <div className="min-w-0">
          <div className="text-xs font-semibold uppercase tracking-wider text-slate-500">
            {totalGpus > 0 ? "Cluster VRAM" : "Cluster RAM Pool"}
          </div>
          <div className="mt-0.5 flex items-baseline gap-2">
            <span className="font-mono text-2xl font-bold text-slate-900">
              {isLoading
                ? "—"
                : totalGpus > 0
                ? `${vramUsedGb}/${vramTotalGb} GB`
                : `${ramTotalGb} GB`}
            </span>
            <span className="rounded bg-slate-100 px-1.5 py-0.5 text-[10px] font-mono text-slate-600">
              {totalGpus > 0 ? `${vramPct}%` : "System RAM"}
            </span>
          </div>
        </div>
      </Card>

      {/* In-Flight Serving Requests */}
      <Card className="flex items-center gap-3.5 p-4">
        <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-amber-50 text-amber-600 ring-1 ring-amber-100">
          <Activity className="h-5 w-5" />
        </div>
        <div className="min-w-0">
          <div className="text-xs font-semibold uppercase tracking-wider text-slate-500">
            Active Requests
          </div>
          <div className="mt-0.5 flex items-baseline gap-2">
            <span className="font-mono text-2xl font-bold text-slate-900">
              {isLoading ? "—" : totalRequests}
            </span>
            <span className="text-xs text-slate-500">
              {totalRequests === 0 ? "idle (0 in-flight)" : "in-flight"}
            </span>
          </div>
        </div>
      </Card>
    </div>
  );
}
