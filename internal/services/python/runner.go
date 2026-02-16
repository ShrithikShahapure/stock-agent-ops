package python

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
)

// Runner executes Python CLI commands
type Runner struct {
	pythonPath string
	scriptPath string
	timeout    time.Duration
	env        []string
}

// NewRunner creates a new Python CLI runner
func NewRunner(cfg *config.Config) *Runner {
	// Build environment variables to pass to Python
	env := os.Environ()

	// Add/override specific env vars
	envVars := map[string]string{
		"MLFLOW_TRACKING_URI":  cfg.MLflowURI,
		"QDRANT_HOST":          cfg.QdrantHost,
		"QDRANT_PORT":          cfg.QdrantPort,
		"LLAMA_CPP_BASE_URL":   cfg.LlamaCppURL,
		"LLAMA_CPP_EMBED_URL":  cfg.LlamaCppEmbedURL,
		"FMI_API_KEY":          cfg.FinnhubKey,
		"LLM_MODEL":            cfg.LLMModel,
	}

	for k, v := range envVars {
		if v != "" {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return &Runner{
		pythonPath: cfg.PythonPath,
		scriptPath: cfg.ScriptPath,
		timeout:    time.Duration(cfg.PythonTimeout) * time.Second,
		env:        env,
	}
}

// Result represents the result of a Python CLI execution
type Result struct {
	Data  map[string]interface{}
	Error string
}

// Execute runs a Python CLI command and returns the result
func (r *Runner) Execute(ctx context.Context, args ...string) (*Result, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Build command
	cmdArgs := append([]string{r.scriptPath}, args...)
	cmd := exec.CommandContext(ctx, r.pythonPath, cmdArgs...)
	cmd.Env = r.env

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err := cmd.Run()

	// Parse output
	var result Result
	if stdout.Len() > 0 {
		if jsonErr := json.Unmarshal(stdout.Bytes(), &result.Data); jsonErr != nil {
			// If output isn't valid JSON, include it as error
			result.Error = fmt.Sprintf("Invalid JSON output: %s", stdout.String())
		}
	}

	// Check for error
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("command timed out after %v", r.timeout)
		}

		// Check if we got JSON error output
		if result.Data != nil {
			if errMsg, ok := result.Data["error"].(string); ok {
				return nil, fmt.Errorf("%s", errMsg)
			}
		}

		// Include stderr in error
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return nil, fmt.Errorf("command failed: %s", errMsg)
	}

	// Check for error in JSON output
	if result.Data != nil {
		if errMsg, ok := result.Data["error"].(string); ok {
			return nil, fmt.Errorf("%s", errMsg)
		}
	}

	return &result, nil
}

// TrainParent runs the train-parent command
func (r *Runner) TrainParent(ctx context.Context) (*Result, error) {
	return r.Execute(ctx, "train-parent")
}

// TrainChild runs the train-child command for a specific ticker
func (r *Runner) TrainChild(ctx context.Context, ticker string) (*Result, error) {
	return r.Execute(ctx, "train-child", "--ticker", ticker)
}

// PredictParent runs the predict-parent command
func (r *Runner) PredictParent(ctx context.Context) (*Result, error) {
	return r.Execute(ctx, "predict-parent")
}

// PredictChild runs the predict-child command for a specific ticker
func (r *Runner) PredictChild(ctx context.Context, ticker string) (*Result, error) {
	return r.Execute(ctx, "predict-child", "--ticker", ticker)
}

// Analyze runs the analyze command for a specific ticker
func (r *Runner) Analyze(ctx context.Context, ticker string, threadID string) (*Result, error) {
	args := []string{"analyze", "--ticker", ticker}
	if threadID != "" {
		args = append(args, "--thread-id", threadID)
	}
	return r.Execute(ctx, args...)
}

// MonitorParent runs the monitor-parent command
func (r *Runner) MonitorParent(ctx context.Context) (*Result, error) {
	return r.Execute(ctx, "monitor-parent")
}

// MonitorTicker runs the monitor-ticker command for a specific ticker
func (r *Runner) MonitorTicker(ctx context.Context, ticker string) (*Result, error) {
	return r.Execute(ctx, "monitor-ticker", "--ticker", ticker)
}
