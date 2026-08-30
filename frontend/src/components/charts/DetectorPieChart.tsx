import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";
import { Card } from "@/components/ui/card";
import { ShieldCheck } from "lucide-react";

interface DecisionDistributionProps {
  allowed?: number;
  blocked?: number;
  review?: number;
  className?: string;
}

export function DetectorPieChart({
  allowed = 100,
  blocked = 20,
  review = 0,
  className,
}: DecisionDistributionProps) {
  const total = allowed + blocked + review;

  const data = [
    { name: "ALLOW", value: allowed, color: "#10b981" },
    { name: "BLOCK", value: blocked, color: "#f43f5e" },
    ...(review > 0 ? [{ name: "REVIEW", value: review, color: "#f59e0b" }] : []),
  ];

  return (
    <Card className={className}>
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600 ring-1 ring-emerald-100">
            <ShieldCheck className="h-4 w-4" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-slate-900">Decision Breakdown</h3>
            <p className="text-xs text-slate-500">Security enforcement distribution</p>
          </div>
        </div>
      </div>

      <div className="flex flex-col items-center sm:flex-row">
        <div className="relative h-48 w-48 shrink-0">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={data}
                cx="50%"
                cy="50%"
                innerRadius={52}
                outerRadius={72}
                paddingAngle={4}
                dataKey="value"
              >
                {data.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} stroke="#ffffff" strokeWidth={2} />
                ))}
              </Pie>
              <Tooltip
                formatter={(value: unknown, name: unknown) => [
                  `${value} scans (${total > 0 && typeof value === "number" ? ((value / total) * 100).toFixed(1) : 0}%)`,
                  String(name),
                ]}
                contentStyle={{
                  backgroundColor: "#ffffff",
                  borderRadius: "8px",
                  border: "1px solid #e2e8f0",
                  fontSize: "12px",
                }}
              />
            </PieChart>
          </ResponsiveContainer>
          <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center text-center">
            <span className="font-mono text-xl font-bold text-slate-900">{total}</span>
            <span className="text-[10px] uppercase tracking-wider text-slate-400">Total Scans</span>
          </div>
        </div>

        <div className="mt-4 flex w-full flex-col justify-center space-y-2.5 sm:mt-0 sm:pl-4">
          {data.map((item) => {
            const pct = total > 0 ? ((item.value / total) * 100).toFixed(1) : "0";
            return (
              <div key={item.name} className="flex items-center justify-between text-xs">
                <div className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: item.color }} />
                  <span className="font-medium text-slate-700">{item.name}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="font-mono font-semibold text-slate-900">{item.value}</span>
                  <span className="w-10 text-right font-mono text-slate-400">{pct}%</span>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </Card>
  );
}
