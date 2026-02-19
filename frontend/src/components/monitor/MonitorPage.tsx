import { useMonitor } from "../../hooks/useMonitor";
import TickerInput from "../common/TickerInput";
import LoadingSpinner from "../common/LoadingSpinner";
import ErrorAlert from "../common/ErrorAlert";
import ParentControls from "./ParentControls";
import AgentIntegrity from "./AgentIntegrity";
import DriftStatus from "./DriftStatus";
import SystemLogs from "./SystemLogs";
import OutputsBrowser from "./OutputsBrowser";

export default function MonitorPage() {
  const { loading, error, result, ticker, monitorTicker } = useMonitor();

  const isParent = ticker === "^GSPC";

  return (
    <div className="animate-fade-in-up">
      <div className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight text-white">
          Monitoring Dashboard
        </h1>
        <p className="mt-1 text-sm text-[#8b949e]">
          Model health, drift detection, and agent evaluation
        </p>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-[300px_1fr]">
        {/* Sidebar */}
        <div className="space-y-5">
          <ParentControls
            onHealthAudit={() => monitorTicker("^GSPC")}
            monitorLoading={loading}
          />

          <div className="glass-card p-5">
            <div className="mb-3 flex items-center gap-2">
              <svg className="h-4 w-4 text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              <h3 className="text-xs font-semibold uppercase tracking-wider text-[#8b949e]">
                Analyze Ticker
              </h3>
            </div>
            <TickerInput
              onSubmit={monitorTicker}
              placeholder="AAPL, NVDA..."
              buttonLabel="Check"
              disabled={loading}
            />
          </div>

          <SystemLogs />
        </div>

        {/* Main content */}
        <div className="space-y-5">
          {loading && (
            <div className="glass-card p-8 text-center">
              <LoadingSpinner text={`Running diagnostics for ${ticker}...`} />
            </div>
          )}

          {error && <ErrorAlert message={error} />}

          {result && ticker && (
            <div className="animate-fade-in-up">
              <div className="mb-4 flex items-center gap-2">
                <h2 className="text-lg font-semibold text-white">
                  Results for <span className="gradient-text">{ticker}</span>
                </h2>
                {isParent && (
                  <span className="rounded-full bg-purple-500/10 px-2.5 py-0.5 text-xs font-medium text-purple-400 border border-purple-500/20">
                    Parent Model
                  </span>
                )}
              </div>

              {isParent ? (
                <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
                  {result.eval ? (
                    <AgentIntegrity data={result.eval} />
                  ) : (
                    <EmptyCard text="No audit logs found. Run a quality check to initiate." />
                  )}
                  {result.drift ? (
                    <DriftStatus data={result.drift} />
                  ) : (
                    <EmptyCard text="No drift analysis found. Run Parent Health Audit to generate." />
                  )}
                </div>
              ) : (
                <div>
                  {result.eval ? (
                    <AgentIntegrity data={result.eval} />
                  ) : (
                    <EmptyCard text="No audit logs found for this ticker. Run a quality check to initiate." />
                  )}
                </div>
              )}
            </div>
          )}

          {!result && !loading && !error && (
            <div className="space-y-5">
              <div className="glass-card p-8 text-center animate-fade-in-up-delay-1">
                <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-2xl bg-blue-500/10">
                  <svg className="h-6 w-6 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                  </svg>
                </div>
                <h3 className="text-sm font-semibold text-white">Ready to Monitor</h3>
                <p className="mt-1 text-xs text-[#8b949e]">
                  Use the controls on the left to check model health or evaluate a stock agent.
                </p>
              </div>
              <div className="animate-fade-in-up-delay-2">
                <OutputsBrowser />
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function EmptyCard({ text }: { text: string }) {
  return (
    <div className="glass-card flex items-center gap-3 p-5">
      <svg className="h-5 w-5 shrink-0 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <p className="text-sm text-[#8b949e]">{text}</p>
    </div>
  );
}
