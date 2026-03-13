package envelope

import (
	"testing"
)

func TestClassifyHTTPError_ParsesFrontAPIError(t *testing.T) {
	body := []byte(`{"_error":{"status":404,"title":"Not found","message":"Conversation not found cnv_bad"}}`)
	msg, code, _ := ClassifyHTTPError(404, body)
	if msg != "Conversation not found cnv_bad" {
		t.Errorf("expected parsed message, got %q", msg)
	}
	if code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %q", code)
	}
}

func TestClassifyHTTPError_FallsBackToHTTPStatus(t *testing.T) {
	body := []byte(`some raw error text`)
	msg, code, _ := ClassifyHTTPError(500, body)
	if msg != "HTTP 500" {
		t.Errorf("expected HTTP 500 fallback, got %q", msg)
	}
	if code != "API_ERROR" {
		t.Errorf("expected API_ERROR, got %q", code)
	}
}

func TestClassifyHTTPError_EmptyBody(t *testing.T) {
	msg, _, _ := ClassifyHTTPError(500, nil)
	if msg != "HTTP 500" {
		t.Errorf("expected fallback message, got %q", msg)
	}
}

func TestClassifyHTTPError_StatusCodes(t *testing.T) {
	tests := []struct {
		status int
		code   string
	}{
		{401, "UNAUTHORIZED"},
		{403, "FORBIDDEN"},
		{404, "NOT_FOUND"},
		{429, "RATE_LIMITED"},
		{500, "API_ERROR"},
		{503, "API_ERROR"},
	}
	for _, tt := range tests {
		_, code, _ := ClassifyHTTPError(tt.status, []byte("x"))
		if code != tt.code {
			t.Errorf("status %d: expected code %q, got %q", tt.status, tt.code, code)
		}
	}
}

func TestClassifyHTTPError_NonErrorJSON(t *testing.T) {
	body := []byte(`{"message":"not the right shape"}`)
	msg, _, _ := ClassifyHTTPError(400, body)
	if msg != "HTTP 400" {
		t.Errorf("expected HTTP 400 fallback for non-_error JSON, got %q", msg)
	}
}

func TestNextPageToken(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://api2.frontapp.com/conversations?page_token=abc123", "abc123"},
		{"https://api2.frontapp.com/conversations?page_token=abc123&limit=25", "abc123"},
		{"", ""},
		{"https://api2.frontapp.com/conversations", ""},
		{"://invalid", ""},
	}
	for _, tt := range tests {
		got := NextPageToken(tt.url)
		if got != tt.want {
			t.Errorf("NextPageToken(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
