interface Props {
  label: string;
  value: string;
  borderColor?: string;
  valueColor?: string;
  icon?: React.ReactNode;
}

export default function MetricCard({
  label,
  value,
  borderColor,
  valueColor = "#e6edf3",
  icon,
}: Props) {
  return (
    <div
      className="glass-card p-5 text-center transition-all hover:scale-[1.02]"
      style={borderColor ? { borderColor: `${borderColor}33`, boxShadow: `0 0 20px ${borderColor}15` } : undefined}
    >
      {icon && <div className="mb-2 flex justify-center text-gray-500">{icon}</div>}
      <div className="text-xs font-medium uppercase tracking-wider text-[#8b949e]">
        {label}
      </div>
      <div
        className="mt-1.5 text-2xl font-bold tracking-tight"
        style={{ color: valueColor }}
      >
        {value}
      </div>
    </div>
  );
}
