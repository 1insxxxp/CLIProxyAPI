package executor

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

const antigravityPromptFeedbackMessageLimit = 2048

// antigravityPromptFeedbackError turns Gemini's HTTP 200 policy rejection into
// a non-retryable client error. The upstream returns promptFeedback without a
// candidate for these requests, which otherwise looks like a successful empty
// response to downstream clients.
func antigravityPromptFeedbackError(payload []byte) (statusErr, bool) {
	if len(payload) == 0 || !gjson.ValidBytes(payload) {
		return statusErr{}, false
	}

	response := gjson.ParseBytes(payload)
	if wrapped := response.Get("response"); wrapped.Exists() {
		response = wrapped
	}
	if candidates := response.Get("candidates"); candidates.IsArray() && len(candidates.Array()) > 0 {
		return statusErr{}, false
	}

	feedback := response.Get("promptFeedback")
	reason := strings.TrimSpace(feedback.Get("blockReason").String())
	if reason == "" {
		return statusErr{}, false
	}
	message := strings.TrimSpace(feedback.Get("blockReasonMessage").String())
	if message == "" {
		message = "The upstream provider rejected the prompt. Review the text and attached media, then try again."
	}
	if len(message) > antigravityPromptFeedbackMessageLimit {
		message = message[:antigravityPromptFeedbackMessageLimit]
	}

	body, marshalErr := json.Marshal(map[string]any{
		"error": map[string]any{
			"code":    http.StatusBadRequest,
			"message": "Upstream prompt blocked (" + reason + "): " + message,
			"status":  "INVALID_ARGUMENT",
			"details": []map[string]string{{"reason": reason}},
		},
	})
	if marshalErr != nil {
		return statusErr{code: http.StatusBadRequest, msg: "upstream prompt blocked: " + reason}, true
	}
	return statusErr{code: http.StatusBadRequest, msg: string(body)}, true
}
