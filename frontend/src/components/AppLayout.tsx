import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";

const links = [
  { to: "/dashboard", label: "Dashboard" },
  { to: "/security", label: "Security" },
  { to: "/security/scans", label: "Scans" },
  { to: "/inference", label: "Inference" },
  { to: "/workers", label: "Workers" },
  { to: "/serving", label: "Serving" },
  { to: "/models", label: "Models" },
  { to: "/metrics", label: "Metrics" },
  { to: "/settings", label: "Settings" },
];

export function AppLayout() {
  const logout = useAuth((s) => s.logout);
  const navigate = useNavigate();

  return (
    <div className="grid min-h-screen grid-cols-[240px_1fr] bg-canvas">
      <aside className="border-r border-line bg-white px-4 py-6 shadow-sm">
        <div className="mb-8">
          <div className="font-mono text-xs uppercase tracking-[0.2em] text-accent">SPIDER</div>
          <div className="text-sm text-slate-500">Runtime defense</div>
        </div>
        <nav className="space-y-1">
          {links.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              className={({ isActive }) =>
                `block rounded-lg px-3 py-2 text-sm transition ${
                  isActive
                    ? "bg-blue-50 font-medium text-accent"
                    : "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
                }`
              }
            >
              {link.label}
            </NavLink>
          ))}
        </nav>
        <button
          className="mt-10 text-xs text-slate-500 transition hover:text-slate-800"
          onClick={() => {
            logout();
            navigate("/login");
          }}
        >
          Sign out
        </button>
      </aside>
      <main className="min-h-screen overflow-auto bg-canvas p-8">
        <Outlet />
      </main>
    </div>
  );
}
