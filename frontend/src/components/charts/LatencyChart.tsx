import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { Card } from "@/components/ui/card";
import { Clock } from "lucide-react";

interface LatencyChartProps {
  avgDetectionMs?: number;
  p95DetectionMs?: number;
  avgOverheadMs?: number;
  className?: string;
}

export function LatencyChart({
  avgDetectionMs = 12.4,
  p95DetectionMs = 28.5,
  avgOverheadMs = 4.2,
  className,
}: LatencyChartProps) {
  const data = [
    { name: "Avg Detection", value: Number(avgDetectionMs.toFixed(1)), color: "#3b82f6" },
    { name: "P95 Detection", value: Number(p95DetectionMs.toFixed(1)), color: "#6366f1" },
    { name: "Security Overhead", value: Number(avgOverheadMs.toFixed(1)), color: "#10b981" },
  ];

  return (
    <Card className={className}>
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-indigo-50 text-indigo-600 ring-1 ring-indigo-100">
            <Clock className="h-4 w-4" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-slate-900">Pipeline Latency Breakdown</h3>
            <p className="text-xs text-slate-500">Latency metrics in milliseconds (ms)</p>
          </div>
        </div>
      </div>

      <div className="h-64 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={data} layout="vertical" margin={{ top: 10, right: 30, left: 20, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="#f1f5f9" />
            <XAxis
              type="number"
              stroke="#94a3b8"
              fontSize={11}
              tickLine={false}
              axisLine={false}
              unit=" ms"
            />
            <YAxis
              type="category"
              dataKey="name"
              stroke="#64748b"
              fontSize={11}
              tickLine={false}
              axisLine={false}
            />
            <Tooltip
              formatter={(value: unknown) => [`${value} ms`, "Latency"]}
              contentStyle={{
                backgroundColor: "#ffffff",
                borderRadius: "8px",
                border: "1px solid #e2e8f0",
                fontSize: "12px",
              }}
            />
            <Bar dataKey="value" radius={[0, 6, 6, 0]} barSize={20}>
              {data.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.color} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>
    </Card>
  );
}
