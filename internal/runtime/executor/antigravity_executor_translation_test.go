package executor

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestAntigravityTranslationPayloadWithRequestedModel(t *testing.T) {
	tests := []struct {
		name           string
		payload        string
		requestedModel string
		wantModel      string
	}{
		{
			name:           "restores removed model",
			payload:        `{"messages":[{"role":"user","content":"hi"}]}`,
			requestedModel: "claude-opus-4-6",
			wantModel:      "claude-opus-4-6",
		},
		{
			name:           "restores original alias over rewritten target",
			payload:        `{"model":"claude-opus-4-6-thinking","messages":[]}`,
			requestedModel: "claude-opus-4.6(none)",
			wantModel:      "claude-opus-4.6(none)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := antigravityTranslationPayloadWithRequestedModel([]byte(tt.payload), tt.requestedModel)
			if got := gjson.GetBytes(output, "model").String(); got != tt.wantModel {
				t.Fatalf("model = %q, want %q: %s", got, tt.wantModel, output)
			}
		})
	}
}
