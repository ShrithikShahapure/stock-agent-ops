package python

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fakeRunner builds a Runner that executes a temporary helper script.
func fakeRunner(t *testing.T, script string) *Runner {
	t.Helper()

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "fake_cli.py")
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("fakeRunner: write script: %v", err)
	}

	return &Runner{
		pythonPath: "python3",
		scriptPath: scriptPath,
		timeout:    5 * time.Second,
		env:        os.Environ(),
	}
}

func TestExecute_ValidJSON(t *testing.T) {
	r := fakeRunner(t, `import json, sys; print(json.dumps({"status": "ok"}))`)

	result, err := r.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute(valid JSON) err = %v, want nil", err)
	}
	if result.Data["status"] != "ok" {
		t.Errorf("Execute(valid JSON) data = %v, want status=ok", result.Data)
	}
}

func TestExecute_NonZeroExit(t *testing.T) {
	r := fakeRunner(t, `import sys; sys.stderr.write("something went wrong\n"); sys.exit(1)`)

	_, err := r.Execute(context.Background())
	if err == nil {
		t.Fatal("Execute(non-zero exit) err = nil, want error")
	}
}

func TestExecute_JSONErrorField(t *testing.T) {
	r := fakeRunner(t, `import json, sys; print(json.dumps({"error": "model missing"})); sys.exit(1)`)

	_, err := r.Execute(context.Background())
	if err == nil {
		t.Fatal("Execute(JSON error field) err = nil, want error")
	}
	if err.Error() != "model missing" {
		t.Errorf("Execute(JSON error field) err = %q, want \"model missing\"", err.Error())
	}
}

func TestExecute_Timeout(t *testing.T) {
	r := fakeRunner(t, `import time; time.sleep(30)`)
	r.timeout = 100 * time.Millisecond // very short timeout

	ctx := context.Background()
	_, err := r.Execute(ctx)
	if err == nil {
		t.Fatal("Execute(timeout) err = nil, want timeout error")
	}
}

func TestExecute_InvalidJSONOutput(t *testing.T) {
	r := fakeRunner(t, `print("not json")`)

	result, err := r.Execute(context.Background())
	// The runner should not error on invalid JSON stdout alone (no non-zero exit),
	// but it should populate result.Error
	if err != nil {
		t.Fatalf("Execute(invalid JSON stdout, zero exit) err = %v", err)
	}
	if result.Error == "" {
		t.Error("Execute(invalid JSON stdout) result.Error should be non-empty")
	}
}

func TestExecute_PassesArgs(t *testing.T) {
	// Script that prints the first CLI arg as JSON
	r := fakeRunner(t, `import json, sys; print(json.dumps({"arg": sys.argv[1] if len(sys.argv) > 1 else ""}))`)

	result, err := r.Execute(context.Background(), "predict-child", "--ticker", "AAPL")
	if err != nil {
		t.Fatalf("Execute(args) err = %v", err)
	}
	if result.Data["arg"] != "predict-child" {
		t.Errorf("Execute(args) first arg = %v, want \"predict-child\"", result.Data["arg"])
	}
}
