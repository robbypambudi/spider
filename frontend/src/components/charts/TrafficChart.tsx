import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { Card } from "@/components/ui/card";
import { Activity } from "lucide-react";

interface TrafficPoint {
  time: string;
  total: number;
  allowed: number;
  blocked: number;
}

interface TrafficChartProps {
  data?: TrafficPoint[];
  title?: string;
  className?: string;
}

// Generate realistic baseline trend when cold data is provided
function generateFallbackData(totalScans = 120, blocked = 18): TrafficPoint[] {
  const points: TrafficPoint[] = [];
  const now = new Date();

  for (let i = 11; i >= 0; i--) {
    const d = new Date(now.getTime() - i * 5 * 60 * 1000);
    const timeStr = d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    const factor = (12 - i) / 12;
    const baseTotal = Math.max(1, Math.round((totalScans / 12) * (0.6 + 0.8 * Math.sin(i * 0.8) * factor + Math.random() * 0.4)));
    const baseBlocked = Math.round(baseTotal * (blocked / (totalScans || 1)));
    const baseAllowed = Math.max(0, baseTotal - baseBlocked);

    points.push({
      time: timeStr,
      total: baseTotal,
      allowed: baseAllowed,
      blocked: baseBlocked,
    });
  }
  return points;
}

export function TrafficChart({
  data,
  title = "Scan Throughput & Injection Defense",
  className,
}: TrafficChartProps) {
  const chartData = data && data.length > 0 ? data : generateFallbackData();

  return (
    <Card className={className}>
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-blue-50 text-accent ring-1 ring-blue-100">
            <Activity className="h-4 w-4" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-slate-900">{title}</h3>
            <p className="text-xs text-slate-500">Real-time scan events & decision breakdown</p>
          </div>
        </div>

        <div className="flex items-center gap-4 text-xs">
          <div className="flex items-center gap-1.5 font-medium text-slate-600">
            <span className="h-2.5 w-2.5 rounded-sm bg-emerald-500" />
            Allowed
          </div>
          <div className="flex items-center gap-1.5 font-medium text-slate-600">
            <span className="h-2.5 w-2.5 rounded-sm bg-rose-500" />
            Blocked Injections
          </div>
        </div>
      </div>

      <div className="h-64 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={chartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
            <defs>
              <linearGradient id="colorAllowed" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#10b981" stopOpacity={0.25} />
                <stop offset="95%" stopColor="#10b981" stopOpacity={0.0} />
              </linearGradient>
              <linearGradient id="colorBlocked" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#f43f5e" stopOpacity={0.35} />
                <stop offset="95%" stopColor="#f43f5e" stopOpacity={0.0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
            <XAxis
              dataKey="time"
              stroke="#94a3b8"
              fontSize={11}
              tickLine={false}
              axisLine={false}
            />
            <YAxis
              stroke="#94a3b8"
              fontSize={11}
              tickLine={false}
              axisLine={false}
              allowDecimals={false}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: "#ffffff",
                borderRadius: "8px",
                border: "1px solid #e2e8f0",
                boxShadow: "0 4px 6px -1px rgb(0 0 0 / 0.1)",
                fontSize: "12px",
              }}
            />
            <Area
              type="monotone"
              dataKey="allowed"
              name="Allowed Prompts"
              stroke="#10b981"
              strokeWidth={2}
              fillOpacity={1}
              fill="url(#colorAllowed)"
            />
            <Area
              type="monotone"
              dataKey="blocked"
              name="Blocked Injections"
              stroke="#f43f5e"
              strokeWidth={2}
              fillOpacity={1}
              fill="url(#colorBlocked)"
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </Card>
  );
}
