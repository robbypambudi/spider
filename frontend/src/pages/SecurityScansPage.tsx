import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";
import { formatMs } from "@/lib/utils";

export function SecurityScansPage() {
  const { data } = useQuery({
    queryKey: ["security", "scans"],
    queryFn: api.scans,
    staleTime: 10_000,
  });

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold text-slate-900">Scan history</h1>
      <Card className="overflow-x-auto p-0">
        <table className="w-full text-left text-sm text-slate-700">
          <thead className="border-b border-line bg-slate-50 text-xs uppercase text-slate-500">
            <tr>
              <th className="px-4 py-3 font-medium">Time</th>
              <th className="font-medium">Decision</th>
              <th className="font-medium">Score</th>
              <th className="font-medium">Detector</th>
              <th className="font-medium">Policy</th>
              <th className="font-medium">Chunks</th>
              <th className="font-medium">Latency</th>
              <th className="font-medium">Model</th>
            </tr>
          </thead>
          <tbody>
            {(data ?? []).map((scan) => (
              <tr key={scan.id} className="border-b border-line transition hover:bg-slate-50">
                <td className="px-4 py-3 font-mono text-xs">
                  <Link className="font-medium text-accent hover:underline" to={`/security/scans/${scan.id}`}>
                    {new Date(scan.created_at).toLocaleString()}
                  </Link>
                </td>
                <td>
                  <Badge value={scan.decision} />
                </td>
                <td className="font-mono">{scan.score.toFixed(3)}</td>
                <td className="font-mono">{scan.detector}</td>
                <td>{scan.policy}</td>
                <td>{scan.chunks_scanned}</td>
                <td>{formatMs(scan.latency_ms)}</td>
                <td className="font-mono text-xs">{scan.model_target ?? "—"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
