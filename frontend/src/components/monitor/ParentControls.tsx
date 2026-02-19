import { useTraining } from "../../hooks/useTraining";
import ProgressBar from "../common/ProgressBar";

interface Props {
  onHealthAudit: () => void;
  monitorLoading: boolean;
}

export default function ParentControls({ onHealthAudit, monitorLoading }: Props) {
  const { status, message, progress, trainParent } = useTraining();

  return (
    <div className="glass-card p-5">
      <div className="mb-4 flex items-center gap-2">
        <svg className="h-4 w-4 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        <h3 className="text-xs font-semibold uppercase tracking-wider text-[#8b949e]">
          Parent Management
        </h3>
      </div>

      <div className="space-y-2">
        <button
          onClick={trainParent}
          disabled={status === "running"}
          className="btn-glow w-full px-4 py-2.5 text-sm"
        >
          Train Parent (^GSPC)
        </button>
        <button
          onClick={onHealthAudit}
          disabled={monitorLoading}
          className="btn-secondary w-full px-4 py-2.5 text-sm"
        >
          Health Audit
        </button>
      </div>

      {status === "running" && (
        <div className="mt-4">
          <ProgressBar value={progress} label="Training" step={message} />
        </div>
      )}
      {status === "completed" && (
        <div className="mt-3 flex items-center gap-1.5 text-xs text-green-400">
          <span className="pulse-dot pulse-dot-green" />
          {message}
        </div>
      )}
      {status === "failed" && (
        <p className="mt-3 text-xs text-red-400">{message}</p>
      )}
    </div>
  );
}
