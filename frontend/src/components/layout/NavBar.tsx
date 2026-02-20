import { NavLink } from "react-router-dom";
import { useEffect, useState } from "react";
import { API_URL } from "../../config";
import { useTheme } from "../../context/ThemeContext";

const links = [
  { to: "/analyze", label: "Analysis", icon: "M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" },
  { to: "/monitor", label: "Monitor", icon: "M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" },
] as const;

function SunIcon() {
  return (
    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364-6.364l-.707.707M6.343 17.657l-.707.707M17.657 17.657l-.707-.707M6.343 6.343l-.707-.707M12 8a4 4 0 100 8 4 4 0 000-8z" />
    </svg>
  );
}

function MoonIcon() {
  return (
    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
    </svg>
  );
}

export default function NavBar() {
  const [apiUp, setApiUp] = useState<boolean | null>(null);
  const { theme, toggleTheme } = useTheme();

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
    <nav
      className="sticky top-0 z-50 backdrop-blur-xl transition-colors duration-200"
      style={{
        backgroundColor: "var(--bg-nav)",
        borderBottom: "1px solid var(--border-subtle)",
      }}
    >
      <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-3">
        <div className="flex items-center gap-6">
          {/* Logo */}
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
              <svg className="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
              </svg>
            </div>
            <span className="text-base font-bold tracking-tight" style={{ color: "var(--text-primary)" }}>
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
          {/* API status */}
          <div className="flex items-center gap-1.5">
            <div className={`pulse-dot ${apiUp === true ? "pulse-dot-green" : apiUp === false ? "pulse-dot-red" : "pulse-dot-yellow"}`} />
            <span className="text-xs" style={{ color: "var(--text-dim)" }}>
              {apiUp === true ? "API Connected" : apiUp === false ? "API Offline" : "Checking..."}
            </span>
          </div>

          {/* Theme toggle */}
          <button
            onClick={toggleTheme}
            aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
            className="flex h-8 w-8 items-center justify-center rounded-lg transition-all duration-200"
            style={{
              background: "var(--btn-secondary-bg)",
              border: "1px solid var(--border-card)",
              color: "var(--text-muted)",
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLButtonElement).style.background = "var(--btn-secondary-hover)";
              (e.currentTarget as HTMLButtonElement).style.color = "var(--text-primary)";
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLButtonElement).style.background = "var(--btn-secondary-bg)";
              (e.currentTarget as HTMLButtonElement).style.color = "var(--text-muted)";
            }}
          >
            {theme === "dark" ? <SunIcon /> : <MoonIcon />}
          </button>
        </div>
      </div>
    </nav>
  );
}
