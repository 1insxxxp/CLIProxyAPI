package executor

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAntigravityPromptFeedbackError(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr bool
	}{
		{
			name:    "wrapped prohibited content",
			payload: `{"response":{"promptFeedback":{"blockReason":"PROHIBITED_CONTENT","blockReasonMessage":"blocked by policy"},"usageMetadata":{"promptTokenCount":42}}}`,
			wantErr: true,
		},
		{
			name:    "plain prohibited content",
			payload: `{"promptFeedback":{"blockReason":"PROHIBITED_CONTENT","blockReasonMessage":"blocked by policy"}}`,
			wantErr: true,
		},
		{
			name:    "candidate response wins over feedback",
			payload: `{"response":{"promptFeedback":{"blockReason":"SAFETY"},"candidates":[{"content":{"parts":[{"text":"ok"}]}}]}}`,
			wantErr: false,
		},
		{
			name:    "empty response without explicit feedback",
			payload: `{"response":{"usageMetadata":{"promptTokenCount":42}}}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, ok := antigravityPromptFeedbackError([]byte(tt.payload))
			if ok != tt.wantErr {
				t.Fatalf("ok = %v, want %v", ok, tt.wantErr)
			}
			if !tt.wantErr {
				return
			}
			if got := err.StatusCode(); got != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", got, http.StatusBadRequest)
			}
			var body map[string]any
			if unmarshalErr := json.Unmarshal([]byte(err.Error()), &body); unmarshalErr != nil {
				t.Fatalf("error body is not JSON: %v", unmarshalErr)
			}
			if !strings.Contains(err.Error(), "PROHIBITED_CONTENT") || !strings.Contains(err.Error(), "blocked by policy") {
				t.Fatalf("error body does not preserve actionable feedback: %s", err.Error())
			}
		})
	}
}
