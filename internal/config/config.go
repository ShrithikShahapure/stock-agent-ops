package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	Port            string
	GracefulTimeout int

	// Redis
	RedisHost string
	RedisPort string
	RedisDB   int

	// Python
	PythonPath string
	ScriptPath string

	// Paths
	OutputsDir     string
	LogsDir        string
	ParentDir      string
	ParentTicker   string
	FeatureStoreDir string

	// External Services
	MLflowURI      string
	QdrantHost     string
	QdrantPort     string
	LlamaCppURL    string
	LlamaCppEmbedURL string
	FinnhubKey     string
	LLMModel       string

	// Timeouts (seconds)
	PythonTimeout  int
	TrainingTimeout int

	// Workers
	MaxWorkers int
}

// Load reads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		// Server
		Port:            getEnv("PORT", "8000"),
		GracefulTimeout: getEnvInt("GRACEFUL_TIMEOUT", 30),

		// Redis
		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),
		RedisDB:   getEnvInt("REDIS_DB", 0),

		// Python
		PythonPath: getEnv("PYTHON_PATH", "python"),
		ScriptPath: getEnv("SCRIPT_PATH", "scripts/ml_cli.py"),

		// Paths
		OutputsDir:      getEnv("OUTPUTS_DIR", "outputs"),
		LogsDir:         getEnv("LOGS_DIR", "logs"),
		ParentDir:       getEnv("PARENT_DIR", "outputs/parent"),
		ParentTicker:    getEnv("PARENT_TICKER", "^GSPC"),
		FeatureStoreDir: getEnv("FEATURE_STORE_DIR", "feature_store"),

		// External Services
		MLflowURI:        getEnv("MLFLOW_TRACKING_URI", ""),
		QdrantHost:       getEnv("QDRANT_HOST", "localhost"),
		QdrantPort:       getEnv("QDRANT_PORT", "6333"),
		LlamaCppURL:      getEnv("LLAMA_CPP_BASE_URL", "http://localhost:8080/v1"),
		LlamaCppEmbedURL: getEnv("LLAMA_CPP_EMBED_URL", ""),
		FinnhubKey:       getEnv("FMI_API_KEY", ""),
		LLMModel:         getEnv("LLM_MODEL", "qwen3-7b"),

		// Timeouts
		PythonTimeout:   getEnvInt("PYTHON_TIMEOUT", 120),
		TrainingTimeout: getEnvInt("TRAINING_TIMEOUT", 7200),

		// Workers
		MaxWorkers: getEnvInt("MAX_WORKERS", 4),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
