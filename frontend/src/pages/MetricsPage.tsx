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
      <h1 className="text-2xl font-medium">Metrics</h1>
      <p className="text-sm text-zinc-400">
        JSON summary for the dashboard. Prometheus scrape endpoint is <code>/metrics</code>.
      </p>
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          Scans <div className="mt-2 font-mono text-2xl">{data?.total_scans ?? 0}</div>
        </Card>
        <Card>
          Detection rate <div className="mt-2 font-mono text-2xl">{formatPct(data?.detection_rate)}</div>
        </Card>
        <Card>
          Overhead <div className="mt-2 font-mono text-2xl">{formatMs(data?.avg_security_overhead_ms)}</div>
        </Card>
      </div>
    </div>
  );
}
