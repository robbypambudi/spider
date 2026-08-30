import { cn } from "@/lib/utils";
import type { ButtonHTMLAttributes, ReactNode } from "react";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "outline" | "ghost" | "danger";
  size?: "sm" | "md" | "lg";
  icon?: ReactNode;
}

export function Button({
  className,
  variant = "primary",
  size = "md",
  icon,
  children,
  ...props
}: ButtonProps) {
  const sizeStyles = {
    sm: "px-2.5 py-1.5 text-xs font-medium gap-1.5",
    md: "px-3.5 py-2 text-sm font-medium gap-2",
    lg: "px-4 py-2.5 text-base font-medium gap-2.5",
  };

  const variantStyles = {
    primary:
      "bg-accent text-white shadow-sm hover:bg-accent-hover active:bg-blue-800 ring-1 ring-accent/20",
    secondary:
      "bg-slate-100 text-slate-800 hover:bg-slate-200 active:bg-slate-300 ring-1 ring-slate-200",
    outline:
      "border border-line bg-white text-slate-700 shadow-sm hover:bg-slate-50 hover:text-slate-900 active:bg-slate-100",
    ghost:
      "bg-transparent text-slate-600 hover:bg-slate-100 hover:text-slate-900 active:bg-slate-200",
    danger:
      "bg-rose-600 text-white shadow-sm hover:bg-rose-700 active:bg-rose-800 ring-1 ring-rose-600/20",
  };

  return (
    <button
      className={cn(
        "inline-flex items-center justify-center rounded-lg transition-all duration-150 focus:outline-none focus:ring-2 focus:ring-accent/20 disabled:cursor-not-allowed disabled:opacity-50",
        sizeStyles[size],
        variantStyles[variant],
        className,
      )}
      {...props}
    >
      {icon && <span className="shrink-0">{icon}</span>}
      {children}
    </button>
  );
}

