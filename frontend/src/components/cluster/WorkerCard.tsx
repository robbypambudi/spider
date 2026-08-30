import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { GpuResourceBar } from "./GpuResourceBar";
import type { WorkerView } from "@/types/api";
import {
  Activity,
  ArrowUpRight,
  Boxes,
  Cpu,
  Edit2,
  Globe,
  HardDrive,
  Layers,
  Server,
  Trash2,
} from "lucide-react";

interface WorkerCardProps {
  worker: WorkerView;
  onEdit?: (worker: WorkerView) => void;
  onDelete?: (worker: WorkerView) => void;
}

export function WorkerCard({ worker, onEdit, onDelete }: WorkerCardProps) {
  const isOnline = worker.status === "ONLINE" || worker.status === "READY";
  const gpus = worker.resources.gpus ?? [];
  const ramMb = worker.resources.memory_total_mb ?? 0;
  const ramGb = (ramMb / 1024).toFixed(1);

  return (
    <Card hoverable className="flex flex-col justify-between space-y-4 p-5">
      {/* Card Header */}
      <div className="flex items-start justify-between gap-3 border-b border-line-subtle pb-3.5">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Server className="h-4 w-4 text-slate-500" />
            <Link
              to={`/workers/${worker.worker_id}`}
              className="group flex items-center gap-1 font-mono text-sm font-bold text-slate-900 transition hover:text-accent"
            >
              <span>{worker.worker_id}</span>
              <ArrowUpRight className="h-3.5 w-3.5 opacity-0 transition-opacity group-hover:opacity-100" />
            </Link>
          </div>
          <div className="flex flex-wrap items-center gap-2 text-xs text-slate-500">
            <span className="font-mono">{worker.hostname}</span>
            {worker.site && (
              <span className="inline-flex items-center gap-1 rounded bg-slate-100 px-1.5 py-0.5 text-[11px] font-medium text-slate-600">
                <Globe className="h-3 w-3" />
                {worker.site}
              </span>
            )}
            <span className="font-mono text-[10px] text-slate-400">v{worker.version}</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Badge
            value={worker.status}
            showDot
            pulse={isOnline}
            className="shrink-0"
          />

          {/* Quick Actions (Edit & Kick Out) */}
          <div className="flex items-center gap-1">
            {onEdit && (
              <button
                type="button"
                onClick={() => onEdit(worker)}
                title="Edit Worker Info"
                className="flex h-7 w-7 items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-700 transition"
              >
                <Edit2 className="h-3.5 w-3.5" />
              </button>
            )}
            {onDelete && (
              <button
                type="button"
                onClick={() => onDelete(worker)}
                title="Kick out / Remove worker"
                className="flex h-7 w-7 items-center justify-center rounded-lg text-slate-400 hover:bg-rose-50 hover:text-rose-600 transition"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Hardware Resources */}
      <div className="space-y-3">
        {/* CPU & RAM */}
        <div className="grid grid-cols-2 gap-2 text-xs">
          <div className="flex items-center gap-2 rounded-lg bg-slate-50 p-2 text-slate-700">
            <Cpu className="h-4 w-4 text-slate-400" />
            <div>
              <div className="text-[10px] uppercase text-slate-400 font-semibold">Compute</div>
              <div className="font-mono font-semibold">{worker.resources.cpu_total} vCPU</div>
            </div>
          </div>
          <div className="flex items-center gap-2 rounded-lg bg-slate-50 p-2 text-slate-700">
            <HardDrive className="h-4 w-4 text-slate-400" />
            <div>
              <div className="text-[10px] uppercase text-slate-400 font-semibold">Memory</div>
              <div className="font-mono font-semibold">{ramGb} GB RAM</div>
            </div>
          </div>
        </div>

        {/* GPUs */}
        {gpus.length > 0 ? (
          <div className="space-y-2">
            {gpus.map((gpu) => (
              <GpuResourceBar key={gpu.index} gpu={gpu} />
            ))}
          </div>
        ) : (
          <div className="flex items-center gap-2 rounded-lg border border-dashed border-line p-2.5 text-xs text-slate-500">
            <Layers className="h-4 w-4 text-slate-400" />
            <span>CPU-only worker node (No GPU attached)</span>
          </div>
        )}
      </div>

      {/* Loaded Models & Serving Status */}
      <div className="space-y-2 border-t border-line-subtle pt-3">
        <div className="flex items-center justify-between text-xs">
          <span className="flex items-center gap-1 font-semibold uppercase tracking-wider text-slate-400 text-[10px]">
            <Boxes className="h-3 w-3" />
            Serving Models
          </span>
          <span className="flex items-center gap-1 font-mono text-[11px] text-slate-600">
            <Activity className="h-3 w-3 text-amber-500" />
            {worker.running_requests} in-flight
          </span>
        </div>

        <div className="flex flex-wrap gap-1.5">
          {worker.models.length > 0 ? (
            worker.models.map((model) => (
              <span
                key={model.name}
                className="inline-flex items-center gap-1 rounded-md bg-blue-50/80 px-2 py-0.5 font-mono text-[11px] font-medium text-accent ring-1 ring-blue-200/60"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-accent" />
                {model.name}
              </span>
            ))
          ) : (
            <span className="text-xs text-slate-400 italic">No models loaded</span>
          )}
        </div>
      </div>
    </Card>
  );
}
