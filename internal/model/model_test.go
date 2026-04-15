package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProcessDelayUnmarshalJSON(t *testing.T) {
	var delay ProcessDelay
	if err := json.Unmarshal([]byte(`"150ms"`), &delay); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if got := delay.Delay(); got != 150*time.Millisecond {
		t.Fatalf("expected 150ms, got %v", got)
	}
}

func TestProcessDelayUnmarshalJSONInvalid(t *testing.T) {
	var delay ProcessDelay
	if err := json.Unmarshal([]byte(`"not-a-duration"`), &delay); err == nil {
		t.Fatalf("expected unmarshal to fail for invalid duration")
	}
}
