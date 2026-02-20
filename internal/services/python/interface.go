package python

import "context"

// RunnerInterface defines the contract for executing Python CLI commands.
// Using an interface allows handlers to be tested with mock implementations.
type RunnerInterface interface {
	TrainParent(ctx context.Context) (*Result, error)
	TrainChild(ctx context.Context, ticker string) (*Result, error)
	PredictParent(ctx context.Context) (*Result, error)
	PredictChild(ctx context.Context, ticker string) (*Result, error)
	Analyze(ctx context.Context, ticker, threadID string) (*Result, error)
	MonitorParent(ctx context.Context) (*Result, error)
	MonitorTicker(ctx context.Context, ticker string) (*Result, error)
}
