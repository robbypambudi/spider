import { cn } from "@/lib/utils";

interface ProgressProps {
  value: number; // 0 to 100
  max?: number;
  className?: string;
  barClassName?: string;
  colorVariant?: "default" | "emerald" | "amber" | "rose" | "dynamic";
}

export function Progress({
  value,
  max = 100,
  className,
  barClassName,
  colorVariant = "dynamic",
}: ProgressProps) {
  const percentage = Math.min(Math.max(0, (value / max) * 100), 100);

  let variantStyle = "bg-accent";
  if (colorVariant === "emerald") {
    variantStyle = "bg-emerald-500";
  } else if (colorVariant === "amber") {
    variantStyle = "bg-amber-500";
  } else if (colorVariant === "rose") {
    variantStyle = "bg-rose-500";
  } else if (colorVariant === "dynamic") {
    if (percentage >= 85) {
      variantStyle = "bg-rose-500";
    } else if (percentage >= 65) {
      variantStyle = "bg-amber-500";
    } else {
      variantStyle = "bg-emerald-500";
    }
  }

  return (
    <div
      className={cn(
        "relative h-2 w-full overflow-hidden rounded-full bg-slate-100 ring-1 ring-inset ring-slate-200/50",
        className,
      )}
    >
      <div
        className={cn("h-full transition-all duration-300 ease-out", variantStyle, barClassName)}
        style={{ width: `${percentage}%` }}
      />
    </div>
  );
}
