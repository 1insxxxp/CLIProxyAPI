package antigravity

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/thinking"
	"github.com/tidwall/gjson"
)

func TestApplyPreservesZeroBudgetForClaudeModelThatAllowsZero(t *testing.T) {
	applier := NewApplier()
	model := &registry.ModelInfo{
		ID: "claude-opus-4-6-thinking",
		Thinking: &registry.ThinkingSupport{
			Min:         1024,
			Max:         64000,
			ZeroAllowed: true,
		},
	}
	body := []byte(`{"request":{"generationConfig":{"thinkingConfig":{"thinkingBudget":0,"includeThoughts":false},"maxOutputTokens":4096}}}`)

	output, err := applier.Apply(body, thinking.ThinkingConfig{
		Mode:   thinking.ModeNone,
		Budget: 0,
	}, model)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	config := gjson.GetBytes(output, "request.generationConfig.thinkingConfig")
	if !config.Exists() {
		t.Fatalf("thinkingConfig was removed: %s", output)
	}
	if got := config.Get("thinkingBudget").Int(); got != 0 {
		t.Fatalf("thinkingBudget = %d, want 0: %s", got, output)
	}
	if got := config.Get("includeThoughts"); !got.Exists() || got.Bool() {
		t.Fatalf("includeThoughts = %s, want false: %s", got.Raw, output)
	}
}
