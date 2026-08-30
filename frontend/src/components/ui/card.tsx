import { cn } from "@/lib/utils";
import type { HTMLAttributes } from "react";

export function Card({
  className,
  hoverable = false,
  ...props
}: HTMLAttributes<HTMLDivElement> & { hoverable?: boolean }) {
  return (
    <div
      className={cn(
        "rounded-xl border border-line bg-panel p-5 shadow-card transition-all duration-200",
        hoverable && "hover:border-slate-300 hover:shadow-card-hover",
        className,
      )}
      {...props}
    />
  );
}

