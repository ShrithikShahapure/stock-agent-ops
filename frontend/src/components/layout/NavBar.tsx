import { NavLink } from "react-router-dom";
import { useEffect, useState } from "react";
import { API_URL } from "../../config";

const links = [
  { to: "/analyze", label: "Analysis", icon: "M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" },
  { to: "/monitor", label: "Monitor", icon: "M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" },
] as const;

export default function NavBar() {
  const [apiUp, setApiUp] = useState<boolean | null>(null);

  useEffect(() => {
    fetch(`${API_URL}/health`)
      .then((r) => setApiUp(r.ok))
      .catch(() => setApiUp(false));
    const interval = setInterval(() => {
      fetch(`${API_URL}/health`)
        .then((r) => setApiUp(r.ok))
        .catch(() => setApiUp(false));
    }, 30_000);
    return () => clearInterval(interval);
  }, []);

  return (
    <nav className="border-b border-white/[0.06] bg-[#0d1117]/80 backdrop-blur-xl sticky top-0 z-50">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-3">
        <div className="flex items-center gap-6">
          {/* Logo */}
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
              <svg className="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
              </svg>
            </div>
            <span className="text-base font-bold tracking-tight text-white">
              Stock Agent Ops
            </span>
          </div>

          {/* Nav Links */}
          <div className="flex gap-1">
            {links.map((l) => (
              <NavLink
                key={l.to}
                to={l.to}
                className={({ isActive }) =>
                  `flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium transition-all ${
                    isActive
                      ? "bg-white/[0.08] text-white shadow-sm"
                      : "text-gray-400 hover:text-gray-200 hover:bg-white/[0.04]"
                  }`
                }
              >
                <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d={l.icon} />
                </svg>
                {l.label}
              </NavLink>
            ))}
          </div>
        </div>

        {/* Right side */}
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5">
            <div className={`pulse-dot ${apiUp === true ? "pulse-dot-green" : apiUp === false ? "pulse-dot-red" : "pulse-dot-yellow"}`} />
            <span className="text-xs text-gray-500">
              {apiUp === true ? "API Connected" : apiUp === false ? "API Offline" : "Checking..."}
            </span>
          </div>
        </div>
      </div>
    </nav>
  );
}
