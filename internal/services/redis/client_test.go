package redis

import (
	"fmt"
	"testing"
)

// TestTaskKey verifies the key format used for task status storage.
// This format must stay stable: redis client and handlers depend on it.
func TestTaskKey(t *testing.T) {
	tests := []struct {
		taskID string
		want   string
	}{
		{"parent_training", "task_status:parent_training"},
		{"aapl", "task_status:aapl"},
		{"some-task-123", "task_status:some-task-123"},
	}
	for _, tc := range tests {
		t.Run(tc.taskID, func(t *testing.T) {
			got := TaskKey(tc.taskID)
			if got != tc.want {
				t.Errorf("TaskKey(%q) = %q, want %q", tc.taskID, got, tc.want)
			}
		})
	}
}

// TestCacheKey verifies the key format used for prediction caching.
// This format must stay stable: cache service and system/cache endpoint depend on it.
func TestCacheKey(t *testing.T) {
	tests := []struct {
		ticker string
		want   string
	}{
		{"aapl", "predict_child_aapl"},
		{"tsla", "predict_child_tsla"},
		{"^gspc", "predict_child_^gspc"},
	}
	for _, tc := range tests {
		t.Run(tc.ticker, func(t *testing.T) {
			got := CacheKey(tc.ticker)
			if got != tc.want {
				t.Errorf("CacheKey(%q) = %q, want %q", tc.ticker, got, tc.want)
			}
		})
	}
}

// TestRateLimitKey verifies the key format used for rate limiting.
func TestRateLimitKey(t *testing.T) {
	key := RateLimitKey("train", 1234567890)
	want := fmt.Sprintf("rate_limit:train:%d", 1234567890)
	if key != want {
		t.Errorf("RateLimitKey = %q, want %q", key, want)
	}
}
