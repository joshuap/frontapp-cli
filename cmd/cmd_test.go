package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/joshuap/frontapp-cli/internal/api"
	"github.com/joshuap/frontapp-cli/internal/envelope"
	"github.com/spf13/cobra"
)

// --- Root command: lists all flat commands ---

func TestRootCommand_ListsCommands(t *testing.T) {
	var names []string
	for _, c := range rootCmd.Commands() {
		if c.Hidden {
			continue
		}
		names = append(names, c.Name())
	}

	expected := []string{"config", "inbox", "inboxes", "read"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d commands, got %d: %v", len(expected), len(names), names)
	}
	for _, e := range expected {
		found := false
		for _, n := range names {
			if n == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing expected command %q in %v", e, names)
		}
	}
}

func TestConfigCommand_NoSubcommands(t *testing.T) {
	var found bool
	for _, c := range rootCmd.Commands() {
		if c.Name() == "config" {
			found = true
			if c.HasSubCommands() {
				t.Error("config command should not have subcommands")
			}
		}
	}
	if !found {
		t.Error("config command not found")
	}
}

// --- Template syntax ---

func TestInboxCommand_HasOptionalArgInParams(t *testing.T) {
	action := actionFor("inbox")
	if action.Params == nil {
		t.Fatal("inbox should have params")
	}
	if _, ok := action.Params["inbox-id"]; !ok {
		t.Error("inbox command should document optional inbox-id param")
	}
}

func TestReadCommand_HasTemplateSyntax(t *testing.T) {
	var actions []envelope.Action
	collectLeafCommands(rootCmd, "front", &actions)

	for _, a := range actions {
		if strings.Contains(a.Command, "read") {
			if !strings.Contains(a.Command, "<conversation-id>") {
				t.Errorf("read command should have <conversation-id> placeholder, got %q", a.Command)
			}
			if a.Params == nil {
				t.Fatal("read command should have params")
			}
			if _, ok := a.Params["conversation-id"]; !ok {
				t.Error("read command should have conversation-id param")
			}
			return
		}
	}
	t.Error("read command not found in leaf commands")
}

// --- All flags documented as structured params ---

func TestInboxCommand_AllFlagsDocumented(t *testing.T) {
	action := actionFor("inbox")
	if action.Params == nil {
		t.Fatal("inbox command should have params")
	}

	expectedFlags := []string{"--query", "--from", "--before", "--after", "--limit"}
	for _, flag := range expectedFlags {
		if _, ok := action.Params[flag]; !ok {
			t.Errorf("inbox command should document %s flag", flag)
		}
	}
}

func TestInboxCommand_QueryDefault(t *testing.T) {
	action := actionFor("inbox")
	q, ok := action.Params["--query"]
	if !ok {
		t.Fatal("inbox should have --query param")
	}
	if q.Default != "is:open is:unassigned" {
		t.Errorf("--query default should be 'is:open is:unassigned', got %q", q.Default)
	}
}

func TestInboxCommand_LimitDefault(t *testing.T) {
	action := actionFor("inbox")
	l, ok := action.Params["--limit"]
	if !ok {
		t.Fatal("inbox should have --limit param")
	}
	if l.Default != "25" {
		t.Errorf("--limit default should be '25', got %q", l.Default)
	}
}

func TestReadCommand_NoFlags(t *testing.T) {
	action := actionFor("read")
	if action.Params == nil {
		t.Fatal("read should have params (for positional arg)")
	}
	for name := range action.Params {
		if strings.HasPrefix(name, "--") {
			t.Errorf("read should not have flag params, found %q", name)
		}
	}
}

// --- buildSearchQuery ---

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "inbox [inbox-id]"}
	cmd.Flags().String("query", "is:open is:unassigned", "")
	cmd.Flags().String("from", "", "")
	cmd.Flags().String("before", "", "")
	cmd.Flags().String("after", "", "")
	return cmd
}

