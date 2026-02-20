import { Outlet } from "react-router-dom";
import NavBar from "./NavBar";

export default function AppShell() {
  return (
    <div className="min-h-screen transition-colors duration-200" style={{ backgroundColor: "var(--bg-primary)" }}>
      {/* Subtle gradient overlay at top */}
      <div className="fixed inset-0 pointer-events-none z-0">
        <div
          className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] rounded-full blur-3xl"
          style={{ background: `radial-gradient(ellipse, var(--gradient-overlay), transparent)` }}
        />
      </div>

      <div className="relative z-10">
        <NavBar />
        <main className="mx-auto max-w-7xl px-6 py-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
