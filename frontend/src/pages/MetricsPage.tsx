import { useQuery } from "@tanstack/react-query";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";
import { formatMs, formatPct } from "@/lib/utils";

export function MetricsPage() {
  const { data } = useQuery({
    queryKey: ["metrics", "summary"],
    queryFn: api.summary,
    staleTime: 15_000,
  });

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold text-slate-900">Metrics</h1>
      <p className="text-sm text-slate-500">
        JSON summary for the dashboard. Prometheus scrape endpoint is{" "}
        <code className="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-700">/metrics</code>.
      </p>
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <div className="text-sm text-slate-500">Scans</div>
          <div className="mt-2 font-mono text-2xl text-slate-900">{data?.total_scans ?? 0}</div>
        </Card>
        <Card>
          <div className="text-sm text-slate-500">Detection rate</div>
          <div className="mt-2 font-mono text-2xl text-slate-900">{formatPct(data?.detection_rate)}</div>
        </Card>
        <Card>
          <div className="text-sm text-slate-500">Overhead</div>
          <div className="mt-2 font-mono text-2xl text-slate-900">{formatMs(data?.avg_security_overhead_ms)}</div>
        </Card>
      </div>
    </div>
  );
}
