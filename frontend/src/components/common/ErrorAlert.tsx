interface Props {
  message: string;
  onRetry?: () => void;
  onDismiss?: () => void;
}

export default function ErrorAlert({ message, onRetry, onDismiss }: Props) {
  return (
    <div className="animate-fade-in-up rounded-xl border border-red-500/20 bg-red-500/[0.06] px-5 py-4">
      <div className="flex items-start gap-3">
        <div className="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-red-500/20">
          <svg className="h-3 w-3 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
        <div className="flex-1">
          <p className="text-sm text-red-300">{message}</p>
          {(onRetry || onDismiss) && (
            <div className="mt-2 flex gap-2">
              {onRetry && (
                <button onClick={onRetry} className="text-xs font-medium text-red-400 hover:text-red-300 transition-colors">
                  Try Again
                </button>
              )}
              {onDismiss && (
                <button onClick={onDismiss} className="text-xs text-gray-500 hover:text-gray-400 transition-colors">
                  Dismiss
                </button>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
