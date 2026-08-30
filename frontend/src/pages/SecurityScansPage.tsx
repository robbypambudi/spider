import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Tabs } from "@/components/ui/tabs";
import { api } from "@/services/api";
import { formatMs } from "@/lib/utils";
import { ArrowUpRight, RefreshCw, Search } from "lucide-react";

export function SecurityScansPage() {
  const [search, setSearch] = useState("");
  const [decisionFilter, setDecisionFilter] = useState("ALL");

  const { data, refetch, isFetching } = useQuery({
    queryKey: ["security", "scans"],
    queryFn: api.scans,
    staleTime: 10_000,
  });

  const scans = data ?? [];

  const filteredScans = scans.filter((s) => {
    const matchesSearch =
      s.id.toLowerCase().includes(search.toLowerCase()) ||
      (s.policy && s.policy.toLowerCase().includes(search.toLowerCase())) ||
      (s.detector && s.detector.toLowerCase().includes(search.toLowerCase()));

    const matchesDecision = decisionFilter === "ALL" || s.decision === decisionFilter;

    return matchesSearch && matchesDecision;
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Security Scan History</h1>
          <p className="text-sm text-slate-500">
            Historical audit log of all inspected prompts, threat scores, and enforcement decisions.
          </p>
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

      {/* Filter and Search Bar */}
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-line bg-white p-3 shadow-card">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-400" />
          <input
            type="text"
            placeholder="Search by scan ID, policy, detector..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-lg border border-line bg-slate-50/50 pl-8 pr-3 py-1.5 text-xs text-slate-900 placeholder:text-slate-400 outline-none focus:border-accent focus:bg-white focus:ring-1 focus:ring-accent"
          />
        </div>

        <Tabs
          tabs={[
            { id: "ALL", label: "All", count: scans.length },
            { id: "ALLOW", label: "Allowed", count: scans.filter((s) => s.decision === "ALLOW").length },
            { id: "BLOCK", label: "Blocked", count: scans.filter((s) => s.decision === "BLOCK").length },
          ]}
          activeTab={decisionFilter}
          onChange={setDecisionFilter}
        />
      </div>

      {/* Table */}
      <Card className="overflow-x-auto p-0">
        <table className="w-full text-left text-xs text-slate-700">
          <thead className="border-b border-line bg-slate-50 text-[11px] uppercase tracking-wider text-slate-500 font-semibold">
            <tr>
              <th className="px-4 py-3">Timestamp</th>
              <th className="px-4 py-3">Decision</th>
              <th className="px-4 py-3">Threat Score</th>
              <th className="px-4 py-3">Detector</th>
              <th className="px-4 py-3">Policy</th>
              <th className="px-4 py-3">Chunks</th>
              <th className="px-4 py-3">Latency</th>
              <th className="px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-line-subtle font-mono">
            {filteredScans.map((scan) => (
              <tr key={scan.id} className="transition hover:bg-slate-50/70">
                <td className="px-4 py-3 text-slate-500 font-sans">
                  {new Date(scan.created_at).toLocaleString()}
                </td>
                <td className="px-4 py-3">
                  <Badge value={scan.decision} showDot pulse={scan.decision === "BLOCK"} />
                </td>
                <td className="px-4 py-3 font-bold">
                  <span className={scan.decision === "BLOCK" ? "text-rose-600" : "text-emerald-600"}>
                    {scan.score.toFixed(3)}
                  </span>
                </td>
                <td className="px-4 py-3 text-slate-600">{scan.detector}</td>
                <td className="px-4 py-3 text-slate-600">{scan.policy || "default"}</td>
                <td className="px-4 py-3 text-slate-600">{scan.chunks_scanned}</td>
                <td className="px-4 py-3 text-slate-600">{formatMs(scan.latency_ms)}</td>
                <td className="px-4 py-3 font-sans">
                  <Link
                    className="inline-flex items-center gap-1 font-semibold text-accent hover:underline"
                    to={`/security/scans/${scan.id}`}
                  >
                    Inspect
                    <ArrowUpRight className="h-3.5 w-3.5" />
                  </Link>
                </td>
              </tr>
            ))}
            {filteredScans.length === 0 && (
              <tr>
                <td colSpan={8} className="py-10 text-center text-slate-400 font-sans">
                  No scan history records found.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
