interface Props {
  text?: string;
  size?: "sm" | "md";
}

export default function LoadingSpinner({ text, size = "md" }: Props) {
  const dim = size === "sm" ? "h-4 w-4" : "h-6 w-6";
  return (
    <div className="flex items-center gap-3 py-4">
      <div className="relative">
        <div
          className={`${dim} rounded-full border-2 border-blue-500/20`}
          style={{ animation: "spin-slow 2s linear infinite" }}
        />
        <div
          className={`absolute inset-0 ${dim} rounded-full border-2 border-transparent border-t-blue-400`}
          style={{ animation: "spin-slow 0.8s linear infinite" }}
        />
      </div>
      {text && <span className="text-sm text-[#8b949e]">{text}</span>}
    </div>
  );
}
