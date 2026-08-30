import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { SpiderLogo } from "@/components/brand/SpiderLogo";
import { useAuth } from "@/hooks/useAuth";
import {
  BarChart3,
  Bot,
  Boxes,
  Cpu,
  GraduationCap,
  LayoutDashboard,
  LogOut,
  ScanSearch,
  Server,
  Settings,
  Shield,
} from "lucide-react";

const links = [
  { to: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { to: "/security", label: "Security Pipeline", icon: Shield, end: true },
  { to: "/security/scans", label: "Scan History", icon: ScanSearch },
  { to: "/inference", label: "Protected Inference", icon: Bot },
  { to: "/workers", label: "Cluster Workers", icon: Server },
  { to: "/serving", label: "Serving Nodes", icon: Cpu },
  { to: "/models", label: "ML Models", icon: Boxes },
  { to: "/metrics", label: "Metrics & Telemetry", icon: BarChart3 },
  { to: "/settings", label: "Settings", icon: Settings },
];

export function AppLayout() {
  const logout = useAuth((s) => s.logout);
  const email = useAuth((s) => s.email);
  const role = useAuth((s) => s.role);
  const navigate = useNavigate();

  return (
    <div className="grid min-h-screen grid-cols-[260px_1fr] bg-canvas text-slate-900 font-sans">
      {/* Sidebar */}
      <aside className="sticky top-0 flex h-screen flex-col justify-between border-r border-line bg-white px-4 py-5 shadow-sm">
        <div className="space-y-6">
          {/* Logo & Brand Header */}
          <div className="px-2 pt-1">
            <SpiderLogo size={36} variant="full" glow />
          </div>

          {/* Navigation Links */}
          <nav className="space-y-1">
            <div className="px-2 pb-1.5 text-[10px] font-semibold uppercase tracking-wider text-slate-400">
              Control Plane Navigation
            </div>
            {links.map((link) => {
              const Icon = link.icon;
              return (
                <NavLink
                  key={link.to}
                  to={link.to}
                  end={link.end}
                  className={({ isActive }) =>
                    `flex items-center gap-2.5 rounded-lg px-3 py-2 text-xs font-medium transition-all duration-150 ${
                      isActive
                        ? "bg-accent-light text-accent font-semibold shadow-xs ring-1 ring-accent/20"
                        : "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
                    }`
                  }
                >
                  <Icon className="h-4 w-4 shrink-0" />
                  <span>{link.label}</span>
                </NavLink>
              );
            })}
          </nav>
        </div>

        {/* Sidebar Footer with ITS Surabaya Credit & User Profile */}
        <div className="space-y-3 border-t border-line-subtle pt-4">
          {/* Academic Credit */}
          <div className="rounded-lg bg-slate-50 p-2.5 space-y-1 text-slate-600 ring-1 ring-slate-200/60">
            <div className="flex items-center gap-1.5 text-[10px] font-bold uppercase tracking-wider text-slate-500">
              <GraduationCap className="h-3.5 w-3.5 text-accent" />
              <span>Academic R&D</span>
            </div>
            <div className="text-[11px] font-semibold text-slate-800 leading-tight">
              Institut Teknologi Sepuluh Nopember (ITS) Surabaya
            </div>
          </div>

          {/* User Profile & Sign Out */}
          <div className="flex items-center justify-between px-1 text-xs">
            <div className="min-w-0">
              <div className="truncate font-semibold text-slate-800">{email || "admin@spider.local"}</div>
              <div className="font-mono text-[10px] text-slate-400 uppercase">{role || "ADMIN"}</div>
            </div>
            <button
              title="Sign out"
              onClick={() => {
                logout();
                navigate("/login");
              }}
              className="flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 transition hover:bg-rose-50 hover:text-rose-600"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="flex flex-col min-w-0">
        {/* Top Header Bar */}
        <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-line bg-white/85 px-8 backdrop-blur-md">
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 rounded-full bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-700 ring-1 ring-emerald-200">
              <span className="relative flex h-2 w-2">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75" />
                <span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-500" />
              </span>
              Cluster: Active Defense Online
            </div>
          </div>

          <div className="flex items-center gap-3 text-xs text-slate-500 font-mono">
            <span className="hidden sm:inline-flex items-center gap-1.5 rounded-md bg-blue-50 px-2.5 py-0.5 text-[11px] font-sans font-medium text-accent ring-1 ring-blue-200/60">
              <GraduationCap className="h-3.5 w-3.5" />
              ITS Surabaya Collaboration
            </span>
            <span className="rounded-md bg-slate-100 px-2 py-0.5 text-[11px] text-slate-600 ring-1 ring-slate-200">
              Fail-Mode: Closed
            </span>
          </div>
        </header>

        {/* Page Outlet */}
        <main className="flex-1 p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
