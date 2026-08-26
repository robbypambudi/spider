import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { AppLayout } from "@/components/AppLayout";
import { RequireAuth } from "@/components/RequireAuth";
import { DashboardPage } from "@/pages/DashboardPage";
import { InferencePage } from "@/pages/InferencePage";
import { LoginPage } from "@/pages/LoginPage";
import { MetricsPage } from "@/pages/MetricsPage";
import { ModelsPage } from "@/pages/ModelsPage";
import { SecurityPage } from "@/pages/SecurityPage";
import { SecurityScanDetailPage } from "@/pages/SecurityScanDetailPage";
import { SecurityScansPage } from "@/pages/SecurityScansPage";
import { ServingPage } from "@/pages/ServingPage";
import { SettingsPage } from "@/pages/SettingsPage";
import { WorkerDetailPage } from "@/pages/WorkerDetailPage";
import { WorkersPage } from "@/pages/WorkersPage";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 15_000, refetchOnWindowFocus: false },
  },
});

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            element={
              <RequireAuth>
                <AppLayout />
              </RequireAuth>
            }
          >
            <Route path="/dashboard" element={<DashboardPage />} />
            <Route path="/security" element={<SecurityPage />} />
            <Route path="/security/scans" element={<SecurityScansPage />} />
            <Route path="/security/scans/:scanId" element={<SecurityScanDetailPage />} />
            <Route path="/inference" element={<InferencePage />} />
            <Route path="/workers" element={<WorkersPage />} />
            <Route path="/workers/:workerId" element={<WorkerDetailPage />} />
            <Route path="/serving" element={<ServingPage />} />
            <Route path="/models" element={<ModelsPage />} />
            <Route path="/metrics" element={<MetricsPage />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Route>
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
