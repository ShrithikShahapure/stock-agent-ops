import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

interface Props {
  report: string;
}

export default function StrategicReport({ report }: Props) {
  if (!report) {
    return (
      <div className="flex items-center gap-3 rounded-xl border border-yellow-500/20 bg-yellow-500/[0.06] px-5 py-4">
        <svg className="h-5 w-5 shrink-0 text-yellow-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <p className="text-sm text-yellow-300">No report content generated.</p>
      </div>
    );
  }

  return (
    <div className="report-prose text-sm leading-relaxed">
      <Markdown remarkPlugins={[remarkGfm]}>{report}</Markdown>
    </div>
  );
}
