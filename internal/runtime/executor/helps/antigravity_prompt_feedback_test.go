package helps

import (
	"strings"
	"testing"

	"github.com/tidwall/gjson"
)

func TestAntigravityPromptFeedbackErrorBody(t *testing.T) {
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
			body, ok := AntigravityPromptFeedbackErrorBody([]byte(tt.payload))
			if ok != tt.wantErr {
				t.Fatalf("ok = %v, want %v", ok, tt.wantErr)
			}
			if !tt.wantErr {
				return
			}
			if got := gjson.GetBytes(body, "error.code").Int(); got != 400 {
				t.Fatalf("error code = %d, want 400: %s", got, body)
			}
			if !strings.Contains(string(body), "PROHIBITED_CONTENT") || !strings.Contains(string(body), "blocked by policy") {
				t.Fatalf("error body does not preserve actionable feedback: %s", body)
			}
		})
	}
}
