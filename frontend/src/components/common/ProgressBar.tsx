interface Props {
  value: number;
  label?: string;
  step?: string;
}

export default function ProgressBar({ value, label, step }: Props) {
  return (
    <div>
      <div className="mb-1.5 flex justify-between text-xs">
        <span className="text-[#8b949e]">{label ?? "Progress"}</span>
        <span className="font-medium text-gray-300">{Math.round(value)}%</span>
      </div>
      <div className="h-2 w-full overflow-hidden rounded-full bg-white/[0.04]">
        <div
          className="progress-shimmer h-full rounded-full transition-all duration-500 ease-out"
          style={{ width: `${Math.min(value, 100)}%` }}
        />
      </div>
      {step && (
        <p className="mt-1 text-xs text-[#8b949e]">{step}</p>
      )}
    </div>
  );
}
