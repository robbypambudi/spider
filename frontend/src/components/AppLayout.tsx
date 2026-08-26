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
    <div className="grid min-h-screen grid-cols-[220px_1fr]">
      <aside className="border-r border-line bg-zinc-950 px-4 py-6">
        <div className="mb-8">
          <div className="font-mono text-xs uppercase tracking-[0.2em] text-orange-400">SPIDER</div>
          <div className="text-sm text-zinc-400">Runtime defense</div>
        </div>
        <nav className="space-y-1">
          {links.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              className={({ isActive }) =>
                `block rounded px-2 py-1.5 text-sm ${isActive ? "bg-zinc-900 text-orange-400" : "text-zinc-400 hover:text-zinc-100"}`
              }
            >
              {link.label}
            </NavLink>
          ))}
        </nav>
        <button
          className="mt-10 text-xs text-zinc-500 hover:text-zinc-200"
          onClick={() => {
            logout();
            navigate("/login");
          }}
        >
          Sign out
        </button>
      </aside>
      <main className="min-h-screen overflow-auto p-8">
        <Outlet />
      </main>
    </div>
  );
}
