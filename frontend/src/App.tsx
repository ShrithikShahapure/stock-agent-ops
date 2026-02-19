import { Routes, Route, Navigate } from "react-router-dom";
import AppShell from "./components/layout/AppShell";
import AnalysisPage from "./components/analysis/AnalysisPage";
import MonitorPage from "./components/monitor/MonitorPage";

export default function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route path="/analyze" element={<AnalysisPage />} />
        <Route path="/monitor" element={<MonitorPage />} />
        <Route path="*" element={<Navigate to="/analyze" replace />} />
      </Route>
    </Routes>
  );
}
