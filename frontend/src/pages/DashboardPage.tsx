import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { ClusterStatsBar } from "@/components/cluster/ClusterStatsBar";
import { TrafficChart } from "@/components/charts/TrafficChart";
import { LatencyChart } from "@/components/charts/LatencyChart";
import { DetectorPieChart } from "@/components/charts/DetectorPieChart";
import { api } from "@/services/api";
import { formatMs, formatPct } from "@/lib/utils";
import {
  Activity,
  ArrowRight,
  Clock,
  Flame,
  Percent,
  RefreshCw,
  ShieldAlert,
  ShieldCheck,
} from "lucide-react";

export function DashboardPage() {
  const summary = useQuery({
    queryKey: ["metrics", "summary"],
    queryFn: api.summary,
    staleTime: 10_000,
    refetchInterval: 5_000, // auto refresh every 5s
  });

  const workers = useQuery({
    queryKey: ["workers"],
    queryFn: api.workers,
    staleTime: 10_000,
    refetchInterval: 5_000,
  });

  const scans = useQuery({
    queryKey: ["security", "scans"],
    queryFn: api.scans,
    staleTime: 10_000,
    refetchInterval: 5_000,
  });

  const data = summary.data;
  const recentScans = (scans.data ?? []).slice(0, 5);

  return (
    <div className="space-y-7">
      {/* Page Header */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Security Command Center</h1>
          <p className="text-sm text-slate-500">
            Real-time prompt injection defense telemetry & cluster workload monitoring.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              summary.refetch();
              workers.refetch();
              scans.refetch();
            }}
            icon={<RefreshCw className={`h-3.5 w-3.5 ${summary.isFetching ? "animate-spin" : ""}`} />}
          >
            Refresh
          </Button>
          <Link to="/security">
            <Button size="sm" icon={<ShieldCheck className="h-4 w-4" />}>
              Live Scanner
            </Button>
          </Link>
        </div>
      </div>

      {/* Cluster Overview Bar */}
      <ClusterStatsBar workers={workers.data ?? []} isLoading={workers.isLoading} />

      {/* KPI Metric Cards */}
      <div className="grid grid-cols-2 gap-4 md:grid-cols-3 xl:grid-cols-6">
        <Card hoverable className="p-4 space-y-1">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <span>Total Scans</span>
            <Activity className="h-4 w-4 text-blue-500" />
          </div>
          <div className="font-mono text-2xl font-bold text-slate-900">
            {data?.total_scans ?? 0}
          </div>
          <div className="text-[11px] text-slate-400 font-mono">requests filtered</div>
        </Card>

        <Card hoverable className="p-4 space-y-1">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <span>Blocked</span>
            <ShieldAlert className="h-4 w-4 text-rose-500" />
          </div>
          <div className="font-mono text-2xl font-bold text-rose-600">
            {data?.blocked ?? 0}
          </div>
          <div className="text-[11px] text-rose-600 font-medium font-mono">injections stopped</div>
        </Card>

        <Card hoverable className="p-4 space-y-1">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <span>Allowed</span>
            <ShieldCheck className="h-4 w-4 text-emerald-500" />
          </div>
          <div className="font-mono text-2xl font-bold text-emerald-600">
            {data?.allowed ?? 0}
          </div>
          <div className="text-[11px] text-emerald-600 font-medium font-mono">benign prompts</div>
        </Card>

        <Card hoverable className="p-4 space-y-1">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <span>Detection Rate</span>
            <Percent className="h-4 w-4 text-indigo-500" />
          </div>
          <div className="font-mono text-2xl font-bold text-slate-900">
            {formatPct(data?.detection_rate)}
          </div>
          <div className="text-[11px] text-slate-400 font-mono">positive ratio</div>
        </Card>

        <Card hoverable className="p-4 space-y-1">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <span>P95 Latency</span>
            <Clock className="h-4 w-4 text-amber-500" />
          </div>
          <div className="font-mono text-2xl font-bold text-slate-900">
            {formatMs(data?.p95_detection_latency_ms)}
          </div>
          <div className="text-[11px] text-slate-400 font-mono">
            avg {formatMs(data?.avg_detection_latency_ms)}
          </div>
        </Card>

        <Card hoverable className="p-4 space-y-1">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <span>Sec Overhead</span>
            <Flame className="h-4 w-4 text-teal-500" />
          </div>
          <div className="font-mono text-2xl font-bold text-slate-900">
            {formatMs(data?.avg_security_overhead_ms)}
          </div>
          <div className="text-[11px] text-slate-400 font-mono">added pipeline latency</div>
        </Card>
      </div>

      {/* Main Charts Row */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <TrafficChart
          className="lg:col-span-2"
          title="Cluster Defense Activity & Request Volume"
        />
        <DetectorPieChart
          allowed={data?.allowed ?? 0}
          blocked={data?.blocked ?? 0}
          review={data?.review ?? 0}
        />
      </div>

      {/* Bottom Row: Latency Breakdown & Recent Scan Activity */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <LatencyChart
          avgDetectionMs={data?.avg_detection_latency_ms ?? 0}
          p95DetectionMs={data?.p95_detection_latency_ms ?? 0}
          avgOverheadMs={data?.avg_security_overhead_ms ?? 0}
        />

        {/* Live Scan Activity Stream */}
        <Card className="lg:col-span-2 p-5 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-blue-50 text-accent ring-1 ring-blue-100">
                <Activity className="h-4 w-4" />
              </div>
              <div>
                <h3 className="text-sm font-semibold text-slate-900">Recent Security Scan Events</h3>
                <p className="text-xs text-slate-500">Live stream of analyzed prompts & enforcement outcomes</p>
              </div>
            </div>

            <Link
              to="/security/scans"
              className="flex items-center gap-1 text-xs font-medium text-accent hover:underline"
            >
              View all
              <ArrowRight className="h-3.5 w-3.5" />
            </Link>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="border-b border-line bg-slate-50/70 text-[11px] uppercase tracking-wider text-slate-500 font-semibold">
                <tr>
                  <th className="px-3 py-2.5 font-semibold">Decision</th>
                  <th className="px-3 py-2.5 font-semibold">Threat Score</th>
                  <th className="px-3 py-2.5 font-semibold">Latency</th>
                  <th className="px-3 py-2.5 font-semibold">Policy</th>
                  <th className="px-3 py-2.5 font-semibold">Timestamp</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-line-subtle font-mono">
                {recentScans.map((scan) => (
                  <tr key={scan.id} className="transition hover:bg-slate-50/60 font-mono">
                    <td className="px-3 py-2.5">
                      <Badge value={scan.decision} showDot />
                    </td>
                    <td className="px-3 py-2.5 font-bold">
                      <span className={scan.decision === "BLOCK" ? "text-rose-600" : "text-emerald-600"}>
                        {scan.score.toFixed(3)}
                      </span>
                    </td>
                    <td className="px-3 py-2.5 text-slate-600">
                      {formatMs(scan.latency_ms)}
                    </td>
                    <td className="px-3 py-2.5 text-slate-600">
                      {scan.policy || "default"}
                    </td>
                    <td className="px-3 py-2.5 text-slate-400 text-[11px]">
                      {new Date(scan.created_at).toLocaleTimeString()}
                    </td>
                  </tr>
                ))}
                {recentScans.length === 0 && (
                  <tr>
                    <td colSpan={5} className="py-8 text-center text-slate-400 font-sans">
                      No security scans recorded yet. Use the Security Scanner to inspect prompts.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </Card>
      </div>
    </div>
  );
}
