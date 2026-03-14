package cmd

// Result types for JSON output. Using typed structs instead of
// map[string]interface{} gives compile-time safety on field names
// and omitempty behavior.

type InboxResult struct {
	Total         int                   `json:"total"`
	Showing       int                   `json:"showing"`
	Query         string                `json:"query"`
	Conversations []ConversationSummary `json:"conversations"`
	NextPageToken string                `json:"next_page_token,omitempty"`
}

type ConversationSummary struct {
	ID      string         `json:"id"`
	Subject string         `json:"subject"`
	Status  string         `json:"status"`
	From    ContactSummary `json:"from"`
	Date    string         `json:"date,omitempty"`
	Tags    []string       `json:"tags,omitempty"`
}

type ContactSummary struct {
	Handle string `json:"handle,omitempty"`
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
}

type ReadResult struct {
	Conversation ConversationSummary `json:"conversation"`
	Messages     []MessageSummary    `json:"messages"`
	Truncated    bool                `json:"truncated,omitempty"`
}

type MessageSummary struct {
	ID        string         `json:"id,omitempty"`
	From      *ContactSummary `json:"from,omitempty"`
	Date      string         `json:"date,omitempty"`
	IsInbound *bool          `json:"is_inbound,omitempty"`
	Text      string         `json:"text,omitempty"`
}

type InboxSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type InboxesResult struct {
	User    string         `json:"user,omitempty"`
	Count   int            `json:"count"`
	Inboxes []InboxSummary `json:"inboxes"`
}

type RootResult struct {
	Version string `json:"version"`
}

type ConfigResult struct {
	Path         string `json:"path"`
	TokenCommand string `json:"token_command,omitempty"`
	User         string `json:"user,omitempty"`
}

type ConfigSetResult struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Path  string `json:"path"`
}

type ConfigPathResult struct {
	Path string `json:"path"`
}
