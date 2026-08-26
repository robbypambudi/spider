import { Navigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import type { ReactNode } from "react";

export function RequireAuth({ children }: { children: ReactNode }) {
  const token = useAuth((s) => s.token);
  if (!token) return <Navigate to="/login" replace />;
  return <>{children}</>;
}
