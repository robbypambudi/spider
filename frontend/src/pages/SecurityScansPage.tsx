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
      <h1 className="text-2xl font-medium">Scan history</h1>
      <Card className="overflow-x-auto p-0">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-line text-xs uppercase text-zinc-500">
            <tr>
              <th className="px-4 py-3">Time</th>
              <th>Decision</th>
              <th>Score</th>
              <th>Detector</th>
              <th>Policy</th>
              <th>Chunks</th>
              <th>Latency</th>
              <th>Model</th>
            </tr>
          </thead>
          <tbody>
            {(data ?? []).map((scan) => (
              <tr key={scan.id} className="border-b border-line/70 hover:bg-zinc-900/40">
                <td className="px-4 py-3 font-mono text-xs">
                  <Link className="text-orange-400" to={`/security/scans/${scan.id}`}>
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
