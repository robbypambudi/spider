import { cn } from "@/lib/utils";

interface GaugeProps {
  value: number; // 0 to 100 or 0.0 to 1.0 depending on isNormalized
  max?: number;
  isNormalized?: boolean; // if true, 0.0 to 1.0
  size?: number;
  strokeWidth?: number;
  label?: string;
  sublabel?: string;
  showPercent?: boolean;
  colorVariant?: "default" | "dynamic" | "threat" | "emerald" | "amber" | "rose";
  className?: string;
}

export function Gauge({
  value,
  max = 100,
  isNormalized = false,
  size = 80,
  strokeWidth = 7,
  label,
  sublabel,
  showPercent = true,
  colorVariant = "dynamic",
  className,
}: GaugeProps) {
  const normalizedVal = isNormalized ? value * 100 : (value / max) * 100;
  const clamped = Math.min(Math.max(0, normalizedVal), 100);

  const radius = (size - strokeWidth) / 2;
  const circumference = radius * 2 * Math.PI;
  const offset = circumference - (clamped / 100) * circumference;

  let strokeColor = "#2563eb"; // default accent
  if (colorVariant === "threat" || colorVariant === "dynamic") {
    if (clamped >= 80) {
      strokeColor = "#f43f5e"; // rose-500
    } else if (clamped >= 50) {
      strokeColor = "#f59e0b"; // amber-500
    } else {
      strokeColor = "#10b981"; // emerald-500
    }
  } else if (colorVariant === "emerald") {
    strokeColor = "#10b981";
  } else if (colorVariant === "amber") {
    strokeColor = "#f59e0b";
  } else if (colorVariant === "rose") {
    strokeColor = "#f43f5e";
  }

  return (
    <div className={cn("inline-flex flex-col items-center justify-center", className)}>
      <div className="relative inline-flex items-center justify-center" style={{ width: size, height: size }}>
        <svg className="-rotate-90 transform" width={size} height={size}>
          {/* Background circle */}
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            stroke="#f1f5f9"
            strokeWidth={strokeWidth}
            fill="transparent"
          />
          {/* Active progress circle */}
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            stroke={strokeColor}
            strokeWidth={strokeWidth}
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            strokeLinecap="round"
            fill="transparent"
            className="transition-all duration-500 ease-out"
          />
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center text-center">
          <span className="font-mono text-xs font-semibold text-slate-800">
            {isNormalized ? value.toFixed(2) : Math.round(value)}
            {showPercent && !isNormalized ? "%" : ""}
          </span>
          {sublabel && <span className="text-[9px] text-slate-400">{sublabel}</span>}
        </div>
      </div>
      {label && <span className="mt-1 text-xs font-medium text-slate-600">{label}</span>}
    </div>
  );
}