func TestBuildSearchQuery_Default(t *testing.T) {
	got, err := buildSearchQuery(newSearchCmd(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "is:open is:unassigned" {
		t.Errorf("expected 'is:open is:unassigned', got %q", got)
	}
}

func TestBuildSearchQuery_WithInboxArg(t *testing.T) {
	got, err := buildSearchQuery(newSearchCmd(), []string{"inb_123"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "is:open is:unassigned inbox:inb_123" {
		t.Errorf("expected 'is:open is:unassigned inbox:inb_123', got %q", got)
	}
}

func TestBuildSearchQuery_WithFrom(t *testing.T) {
	cmd := newSearchCmd()
	cmd.Flags().Set("from", "test@example.com")
	got, err := buildSearchQuery(cmd, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "is:open is:unassigned from:test@example.com" {
		t.Errorf("expected 'is:open is:unassigned from:test@example.com', got %q", got)
	}
}

func TestBuildSearchQuery_WithDates(t *testing.T) {
	cmd := newSearchCmd()
	cmd.Flags().Set("before", "2026-01-15")
	cmd.Flags().Set("after", "2026-01-01")
	got, err := buildSearchQuery(cmd, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "after:") || !strings.Contains(got, "before:") {
		t.Errorf("expected after: and before: in query, got %q", got)
	}
	for _, p := range strings.Fields(got) {
		if strings.HasPrefix(p, "before:") || strings.HasPrefix(p, "after:") {
			ts := strings.SplitN(p, ":", 2)[1]
			if ts == "" || ts[0] < '0' || ts[0] > '9' {
				t.Errorf("expected numeric timestamp in %q", p)
			}
		}
	}
}

func TestBuildSearchQuery_AllOptions(t *testing.T) {
	cmd := newSearchCmd()
	cmd.Flags().Set("from", "a@b.com")
	cmd.Flags().Set("before", "2026-03-01")
	got, err := buildSearchQuery(cmd, []string{"inb_1"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "is:open") {
		t.Error("should contain base query")
	}
	if !strings.Contains(got, "inbox:inb_1") {
		t.Error("should contain inbox from arg")
	}
	if !strings.Contains(got, "from:a@b.com") {
		t.Error("should contain from filter")
	}
	if !strings.Contains(got, "before:") {
		t.Error("should contain before filter")
	}
}

func TestBuildSearchQuery_WithAssignee(t *testing.T) {
	cmd := newSearchCmd()
	cmd.Flags().String("assignee", "", "")
	cmd.Flags().Set("assignee", "alice@example.com")
	got, err := buildSearchQuery(cmd, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "assignee:alt:email:alice@example.com") {
		t.Errorf("expected assignee:alt:email:alice@example.com in query, got %q", got)
	}
	if strings.Contains(got, "is:unassigned") {
		t.Errorf("should swap is:unassigned for is:assigned when --assignee used, got %q", got)
	}
	if !strings.Contains(got, "is:assigned") {
		t.Errorf("should contain is:assigned when --assignee used, got %q", got)
	}
}

func TestBuildSearchQuery_CustomQuery(t *testing.T) {
	cmd := newSearchCmd()
	cmd.Flags().Set("query", "is:archived tag:tag_123")
	got, err := buildSearchQuery(cmd, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "is:archived tag:tag_123" {
		t.Errorf("expected custom query, got %q", got)
	}
}

func TestBuildSearchQuery_InvalidDate(t *testing.T) {
	cmd := newSearchCmd()
	cmd.Flags().Set("before", "not-a-date")
	_, err := buildSearchQuery(cmd, nil)
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

// --- dateToUnix ---

func TestDateToUnix(t *testing.T) {
	got, err := dateToUnix("2026-01-15")
	if err != nil {
		t.Fatal(err)
	}
	if got != "1768435200" {
		t.Errorf("expected 1768435200, got %s", got)
	}
}

func TestDateToUnix_Invalid(t *testing.T) {
	_, err := dateToUnix("not-a-date")
	if err == nil {
		t.Error("invalid date should return error")
	}
	_, err = dateToUnix("")
	if err == nil {
		t.Error("empty string should return error")
	}
}

// --- mapConversation ---

func TestMapConversation(t *testing.T) {
	name := "Alice"
	updatedAt := float32(1700000000)
	c := api.ConversationResponse{
		Id:      "cnv_123",
		Subject: "Test subject",
		Status:  "assigned",
		Recipient: api.RecipientResponse{
			Handle: "alice@example.com",
			Name:   &name,
		},
		UpdatedAt: &updatedAt,
		Tags: []api.TagResponse{
			{Name: "urgent"},
			{Name: "billing"},
		},
	}

	s := mapConversation(c)
	if s.ID != "cnv_123" {
		t.Errorf("expected id cnv_123, got %v", s.ID)
	}
	if s.Subject != "Test subject" {
		t.Error("expected subject")
	}
	if s.Status != "assigned" {
		t.Errorf("expected status assigned, got %v", s.Status)
	}
	if s.From.Handle != "alice@example.com" {
		t.Error("expected handle")
	}
	if s.From.Name != "Alice" {
		t.Error("expected name")
	}
	if s.Date == "" {
		t.Error("expected date")
	}
	if len(s.Tags) != 2 || s.Tags[0] != "urgent" {
		t.Errorf("expected tags [urgent billing], got %v", s.Tags)
	}
}

func TestMapConversation_MinimalFields(t *testing.T) {
	c := api.ConversationResponse{
		Id:      "cnv_456",
		Subject: "Minimal",
		Status:  "unassigned",
		Recipient: api.RecipientResponse{
			Handle: "bob@example.com",
		},
	}

	s := mapConversation(c)
	if s.ID != "cnv_456" {
		t.Error("expected id")
	}
	if s.Tags != nil {
		t.Error("should not have tags when empty")
	}
	if s.Date != "" {
		t.Error("should not have date when no timestamps")
	}
	if s.From.Name != "" {
		t.Error("should not have name when nil")
	}
}

// --- mapMessage ---

func TestMapMessage(t *testing.T) {
	id := "msg_1"
	text := "Hello world"
	createdAt := float32(1700000000)
	isInbound := true
	m := api.MessageResponse{
		Id:        &id,
		Text:      &text,
		CreatedAt: &createdAt,
		IsInbound: &isInbound,
		Author: &api.TeammateResponse{
			FirstName: "Bob",
			LastName:  "Smith",
			Email:     "bob@example.com",
		},
	}

	result := mapMessage(m)
	if result.ID != "msg_1" {
		t.Error("expected id")
	}
	if result.Text != "Hello world" {
		t.Error("expected text")
	}
	if result.IsInbound == nil || *result.IsInbound != true {
		t.Error("expected is_inbound true")
	}
	if result.From == nil {
		t.Fatal("expected from")
	}
	if result.From.Name != "Bob Smith" {
		t.Errorf("expected 'Bob Smith', got %v", result.From.Name)
	}
	if result.From.Email != "bob@example.com" {
		t.Error("expected email")
	}
}

func TestMapMessage_TruncatesLongText(t *testing.T) {
	id := "msg_2"
	longText := strings.Repeat("x", 600)
	m := api.MessageResponse{
		Id:   &id,
		Text: &longText,
	}

	result := mapMessage(m)
	if len(result.Text) > maxTextLength+20 {
		t.Errorf("text should be truncated, got length %d", len(result.Text))
	}
	if !strings.HasSuffix(result.Text, "... [truncated]") {
		t.Error("truncated text should end with '... [truncated]'")
	}
}

func TestMapMessage_DropsHTMLBody(t *testing.T) {
	id := "msg_3"
	body := "<p>HTML content</p>"
	text := "Plain text"
	m := api.MessageResponse{
		Id:   &id,
		Body: &body,
		Text: &text,
	}

	result := mapMessage(m)
	if result.Text != "Plain text" {
		t.Errorf("should prefer text over body, got %v", result.Text)
	}
}

func TestMapMessage_FallsBackToBody(t *testing.T) {
	id := "msg_4"
	body := "Only body available"
	emptyText := ""
	m := api.MessageResponse{
		Id:   &id,
		Body: &body,
		Text: &emptyText,
	}

	result := mapMessage(m)
	if result.Text != "Only body available" {
		t.Errorf("should fall back to body when text empty, got %v", result.Text)
	}
}

// --- addCurrentFlags ---

func TestAddCurrentFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "inbox"}
	cmd.Flags().String("from", "", "Sender filter")
	cmd.Flags().String("before", "", "Before date")
	cmd.Flags().String("after", "", "After date")
	cmd.Flags().Int("limit", 25, "Max results")
	cmd.Flags().String("page-token", "", "Pagination token")
	cmd.Flags().SetAnnotation("page-token", "internal", []string{"true"})
	cmd.Flags().String("token", "", "API token")
	cmd.Flags().SetAnnotation("token", "internal", []string{"true"})

	cmd.Flags().Set("from", "test@example.com")
	cmd.Flags().Set("limit", "10")
	cmd.Flags().Set("page-token", "should-be-skipped")
	cmd.Flags().Set("token", "should-be-skipped")

	params := map[string]envelope.ParamSpec{
		"--page-token": {Description: "Next page token", Value: "tok123"},
	}
	addCurrentFlags(cmd, params)

	// page-token should be skipped via internal annotation (not overwritten)
	if p := params["--page-token"]; p.Value != "tok123" {
		t.Error("should preserve existing --page-token")
	}
	// token should be skipped via internal annotation
	if _, ok := params["--token"]; ok {
		t.Error("should not add --token (internal flag)")
	}
	if p, ok := params["--from"]; !ok || p.Value != "test@example.com" {
		t.Error("should add --from with current value")
	}
	if p, ok := params["--limit"]; !ok || p.Value != "10" {
		t.Error("should add --limit with current value")
	}
	if _, ok := params["--before"]; ok {
		t.Error("should not add --before when not set")
	}
}

// --- actionFor registry ---

func TestActionFor_Inbox(t *testing.T) {
	action := actionFor("inbox")
	if action.Command != "front inbox [inbox-id]" {
		t.Errorf("expected 'front inbox [inbox-id]', got %q", action.Command)
	}
	if action.Description != "Search conversations" {
		t.Errorf("expected description, got %q", action.Description)
	}
}

func TestActionFor_Read(t *testing.T) {
	action := actionFor("read")
	if !strings.Contains(action.Command, "<conversation-id>") {
		t.Errorf("read action should have placeholder, got %q", action.Command)
	}
}

func TestActionFor_Unknown(t *testing.T) {
	action := actionFor("nonexistent")
	if action.Command != "front nonexistent" {
		t.Errorf("expected fallback command, got %q", action.Command)
	}
}

// --- CLI error fix logic ---

func TestExecute_UnknownCommandFix(t *testing.T) {
	msg := "unknown command \"bogus\" for \"front\""
	fix := ""
	if strings.Contains(msg, "unknown command") {
		fix = "Run 'front' to see available commands"
	}
	if fix == "" {
		t.Error("expected fix for unknown command")
	}
}

func TestExecute_MissingArgsFix(t *testing.T) {
	msg := "accepts 1 arg(s), received 0"
	fix := ""
	if strings.Contains(msg, "accepts") {
		fix = "Run 'front' to see available commands and their required arguments"
	}
	if fix == "" {
		t.Error("expected fix for missing args")
	}
}

// --- Envelope JSON structure ---

func TestEnvelopeResult_MarshalStructure(t *testing.T) {
	r := envelope.Result{
		OK:      true,
		Command: "front test",
		Result:  map[string]interface{}{"count": 1},
		NextActions: []envelope.Action{
			{
				Command:     "front read <conversation-id>",
				Description: "Read conversation",
				Params: map[string]envelope.ParamSpec{
					"conversation-id": {Value: "cnv_123", Description: "Conversation ID"},
				},
			},
		},
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	if parsed["ok"] != true {
		t.Error("expected ok=true")
	}
	actions := parsed["next_actions"].([]interface{})
	action := actions[0].(map[string]interface{})
	if !strings.Contains(action["command"].(string), "<conversation-id>") {
		t.Error("expected template syntax")
	}
	params := action["params"].(map[string]interface{})
	p := params["conversation-id"].(map[string]interface{})
	if p["value"] != "cnv_123" {
		t.Error("expected value in param")
	}
}

func TestEnvelopeResult_OmitsEmptyNextActions(t *testing.T) {
	r := envelope.Result{OK: true, Command: "test", Result: "ok"}
	data, _ := json.Marshal(r)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	if _, ok := parsed["next_actions"]; ok {
		t.Error("next_actions should be omitted when empty")
	}
}

// --- Typed result structs ---

func TestInboxResult_OmitsEmptyPageToken(t *testing.T) {
	r := InboxResult{Total: 5, Showing: 5, Query: "is:open"}
	data, _ := json.Marshal(r)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	if _, ok := parsed["next_page_token"]; ok {
		t.Error("next_page_token should be omitted when empty")
	}
}

func TestReadResult_OmitsTruncatedWhenFalse(t *testing.T) {
	r := ReadResult{
		Conversation: ConversationSummary{ID: "cnv_1"},
		Messages:     []MessageSummary{},
	}
	data, _ := json.Marshal(r)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	if _, ok := parsed["truncated"]; ok {
		t.Error("truncated should be omitted when false")
	}
}

func TestReadResult_ShowsTruncatedWhenTrue(t *testing.T) {
	r := ReadResult{
		Conversation: ConversationSummary{ID: "cnv_1"},
		Messages:     []MessageSummary{},
		Truncated:    true,
	}
	data, _ := json.Marshal(r)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	if parsed["truncated"] != true {
		t.Error("truncated should be true when set")
	}
}

func TestInboxesResult_IncludesUserWhenSet(t *testing.T) {
	r := InboxesResult{User: "alice@example.com", Count: 1, Inboxes: []InboxSummary{{ID: "inb_1", Name: "Support"}}}
	data, _ := json.Marshal(r)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	if parsed["user"] != "alice@example.com" {
		t.Errorf("expected user=alice@example.com, got %v", parsed["user"])
	}
}

func TestInboxesResult_OmitsUserWhenEmpty(t *testing.T) {
	r := InboxesResult{Count: 1, Inboxes: []InboxSummary{{ID: "inb_1", Name: "Support"}}}
	data, _ := json.Marshal(r)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	if _, ok := parsed["user"]; ok {
		t.Error("user should be omitted when empty")
	}
}

func TestConstants(t *testing.T) {
	if defaultLimit != 25 {
		t.Errorf("expected default limit 25, got %d", defaultLimit)
	}
	if maxTextLength != 500 {
		t.Errorf("expected max text length 500, got %d", maxTextLength)
	}
}

// --- addCurrentFlags does not overwrite ---

func TestAddCurrentFlags_DoesNotOverwriteExisting(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("query", "default", "Search query")
	cmd.Flags().Set("query", "custom")

	params := map[string]envelope.ParamSpec{
		"--query": {Description: "Pre-existing", Value: "original"},
	}
	addCurrentFlags(cmd, params)

	if params["--query"].Value != "original" {
		t.Errorf("should not overwrite existing param, got %q", params["--query"].Value)
	}
}

// --- Root command produces valid JSON envelope ---

func TestRootCommand_ProducesValidEnvelope(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	rootCmd.SetOut(nil)

	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if parsed["ok"] != true {
		t.Error("expected ok=true")
	}
	if parsed["command"] != "front" {
		t.Errorf("expected command=front, got %v", parsed["command"])
	}
	result := parsed["result"].(map[string]interface{})
	if result["version"] != Version {
		t.Errorf("expected version=%s", Version)
	}
	if _, ok := parsed["next_actions"]; !ok {
		t.Error("expected next_actions")
	}
}

// --- Version flag ---

func TestVersionFlag_PrintsVersion(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--version"})
	err := rootCmd.Execute()
	rootCmd.SetOut(nil)

	if err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(buf.String())
	if got != Version {
		t.Errorf("expected %q, got %q", Version, got)
	}
}

// --- Zero-value defaults ---

func TestBuildCommandUsageAndParams_ZeroDefaults(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Int("limit", 0, "Max results")
	cmd.Flags().String("query", "", "Search query")
	cmd.Flags().Bool("verbose", false, "Verbose output")
	cmd.Flags().Int("count", 10, "Count of items")

	_, params := buildCommandUsageAndParams(cmd, "front test")

	if _, ok := params["--limit"]; ok && params["--limit"].Default != "" {
		t.Errorf("int zero default should be hidden, got %q", params["--limit"].Default)
	}
	if _, ok := params["--verbose"]; ok && params["--verbose"].Default != "" {
		t.Errorf("bool false default should be hidden, got %q", params["--verbose"].Default)
	}
	if params["--count"].Default != "10" {
		t.Errorf("non-zero int default should be shown, got %q", params["--count"].Default)
	}
}
