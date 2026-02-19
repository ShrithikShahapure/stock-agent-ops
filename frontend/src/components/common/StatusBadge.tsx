const styles: Record<string, { text: string; dot: string }> = {
  healthy: { text: "text-[#3fb950]", dot: "pulse-dot-green" },
  warning: { text: "text-[#d29922]", dot: "pulse-dot-yellow" },
  degraded: { text: "text-[#d29922]", dot: "pulse-dot-yellow" },
  error: { text: "text-[#f85149]", dot: "pulse-dot-red" },
  critical: { text: "text-[#f85149]", dot: "pulse-dot-red" },
};

interface Props {
  status: string;
}

export default function StatusBadge({ status }: Props) {
  const key = status.toLowerCase();
  const s = styles[key] ?? { text: "text-gray-400", dot: "pulse-dot-blue" };
  return (
    <span className={`inline-flex items-center gap-2 font-semibold ${s.text}`}>
      <span className={`pulse-dot ${s.dot}`} />
      {status}
    </span>
  );
}
