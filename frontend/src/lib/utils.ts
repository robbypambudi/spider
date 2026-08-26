import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatMs(value: number | null | undefined) {
  if (value == null || Number.isNaN(value)) return "—";
  return `${value.toFixed(2)} ms`;
}

export function formatPct(value: number | null | undefined) {
  if (value == null || Number.isNaN(value)) return "—";
  return `${(value * 100).toFixed(1)}%`;
}
