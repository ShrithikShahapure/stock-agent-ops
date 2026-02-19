import ProgressBar from "../common/ProgressBar";

interface Props {
  progress: number;
  elapsed: number;
}

const steps = [
  { threshold: 0, label: "Initializing model architecture..." },
  { threshold: 15, label: "Loading training data..." },
  { threshold: 30, label: "Training LSTM layers..." },
  { threshold: 60, label: "Optimizing weights..." },
  { threshold: 85, label: "Finalizing model..." },
];

function getStep(progress: number) {
  for (let i = steps.length - 1; i >= 0; i--) {
    if (progress >= steps[i].threshold) return steps[i].label;
  }
  return steps[0].label;
}

export default function TrainingProgress({ progress, elapsed }: Props) {
  return (
    <div className="glass-card p-8 animate-fade-in-up">
      <div className="mb-6 flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-yellow-500/10">
          <svg className="h-5 w-5 text-yellow-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5} style={{ animation: "spin-slow 3s linear infinite" }}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </div>
        <div>
          <h3 className="font-semibold text-white">Constructing Neural Model</h3>
          <p className="text-xs text-[#8b949e]">First-time training — this takes 30–60 seconds</p>
        </div>
      </div>

      <ProgressBar value={progress} label="Training Progress" step={getStep(progress)} />

      <div className="mt-4 flex items-center justify-between text-xs text-[#8b949e]">
        <span>{elapsed}s elapsed</span>
        <div className="flex gap-3">
          {steps.slice(0, -1).map((s, i) => (
            <span key={i} className={`flex items-center gap-1 ${progress >= steps[i + 1]?.threshold ? "text-green-400" : ""}`}>
              {progress >= steps[i + 1]?.threshold ? "\u2713" : "\u2022"}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}
