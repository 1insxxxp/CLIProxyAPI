package helps

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

const antigravityPromptFeedbackMessageLimit = 2048

// AntigravityPromptFeedbackErrorBody converts an explicit policy rejection
// returned with HTTP 200 into the standard upstream error envelope.
func AntigravityPromptFeedbackErrorBody(payload []byte) ([]byte, bool) {
	if len(payload) == 0 || !gjson.ValidBytes(payload) {
		return nil, false
	}

	response := gjson.ParseBytes(payload)
	if wrapped := response.Get("response"); wrapped.Exists() {
		response = wrapped
	}
	if candidates := response.Get("candidates"); candidates.IsArray() && len(candidates.Array()) > 0 {
		return nil, false
	}

	feedback := response.Get("promptFeedback")
	reason := strings.TrimSpace(feedback.Get("blockReason").String())
	if reason == "" {
		return nil, false
	}
	message := strings.TrimSpace(feedback.Get("blockReasonMessage").String())
	if message == "" {
		message = "The upstream provider rejected the prompt. Review the text and attached media, then try again."
	}
	if len(message) > antigravityPromptFeedbackMessageLimit {
		message = message[:antigravityPromptFeedbackMessageLimit]
	}

	body, errMarshal := json.Marshal(map[string]any{
		"error": map[string]any{
			"code":    http.StatusBadRequest,
			"message": "Upstream prompt blocked (" + reason + "): " + message,
			"status":  "INVALID_ARGUMENT",
			"details": []map[string]string{{"reason": reason}},
		},
	})
	if errMarshal != nil {
		return []byte(`{"error":{"code":400,"message":"Upstream prompt blocked","status":"INVALID_ARGUMENT"}}`), true
	}
	return body, true
}
