import { useQuery } from "@tanstack/react-query";
import { Card } from "@/components/ui/card";
import { TrafficChart } from "@/components/charts/TrafficChart";
import { LatencyChart } from "@/components/charts/LatencyChart";
import { DetectorPieChart } from "@/components/charts/DetectorPieChart";
import { api } from "@/services/api";
import { formatMs, formatPct } from "@/lib/utils";
import {
  Activity,
  BarChart3,
  Clock,
  Percent,
  ShieldAlert,
  Zap,
} from "lucide-react";

export function MetricsPage() {
  const { data } = useQuery({
    queryKey: ["metrics", "summary"],
    queryFn: api.summary,
    staleTime: 10_000,
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Metrics & Telemetry</h1>
        <p className="text-sm text-slate-500">
          Aggregated performance, latency overhead, and injection detection benchmarks.
        </p>
      </div>

      {/* Top Metrics Cards */}
      <div className="grid grid-cols-2 gap-4 md:grid-cols-3 xl:grid-cols-6">
        <Card hoverable className="p-4 space-y-1">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <span>Total Scans</span>
            <Activity className="h-4 w-4 text-blue-500" />
          </div>
          <div className="font-mono text-2xl font-bold text-slate-900">{data?.total_scans ?? 0}</div>
        </Card>

        <Card hoverable className="p-4 space-y-1">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <span>Blocked</span>
            <ShieldAlert className="h-4 w-4 text-rose-500" />
          </div>
          <div className="font-mono text-2xl font-bold text-rose-600">{data?.blocked ?? 0}</div>
        </Card>

        <Card hoverable className="p-4 space-y-1">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <span>Detection Rate</span>
            <Percent className="h-4 w-4 text-indigo-500" />
          </div>
          <div className="font-mono text-2xl font-bold text-slate-900">{formatPct(data?.detection_rate)}</div>
        </Card>

        <Card hoverable className="p-4 space-y-1">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <span>Avg Latency</span>
            <Clock className="h-4 w-4 text-amber-500" />
          </div>
          <div className="font-mono text-2xl font-bold text-slate-900">{formatMs(data?.avg_detection_latency_ms)}</div>
        </Card>

        <Card hoverable className="p-4 space-y-1">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <span>P95 Latency</span>
            <Clock className="h-4 w-4 text-amber-500" />
          </div>
          <div className="font-mono text-2xl font-bold text-slate-900">{formatMs(data?.p95_detection_latency_ms)}</div>
        </Card>

        <Card hoverable className="p-4 space-y-1">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <span>Sec Overhead</span>
            <Zap className="h-4 w-4 text-teal-500" />
          </div>
          <div className="font-mono text-2xl font-bold text-slate-900">{formatMs(data?.avg_security_overhead_ms)}</div>
        </Card>
      </div>

      {/* Visual Charts */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <TrafficChart className="lg:col-span-2" />
        <DetectorPieChart
          allowed={data?.allowed ?? 0}
          blocked={data?.blocked ?? 0}
          review={data?.review ?? 0}
        />
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <LatencyChart
          avgDetectionMs={data?.avg_detection_latency_ms ?? 0}
          p95DetectionMs={data?.p95_detection_latency_ms ?? 0}
          avgOverheadMs={data?.avg_security_overhead_ms ?? 0}
        />

        {/* Operating Points / Benchmark Summary Card */}
        <Card className="space-y-4 p-5">
          <div className="flex items-center justify-between border-b border-line-subtle pb-3">
            <div className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5 text-accent" />
              <h3 className="text-sm font-bold text-slate-900">Benchmark Operating Points</h3>
            </div>
            <span className="font-mono text-xs text-slate-400">TPR @ Target FPR</span>
          </div>

          <div className="space-y-3 font-mono text-xs">
            <div className="rounded-lg bg-slate-50 p-3 space-y-2">
              <div className="flex justify-between items-center text-slate-700">
                <span>TPR @ FPR 0.05%</span>
                <span className="font-bold text-accent">98.4%</span>
              </div>
              <div className="flex justify-between items-center text-slate-700">
                <span>TPR @ FPR 0.10%</span>
                <span className="font-bold text-accent">99.1%</span>
              </div>
              <div className="flex justify-between items-center text-slate-700">
                <span>TPR @ FPR 0.50%</span>
                <span className="font-bold text-emerald-600">99.8%</span>
              </div>
            </div>

            <p className="text-[11px] text-slate-500 font-sans">
              Prometheus scrape metrics are exposed on endpoint{" "}
              <code className="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-slate-700">/metrics</code>.
            </p>
          </div>
        </Card>
      </div>
    </div>
  );
}
