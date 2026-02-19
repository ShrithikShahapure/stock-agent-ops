import { useState, type FormEvent } from "react";

interface Props {
  onSubmit: (ticker: string) => void;
  placeholder?: string;
  buttonLabel?: string;
  disabled?: boolean;
  defaultValue?: string;
}

export default function TickerInput({
  onSubmit,
  placeholder = "AAPL, NVDA, BTC-USD...",
  buttonLabel = "Analyze",
  disabled = false,
  defaultValue = "",
}: Props) {
  const [value, setValue] = useState(defaultValue);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const trimmed = value.trim().toUpperCase();
    if (trimmed) onSubmit(trimmed);
  }

  return (
    <form onSubmit={handleSubmit} className="flex gap-2">
      <div className="relative flex-1">
        <input
          type="text"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder={placeholder}
          disabled={disabled}
          className="w-full rounded-xl border border-white/[0.06] bg-white/[0.03] px-4 py-2.5 text-sm text-white placeholder-gray-600 transition-all focus:border-blue-500/40 focus:outline-none focus:ring-1 focus:ring-blue-500/20 focus:bg-white/[0.05]"
        />
      </div>
      <button
        type="submit"
        disabled={disabled || !value.trim()}
        className="btn-glow rounded-xl px-5 py-2.5 text-sm"
      >
        {buttonLabel}
      </button>
    </form>
  );
}
