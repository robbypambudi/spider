import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Tabs } from "@/components/ui/tabs";
import { ClusterStatsBar } from "@/components/cluster/ClusterStatsBar";
import { WorkerCard } from "@/components/cluster/WorkerCard";
import { api } from "@/services/api";
import type { WorkerView } from "@/types/api";
import {
  Activity,
  Check,
  Edit2,
  LayoutGrid,
  List,
  RefreshCw,
  Search,
  Server,
  Trash2,
  UserX,
  X,
} from "lucide-react";

export function WorkersPage() {
  const queryClient = useQueryClient();
  const [viewMode, setViewMode] = useState<"grid" | "table">("grid");
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("ALL");
  const [refreshInterval, setRefreshInterval] = useState<number>(5000); // 5s default

  // Edit Modal State
  const [editingWorker, setEditingWorker] = useState<WorkerView | null>(null);
  const [editHostname, setEditHostname] = useState("");
  const [editSite, setEditSite] = useState("");

  // Delete / Kick confirmation state
  const [deletingWorker, setDeletingWorker] = useState<WorkerView | null>(null);

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ["workers"],
    queryFn: api.workers,
    refetchInterval: refreshInterval > 0 ? refreshInterval : false,
  });

  const deleteMutation = useMutation({
    mutationFn: (workerId: string) => api.deleteWorker(workerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workers"] });
      queryClient.invalidateQueries({ queryKey: ["metrics"] });
      setDeletingWorker(null);
    },
  });

  const pruneMutation = useMutation({
    mutationFn: api.pruneOfflineWorkers,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workers"] });
      queryClient.invalidateQueries({ queryKey: ["metrics"] });
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ workerId, body }: { workerId: string; body: { hostname?: string; site?: string } }) =>
      api.updateWorker(workerId, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workers"] });
      setEditingWorker(null);
    },
  });

  const workers = data ?? [];
  const offlineCount = workers.filter((w) => w.status === "OFFLINE").length;

  const handleOpenEdit = (w: WorkerView) => {
    setEditingWorker(w);
    setEditHostname(w.hostname);
    setEditSite(w.site ?? "");
  };

  const handleSaveEdit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingWorker) return;
    updateMutation.mutate({
      workerId: editingWorker.worker_id,
      body: { hostname: editHostname, site: editSite },
    });
  };

  // Filtering
  const filteredWorkers = workers.filter((w) => {
    const matchesSearch =
      w.worker_id.toLowerCase().includes(searchQuery.toLowerCase()) ||
      w.hostname.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (w.site && w.site.toLowerCase().includes(searchQuery.toLowerCase()));

    const matchesStatus =
      statusFilter === "ALL" ||
      w.status === statusFilter ||
      (statusFilter === "ONLINE" && (w.status === "ONLINE" || w.status === "READY"));

    return matchesSearch && matchesStatus;
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Cluster Workers</h1>
            <span className="rounded-full bg-slate-100 px-2.5 py-0.5 text-xs font-semibold text-slate-600 ring-1 ring-slate-200">
              {workers.length} nodes
            </span>
          </div>
          <p className="text-sm text-slate-500">
            Real-time telemetry, GPU utilization, and serving capacity across the cluster.
          </p>
        </div>

        {/* Real-time controls & Actions */}
        <div className="flex flex-wrap items-center gap-3">
          {/* Prune All Offline Button */}
          {offlineCount > 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                if (window.confirm(`Kick out / purge all ${offlineCount} offline workers from the cluster?`)) {
                  pruneMutation.mutate();
                }
              }}
              disabled={pruneMutation.isPending}
              className="border-rose-200 text-rose-700 hover:bg-rose-50 hover:text-rose-800"
              icon={<UserX className="h-3.5 w-3.5" />}
            >
              {pruneMutation.isPending ? "Pruning..." : `Kick Out Offline (${offlineCount})`}
            </Button>
          )}

          {/* Auto Refresh dropdown */}
          <div className="flex items-center gap-1.5 rounded-lg border border-line bg-white px-2.5 py-1.5 text-xs">
            <span className="relative flex h-2 w-2">
              <span
                className={`absolute inline-flex h-full w-full rounded-full opacity-75 ${
                  refreshInterval > 0 ? "animate-ping bg-emerald-400" : "bg-slate-300"
                }`}
              />
              <span
                className={`relative inline-flex h-2 w-2 rounded-full ${
                  refreshInterval > 0 ? "bg-emerald-500" : "bg-slate-400"
                }`}
              />
            </span>
            <span className="text-slate-500 font-medium">Auto-poll:</span>
            <select
              value={refreshInterval}
              onChange={(e) => setRefreshInterval(Number(e.target.value))}
              className="bg-transparent font-medium text-slate-800 outline-none cursor-pointer"
            >
              <option value={2000}>2s (Fast)</option>
              <option value={5000}>5s (Default)</option>
              <option value={10000}>10s</option>
              <option value={0}>Off</option>
            </select>
          </div>

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

      {/* Cluster Overview Stats Bar */}
      <ClusterStatsBar workers={workers} isLoading={isLoading} />

      {/* Filter and View Mode Controls */}
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-line bg-white p-3 shadow-card">
        <div className="flex flex-1 flex-wrap items-center gap-3 min-w-[280px]">
          {/* Search Box */}
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              placeholder="Search by worker ID, hostname, site..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full rounded-lg border border-line bg-slate-50/50 pl-8 pr-3 py-1.5 text-xs text-slate-900 placeholder:text-slate-400 outline-none focus:border-accent focus:bg-white focus:ring-1 focus:ring-accent"
            />
          </div>

          {/* Status Filter Tabs */}
          <Tabs
            tabs={[
              { id: "ALL", label: "All", count: workers.length },
              {
                id: "ONLINE",
                label: "Online",
                count: workers.filter((w) => w.status === "ONLINE" || w.status === "READY").length,
              },
              {
                id: "BUSY",
                label: "Busy",
                count: workers.filter((w) => w.status === "BUSY").length,
              },
              {
                id: "OFFLINE",
                label: "Offline",
                count: offlineCount,
              },
            ]}
            activeTab={statusFilter}
            onChange={setStatusFilter}
          />
        </div>

        {/* View Switcher: Grid vs Table */}
        <div className="flex items-center rounded-lg border border-line bg-slate-50 p-0.5">
          <button
            type="button"
            onClick={() => setViewMode("grid")}
            className={`flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs font-medium transition ${
              viewMode === "grid"
                ? "bg-white text-slate-900 shadow-xs ring-1 ring-slate-200"
                : "text-slate-500 hover:text-slate-900"
            }`}
          >
            <LayoutGrid className="h-3.5 w-3.5" />
            Grid
          </button>
          <button
            type="button"
            onClick={() => setViewMode("table")}
            className={`flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs font-medium transition ${
              viewMode === "table"
                ? "bg-white text-slate-900 shadow-xs ring-1 ring-slate-200"
                : "text-slate-500 hover:text-slate-900"
            }`}
          >
            <List className="h-3.5 w-3.5" />
            Table
          </button>
        </div>
      </div>

      {/* Workers Render */}
      {viewMode === "grid" ? (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {filteredWorkers.map((worker) => (
            <WorkerCard
              key={worker.worker_id}
              worker={worker}
              onEdit={handleOpenEdit}
              onDelete={(w) => setDeletingWorker(w)}
            />
          ))}
          {filteredWorkers.length === 0 && !isLoading && (
            <div className="col-span-full py-16 text-center text-slate-400">
              <Server className="mx-auto h-8 w-8 text-slate-300 mb-2" />
              <p className="font-medium text-slate-600">No workers match your filter criteria.</p>
              <p className="text-xs text-slate-400 mt-1">
                Start a local worker using <code className="font-mono bg-slate-100 px-1 py-0.5 rounded">spider worker join</code>.
              </p>
            </div>
          )}
        </div>
      ) : (
        <Card className="overflow-x-auto p-0">
          <table className="w-full text-left text-xs text-slate-700">
            <thead className="border-b border-line bg-slate-50 text-[11px] uppercase tracking-wider text-slate-500 font-semibold">
              <tr>
                <th className="px-4 py-3">Worker ID</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Site / Host</th>
                <th className="px-4 py-3">Compute / RAM</th>
                <th className="px-4 py-3">GPU & VRAM</th>
                <th className="px-4 py-3">Util</th>
                <th className="px-4 py-3">Serving Models</th>
                <th className="px-4 py-3">In-flight</th>
                <th className="px-4 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line-subtle font-mono">
              {filteredWorkers.map((worker) => {
                const gpu = (worker.resources.gpus ?? [])[0];
                const ramGb = ((worker.resources.memory_total_mb ?? 0) / 1024).toFixed(1);
                const vramPct =
                  gpu && gpu.memory_total_mb > 0
                    ? Math.round((gpu.memory_used_mb / gpu.memory_total_mb) * 100)
                    : 0;

                return (
                  <tr key={worker.worker_id} className="transition hover:bg-slate-50/70">
                    <td className="px-4 py-3 font-bold">
                      <Link
                        className="text-accent hover:underline flex items-center gap-1.5"
                        to={`/workers/${worker.worker_id}`}
                      >
                        <Server className="h-3.5 w-3.5 text-slate-400 shrink-0" />
                        {worker.worker_id}
                      </Link>
                    </td>
                    <td className="px-4 py-3">
                      <Badge value={worker.status} showDot pulse={worker.status === "ONLINE"} />
                    </td>
                    <td className="px-4 py-3 text-slate-600">
                      <div className="font-sans font-medium text-slate-900">{worker.site ?? "default"}</div>
                      <div className="text-[10px] text-slate-400">{worker.hostname}</div>
                    </td>
                    <td className="px-4 py-3 text-slate-600">
                      <div>{worker.resources.cpu_total} vCPU</div>
                      <div className="text-[11px] text-slate-400">{ramGb} GB RAM</div>
                    </td>
                    <td className="px-4 py-3 min-w-[160px]">
                      {gpu ? (
                        <div className="space-y-1">
                          <div className="text-[11px] text-slate-700">{gpu.name}</div>
                          <div className="flex items-center gap-2">
                            <Progress value={vramPct} className="h-1.5 flex-1" />
                            <span className="text-[10px] text-slate-500 font-mono">
                              {Math.round(gpu.memory_used_mb / 1024)}/{Math.round(gpu.memory_total_mb / 1024)}GB
                            </span>
                          </div>
                        </div>
                      ) : (
                        <span className="text-slate-400 italic font-sans text-xs">CPU only</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {gpu ? (
                        <span className="rounded bg-slate-100 px-1.5 py-0.5 text-xs font-bold text-slate-800">
                          {gpu.utilization}%
                        </span>
                      ) : (
                        "—"
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {worker.models.map((m) => (
                          <span
                            key={m.name}
                            className="rounded bg-blue-50 px-1.5 py-0.5 text-[10px] text-accent ring-1 ring-blue-200"
                          >
                            {m.name}
                          </span>
                        ))}
                        {worker.models.length === 0 && <span className="text-slate-400 italic font-sans">—</span>}
                      </div>
                    </td>
                    <td className="px-4 py-3 font-bold text-slate-900">
                      <div className="flex items-center gap-1">
                        <Activity className="h-3 w-3 text-amber-500" />
                        {worker.running_requests}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-1 font-sans">
                        <button
                          type="button"
                          onClick={() => handleOpenEdit(worker)}
                          title="Edit Worker Info"
                          className="flex h-7 w-7 items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-700 transition"
                        >
                          <Edit2 className="h-3.5 w-3.5" />
                        </button>
                        <button
                          type="button"
                          onClick={() => setDeletingWorker(worker)}
                          title="Kick out / Remove worker"
                          className="flex h-7 w-7 items-center justify-center rounded-lg text-slate-400 hover:bg-rose-50 hover:text-rose-600 transition"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}
              {filteredWorkers.length === 0 && (
                <tr>
                  <td colSpan={9} className="py-8 text-center text-slate-400 font-sans">
                    No workers match your filter criteria.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </Card>
      )}

      {/* Edit Worker Modal */}
      {editingWorker && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-xs p-4">
          <Card className="w-full max-w-md space-y-4 p-6 shadow-2xl">
            <div className="flex items-center justify-between border-b border-line-subtle pb-3">
              <div className="flex items-center gap-2">
                <Edit2 className="h-4 w-4 text-accent" />
                <h3 className="font-bold text-slate-900 text-sm">Edit Worker Details</h3>
              </div>
              <button
                onClick={() => setEditingWorker(null)}
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
                  value={editingWorker.worker_id}
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
                  onClick={() => setEditingWorker(null)}
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

      {/* Kick Out Confirmation Modal */}
      {deletingWorker && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-xs p-4">
          <Card className="w-full max-w-md space-y-4 p-6 shadow-2xl">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-rose-50 text-rose-600 ring-1 ring-rose-200">
                <Trash2 className="h-5 w-5" />
              </div>
              <div>
                <h3 className="font-bold text-slate-900 text-sm">Kick Out Worker Node</h3>
                <p className="text-xs text-slate-500">
                  Are you sure you want to remove <span className="font-mono font-bold text-slate-800">{deletingWorker.worker_id}</span> from the cluster?
                </p>
              </div>
            </div>

            <p className="text-xs text-slate-600 rounded-lg bg-slate-50 p-3">
              This will remove its registration and historical telemetry. If the worker agent is still running, it will re-register on its next join.
            </p>

            <div className="flex items-center justify-end gap-2 pt-2 border-t border-line-subtle">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setDeletingWorker(null)}
              >
                Cancel
              </Button>
              <Button
                type="button"
                variant="danger"
                size="sm"
                disabled={deleteMutation.isPending}
                onClick={() => deleteMutation.mutate(deletingWorker.worker_id)}
                icon={<Trash2 className="h-3.5 w-3.5" />}
              >
                {deleteMutation.isPending ? "Removing..." : "Kick Out Worker"}
              </Button>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}
