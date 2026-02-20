import { useState } from "react";
import { useAnalysis } from "../../hooks/useAnalysis";
import { parsePredictions } from "../../utils/parsePredictions";
import TickerInput from "../common/TickerInput";
import LoadingSpinner from "../common/LoadingSpinner";
import ErrorAlert from "../common/ErrorAlert";
import KPIRow from "./KPIRow";
import ForecastChart from "./ForecastChart";
import StrategicReport from "./StrategicReport";
import TrainingProgress from "./TrainingProgress";

const features = [
  {
    icon: "M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z",
    title: "Multi-Agent Research",
    desc: "Autonomous LangGraph pipeline with analyst, expert, and critic agents",
  },
  {
    icon: "M13 7h8m0 0v8m0-8l-8 8-4-4-6 6",
    title: "LSTM Forecasting",
    desc: "PyTorch neural network with multi-horizon price predictions (1W · 1M · 1Q)",
  },
  {
    icon: "M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z",
    title: "Sentiment Analysis",
    desc: "Real-time news and market sentiment scoring",
  },
];

export default function AnalysisPage() {
  const {
    phase,
    result,
    error,
    progress,
    trainingElapsed,
    sessionId,
    analyze,
    reset,
  } = useAnalysis();
  const [ticker, setTicker] = useState("");
  const [activeTab, setActiveTab] = useState<"report" | "chart">("report");

  function handleSubmit(t: string) {
    setTicker(t);
    analyze(t);
  }

  // Idle / landing
  if (phase === "idle") {
    return (
      <div className="animate-fade-in-up">
        <div className="mb-8">
          <h1 className="text-3xl font-bold tracking-tight text-white">
            AI Financial Analyst
          </h1>
          <p className="mt-1 text-sm text-[#8b949e]">
            Institutional-grade market intelligence powered by LLM agents
          </p>
        </div>

        <div className="mb-8 max-w-lg">
          <TickerInput
            onSubmit={handleSubmit}
            buttonLabel="Generate Analysis"
            defaultValue="NVDA"
          />
          <p className="mt-2 text-xs text-gray-600">
            Session: {sessionId.slice(0, 8)}
          </p>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          {features.map((f, i) => (
            <div
              key={f.title}
              className={`glass-card p-5 animate-fade-in-up-delay-${i + 1}`}
            >
              <div className="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-blue-500/10">
                <svg className="h-4.5 w-4.5 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d={f.icon} />
                </svg>
              </div>
              <h3 className="text-sm font-semibold text-white">{f.title}</h3>
              <p className="mt-1 text-xs leading-relaxed text-[#8b949e]">{f.desc}</p>
            </div>
          ))}
        </div>
      </div>
    );
  }

  // Analyzing spinner
  if (phase === "analyzing" || phase === "retrying") {
    return (
      <div className="animate-fade-in-up">
        <h1 className="mb-6 text-2xl font-bold tracking-tight text-white">
          Analysis: <span className="gradient-text">{ticker}</span>
        </h1>
        <div className="glass-card p-8 text-center">
          <LoadingSpinner
            text={
              phase === "retrying"
                ? `Re-analyzing ${ticker}...`
                : `Analyzing ${ticker}...`
            }
          />
        </div>
      </div>
    );
  }

  // Training poll
  if (phase === "training") {
    return (
      <div className="animate-fade-in-up">
        <h1 className="mb-6 text-2xl font-bold tracking-tight text-white">
          Analysis: <span className="gradient-text">{ticker}</span>
        </h1>
        <TrainingProgress progress={progress} elapsed={trainingElapsed} />
      </div>
    );
  }

  // Error
  if (phase === "error") {
    return (
      <div className="animate-fade-in-up">
        <h1 className="mb-6 text-2xl font-bold tracking-tight text-white">
          Analysis: <span className="gradient-text">{ticker}</span>
        </h1>
        <ErrorAlert
          message={error ?? "Unknown error"}
          onRetry={() => analyze(ticker)}
          onDismiss={reset}
        />
      </div>
    );
  }

  // Done — show results
  if (phase === "done" && result) {
    const report = result.report ?? result.final_report ?? "";
    const rec = result.recommendation ?? "N/A";
    const conf = result.confidence ?? "N/A";
    const { forecast, history } = parsePredictions(result.predictions);

    return (
      <div className="animate-fade-in-up">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-white">
              Analysis: <span className="gradient-text">{ticker}</span>
            </h1>
            <p className="mt-0.5 text-xs text-gray-600">
              Session: {sessionId.slice(0, 8)}
            </p>
          </div>
          <button
            onClick={reset}
            className="btn-secondary px-4 py-2 text-sm"
          >
            New Analysis
          </button>
        </div>

        <div className="mb-6 animate-fade-in-up-delay-1">
          <KPIRow
            recommendation={rec}
            confidence={conf}
            history={history}
            forecast={forecast}
          />
        </div>

        {/* Tabs */}
        <div className="mb-4 flex gap-1 border-b border-white/[0.06]">
          {[
            { key: "report" as const, label: "Strategic Report", icon: "M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" },
            { key: "chart" as const, label: "Technical Forecast", icon: "M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" },
          ].map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`flex items-center gap-1.5 px-4 py-2.5 text-sm font-medium transition-all ${
                activeTab === tab.key
                  ? "border-b-2 border-blue-500 text-white"
                  : "text-gray-500 hover:text-gray-300"
              }`}
            >
              <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d={tab.icon} />
              </svg>
              {tab.label}
            </button>
          ))}
        </div>

        <div className="animate-fade-in-up-delay-2">
          {activeTab === "report" ? (
            <div className="glass-card p-6">
              <StrategicReport report={report} />
            </div>
          ) : (
            <ForecastChart
              forecast={forecast}
              history={history}
              ticker={ticker}
            />
          )}
        </div>
      </div>
    );
  }

  return null;
}
