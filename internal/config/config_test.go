package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Unset any env overrides that might bleed in from the test environment
	envKeys := []string{
		"PORT", "GRACEFUL_TIMEOUT",
		"REDIS_HOST", "REDIS_PORT", "REDIS_DB",
		"PYTHON_PATH", "SCRIPT_PATH",
		"OUTPUTS_DIR", "LOGS_DIR", "PARENT_DIR", "PARENT_TICKER",
		"PYTHON_TIMEOUT", "TRAINING_TIMEOUT", "MAX_WORKERS",
		"LLM_MODEL",
	}
	for _, k := range envKeys {
		os.Unsetenv(k)
	}

	cfg := Load()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Port", cfg.Port, "8000"},
		{"GracefulTimeout", cfg.GracefulTimeout, 30},
		{"RedisHost", cfg.RedisHost, "localhost"},
		{"RedisPort", cfg.RedisPort, "6379"},
		{"RedisDB", cfg.RedisDB, 0},
		{"PythonPath", cfg.PythonPath, "python"},
		{"ScriptPath", cfg.ScriptPath, "scripts/ml_cli.py"},
		{"OutputsDir", cfg.OutputsDir, "outputs"},
		{"LogsDir", cfg.LogsDir, "logs"},
		{"ParentDir", cfg.ParentDir, "outputs/parent"},
		{"ParentTicker", cfg.ParentTicker, "^GSPC"},
		{"PythonTimeout", cfg.PythonTimeout, 120},
		{"TrainingTimeout", cfg.TrainingTimeout, 7200},
		{"MaxWorkers", cfg.MaxWorkers, 4},
		{"LLMModel", cfg.LLMModel, "qwen3-7b"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("Config.%s = %v, want %v", tc.name, tc.got, tc.want)
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("REDIS_HOST", "redis-server")
	t.Setenv("PYTHON_PATH", "/usr/local/bin/python3")
	t.Setenv("MAX_WORKERS", "8")
	t.Setenv("LLM_MODEL", "gemma-3")

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("Port = %s, want 9090", cfg.Port)
	}
	if cfg.RedisHost != "redis-server" {
		t.Errorf("RedisHost = %s, want redis-server", cfg.RedisHost)
	}
	if cfg.PythonPath != "/usr/local/bin/python3" {
		t.Errorf("PythonPath = %s, want /usr/local/bin/python3", cfg.PythonPath)
	}
	if cfg.MaxWorkers != 8 {
		t.Errorf("MaxWorkers = %d, want 8", cfg.MaxWorkers)
	}
	if cfg.LLMModel != "gemma-3" {
		t.Errorf("LLMModel = %s, want gemma-3", cfg.LLMModel)
	}
}

func TestGetEnvIntInvalidFallsBack(t *testing.T) {
	t.Setenv("MAX_WORKERS", "not-a-number")

	cfg := Load()
	if cfg.MaxWorkers != 4 {
		t.Errorf("MaxWorkers with invalid env = %d, want default 4", cfg.MaxWorkers)
	}
}
