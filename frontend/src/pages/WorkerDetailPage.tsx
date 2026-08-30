import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Gauge } from "@/components/ui/gauge";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { api } from "@/services/api";
import {
  Activity,
  ArrowLeft,
  Boxes,
  Check,
  Cpu,
  Edit2,
  Globe,
  HardDrive,
  Layers,
  RefreshCw,
  Server,
  Trash2,
  X,
} from "lucide-react";

export function WorkerDetailPage() {
  const { workerId = "" } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [isEditing, setIsEditing] = useState(false);
  const [editHostname, setEditHostname] = useState("");
  const [editSite, setEditSite] = useState("");

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ["workers", workerId],
    queryFn: () => api.worker(workerId),
    enabled: Boolean(workerId),
    refetchInterval: 5000,
  });

  const updateMutation = useMutation({
    mutationFn: (body: { hostname?: string; site?: string }) =>
      api.updateWorker(workerId, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workers"] });
      setIsEditing(false);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => api.deleteWorker(workerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workers"] });
      navigate("/workers");
    },
  });

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center text-sm text-slate-500">
        <RefreshCw className="mr-2 h-4 w-4 animate-spin text-accent" />
        Loading worker telemetry...
      </div>
    );
  }

  if (!data) {
    return (
      <div className="space-y-4">
        <Link to="/workers" className="inline-flex items-center gap-1.5 text-xs text-slate-500 hover:text-slate-900">
          <ArrowLeft className="h-3.5 w-3.5" /> Back to workers
        </Link>
        <Card className="py-12 text-center text-slate-500">
          <Server className="mx-auto h-8 w-8 text-slate-300 mb-2" />
          <p className="font-semibold text-slate-800">Worker not found</p>
          <p className="text-xs text-slate-400">Worker ID {workerId} is not currently registered in the cluster.</p>
        </Card>
      </div>
    );
  }

  const isOnline = data.status === "ONLINE" || data.status === "READY";
  const gpus = data.resources.gpus ?? [];
  const ramGb = ((data.resources.memory_total_mb ?? 0) / 1024).toFixed(1);

  const handleOpenEdit = () => {
    setEditHostname(data.hostname);
    setEditSite(data.site ?? "");
    setIsEditing(true);
  };

  const handleSaveEdit = (e: React.FormEvent) => {
    e.preventDefault();
    updateMutation.mutate({ hostname: editHostname, site: editSite });
  };

  return (
    <div className="space-y-6">
      {/* Breadcrumb & Navigation */}
      <div className="flex items-center justify-between">
        <Link
          to="/workers"
          className="inline-flex items-center gap-1.5 text-xs font-medium text-slate-500 hover:text-slate-900 transition"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Back to cluster workers
        </Link>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleOpenEdit}
            icon={<Edit2 className="h-3.5 w-3.5" />}
          >
            Edit Details
          </Button>

          <Button
            variant="outline"
            size="sm"
            className="border-rose-200 text-rose-700 hover:bg-rose-50 hover:text-rose-800"
            onClick={() => {
              if (window.confirm(`Kick out / remove worker ${data.worker_id} from cluster?`)) {
                deleteMutation.mutate();
              }
            }}
            disabled={deleteMutation.isPending}
            icon={<Trash2 className="h-3.5 w-3.5" />}
          >
            {deleteMutation.isPending ? "Removing..." : "Kick Out Worker"}
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            icon={<RefreshCw className={`h-3.5 w-3.5 ${isFetching ? "animate-spin" : ""}`} />}
          >
            Refresh
          </Button>
        </div>
      </div>

      {/* Header Banner */}
      <Card className="flex flex-wrap items-center justify-between gap-4 p-6">
        <div className="flex items-center gap-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-blue-50 text-accent ring-1 ring-blue-100">
            <Server className="h-6 w-6" />
          </div>
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="font-mono text-xl font-bold text-slate-900">{data.worker_id}</h1>
              <Badge value={data.status} showDot pulse={isOnline} />
            </div>
            <div className="mt-1 flex flex-wrap items-center gap-3 text-xs text-slate-500">
              <span className="font-mono">{data.hostname}</span>
              {data.site && (
                <span className="inline-flex items-center gap-1 rounded bg-slate-100 px-2 py-0.5 font-medium text-slate-600">
                  <Globe className="h-3 w-3" />
                  {data.site}
                </span>
              )}
              <span className="font-mono text-slate-400">Agent v{data.version}</span>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-4 border-l border-line-subtle pl-4">
          <div className="text-right">
            <div className="text-[11px] uppercase tracking-wider text-slate-400 font-semibold">Active Workload</div>
            <div className="mt-0.5 flex items-center justify-end gap-1.5 font-mono text-lg font-bold text-slate-900">
              <Activity className="h-4 w-4 text-amber-500" />
              {data.running_requests} in-flight
            </div>
          </div>
        </div>
      </Card>

      {/* Edit Worker Modal */}
      {isEditing && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-xs p-4">
          <Card className="w-full max-w-md space-y-4 p-6 shadow-2xl">
            <div className="flex items-center justify-between border-b border-line-subtle pb-3">
              <div className="flex items-center gap-2">
                <Edit2 className="h-4 w-4 text-accent" />
                <h3 className="font-bold text-slate-900 text-sm">Edit Worker Details</h3>
              </div>
              <button
                onClick={() => setIsEditing(false)}
                className="text-slate-400 hover:text-slate-700"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handleSaveEdit} className="space-y-3 font-sans">
              <div>
                <label className="block text-xs font-semibold text-slate-500 uppercase">
                  Worker ID
                </label>
                <input
                  type="text"
                  disabled
                  value={data.worker_id}
                  className="mt-1 w-full rounded-lg border border-line bg-slate-100 p-2 font-mono text-xs text-slate-500 cursor-not-allowed"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase">
                  Hostname / Alias
                </label>
                <Input
                  type="text"
                  value={editHostname}
                  onChange={(e) => setEditHostname(e.target.value)}
                  placeholder="e.g. gpu-node-01 or my-laptop"
                  className="mt-1 font-mono text-xs"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase">
                  Cluster Site / Region Tag
                </label>
                <Input
                  type="text"
                  value={editSite}
                  onChange={(e) => setEditSite(e.target.value)}
                  placeholder="e.g. us-east-1, on-prem, lab-01"
                  className="mt-1 font-mono text-xs"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-2 border-t border-line-subtle">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setIsEditing(false)}
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  size="sm"
                  disabled={updateMutation.isPending}
                  icon={<Check className="h-3.5 w-3.5" />}
                >
                  {updateMutation.isPending ? "Saving..." : "Save Changes"}
                </Button>
              </div>
            </form>
          </Card>
        </div>
      )}

      {/* Hardware Telemetry Overview */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        {/* CPU Resource */}
        <Card className="flex items-center gap-4 p-5">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-blue-50 text-accent ring-1 ring-blue-100">
            <Cpu className="h-6 w-6" />
          </div>
          <div>
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-500">Processor Cores</div>
            <div className="mt-1 font-mono text-2xl font-bold text-slate-900">
              {data.resources.cpu_total} vCPU
            </div>
            <div className="text-[11px] text-slate-400">Allocated compute threads</div>
          </div>
        </Card>

        {/* RAM Resource */}
        <Card className="flex items-center gap-4 p-5">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 ring-1 ring-indigo-100">
            <HardDrive className="h-6 w-6" />
          </div>
          <div>
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-500">System Memory</div>
            <div className="mt-1 font-mono text-2xl font-bold text-slate-900">
              {ramGb} GB RAM
            </div>
            <div className="text-[11px] text-slate-400">{data.resources.memory_total_mb} MB total</div>
          </div>
        </Card>

        {/* GPU Attached */}
        <Card className="flex items-center gap-4 p-5">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-emerald-50 text-emerald-600 ring-1 ring-emerald-100">
            <Layers className="h-6 w-6" />
          </div>
          <div>
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-500">Accelerators</div>
            <div className="mt-1 font-mono text-2xl font-bold text-slate-900">
              {gpus.length} GPUs
            </div>
            <div className="text-[11px] text-slate-400">
              {gpus.length > 0 ? gpus[0].name : "CPU-only node"}
            </div>
          </div>
        </Card>
      </div>

      {/* GPU Detailed Breakdown */}
      <Card className="space-y-4">
        <div className="flex items-center justify-between border-b border-line-subtle pb-3">
          <div className="flex items-center gap-2">
            <Cpu className="h-5 w-5 text-accent" />
            <h2 className="text-sm font-bold text-slate-900">Attached GPU Hardware & Telemetry</h2>
          </div>
          <span className="font-mono text-xs text-slate-500">{gpus.length} accelerator(s) detected</span>
        </div>

        {gpus.length === 0 ? (
          <div className="rounded-lg border border-dashed border-line p-6 text-center text-xs text-slate-500">
            This node operates in CPU-only mode. GPU accelerators are optional for inference and sidecar detection.
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            {gpus.map((gpu) => {
              const vramUsedGb = (gpu.memory_used_mb / 1024).toFixed(1);
              const vramTotalGb = (gpu.memory_total_mb / 1024).toFixed(1);
              const vramPct =
                gpu.memory_total_mb > 0
                  ? Math.round((gpu.memory_used_mb / gpu.memory_total_mb) * 100)
                  : 0;

              return (
                <div
                  key={gpu.index}
                  className="flex flex-col justify-between rounded-xl border border-line-subtle bg-slate-50/60 p-5 space-y-4"
                >
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="flex items-center gap-2 font-mono text-sm font-bold text-slate-900">
                        <span>GPU [{gpu.index}]</span>
                        <span className="rounded bg-blue-100 px-1.5 py-0.2 text-[10px] text-accent">
                          {gpu.vendor || "NVIDIA"}
                        </span>
                      </div>
                      <div className="text-xs text-slate-600 font-medium">{gpu.name}</div>
                    </div>

                    <Gauge
                      value={gpu.utilization}
                      size={64}
                      strokeWidth={6}
                      sublabel="Util"
                      colorVariant="dynamic"
                    />
                  </div>

                  <div className="space-y-1.5 border-t border-line-subtle pt-3">
                    <div className="flex justify-between text-xs font-medium">
                      <span className="text-slate-500">Dedicated VRAM</span>
                      <span className="font-mono text-slate-800">
                        {vramUsedGb} GB / {vramTotalGb} GB ({vramPct}%)
                      </span>
                    </div>
                    <Progress value={vramPct} colorVariant="dynamic" className="h-2.5" />
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </Card>

      {/* Loaded Models */}
      <Card className="space-y-4">
        <div className="flex items-center justify-between border-b border-line-subtle pb-3">
          <div className="flex items-center gap-2">
            <Boxes className="h-5 w-5 text-indigo-500" />
            <h2 className="text-sm font-bold text-slate-900">Loaded Serving Models</h2>
          </div>
          <span className="font-mono text-xs text-slate-500">{data.models.length} model(s) loaded</span>
        </div>

        {data.models.length === 0 ? (
          <div className="rounded-lg border border-dashed border-line p-6 text-center text-xs text-slate-500">
            No models are currently loaded on this worker node.
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {data.models.map((model) => (
              <div
                key={model.name}
                className="flex items-center justify-between rounded-lg border border-line-subtle bg-white p-3.5 shadow-xs"
              >
                <div className="space-y-0.5">
                  <div className="font-mono text-xs font-bold text-slate-900">{model.name}</div>
                  <div className="text-[11px] text-slate-400">Serving active</div>
                </div>
                <Badge value={model.status || "READY"} showDot />
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  );
}
