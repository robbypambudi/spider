import { Progress } from "@/components/ui/progress";
import type { GPUResource } from "@/types/api";
import { Cpu } from "lucide-react";

export function GpuResourceBar({ gpu }: { gpu: GPUResource }) {
  const vramPercent =
    gpu.memory_total_mb > 0
      ? Math.round((gpu.memory_used_mb / gpu.memory_total_mb) * 100)
      : 0;

  const usedGb = (gpu.memory_used_mb / 1024).toFixed(1);
  const totalGb = (gpu.memory_total_mb / 1024).toFixed(1);

  return (
    <div className="space-y-2 rounded-lg border border-line-subtle bg-slate-50/70 p-3">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-1.5 font-medium text-slate-800">
          <Cpu className="h-4 w-4 text-accent" />
          <span className="text-xs font-semibold">
            GPU [{gpu.index}] {gpu.name}
          </span>
        </div>
        <span className="rounded bg-white px-1.5 py-0.5 font-mono text-[11px] font-semibold text-slate-700 ring-1 ring-slate-200">
          Util: {gpu.utilization}%
        </span>
      </div>

      <div className="space-y-1">
        <div className="flex justify-between text-[11px] font-medium text-slate-500">
          <span>VRAM Usage</span>
          <span className="font-mono">
            {usedGb} GB / {totalGb} GB ({vramPercent}%)
          </span>
        </div>
        <Progress value={vramPercent} colorVariant="dynamic" className="h-2" />
      </div>
    </div>
  );
}
