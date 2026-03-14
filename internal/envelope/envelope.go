package envelope

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
)

type ParamSpec struct {
	Description string   `json:"description"`
	Value       string   `json:"value,omitempty"`
	Default     string   `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Required    bool     `json:"required,omitempty"`
}

type Action struct {
	Command     string               `json:"command"`
	Description string               `json:"description"`
	Params      map[string]ParamSpec `json:"params,omitempty"`
}

type Result struct {
	OK          bool        `json:"ok"`
	Command     string      `json:"command"`
	Result      any         `json:"result"`
	NextActions []Action    `json:"next_actions,omitempty"`
}

type ErrorDetail struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type ErrorResponse struct {
	OK          bool        `json:"ok"`
	Command     string      `json:"command"`
	Error       ErrorDetail `json:"error"`
	Fix         string      `json:"fix,omitempty"`
	NextActions []Action    `json:"next_actions,omitempty"`
}

func PrintResult(command string, result interface{}, actions []Action) {
	FprintResult(os.Stdout, command, result, actions)
}

func FprintResult(w io.Writer, command string, result interface{}, actions []Action) {
	out := Result{
		OK:          true,
		Command:     command,
		Result:      result,
		NextActions: actions,
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func PrintError(command, message, code, fix string, actions []Action) {
	FprintError(os.Stdout, command, message, code, fix, actions)
}

func FprintError(w io.Writer, command, message, code, fix string, actions []Action) {
	out := ErrorResponse{
		OK:      false,
		Command: command,
		Error:   ErrorDetail{Message: message, Code: code},
		Fix:     fix,
		NextActions: actions,
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func ClassifyHTTPError(statusCode int, body []byte) (message, code, fix string) {
	msg := parseAPIError(body)
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", statusCode)
	}

	switch statusCode {
	case 401:
		return msg, "UNAUTHORIZED", "Set FRONT_API_TOKEN or run: front config set token_command '<command>'"
	case 404:
		return msg, "NOT_FOUND", "Check the resource ID and try again"
	case 403:
		return msg, "FORBIDDEN", "Check that your API token has the required scopes"
	case 429:
		return msg, "RATE_LIMITED", "Wait and retry"
	default:
		return msg, "API_ERROR", fmt.Sprintf("API returned status %d", statusCode)
	}
}

// parseAPIError extracts the message from Front API error responses.
// Front returns errors as {"_error":{"status":404,"title":"Not found","message":"..."}}.
func parseAPIError(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var parsed struct {
		Err *struct {
			Message string `json:"message"`
		} `json:"_error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Err != nil && parsed.Err.Message != "" {
		return parsed.Err.Message
	}
	return ""
}

// NextPageToken extracts the page_token query parameter from a pagination URL.
func NextPageToken(nextURL string) string {
	if nextURL == "" {
		return ""
	}
	u, err := url.Parse(nextURL)
	if err != nil {
		return ""
	}
	return u.Query().Get("page_token")
}
