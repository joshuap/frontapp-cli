package cmd

import (
	"context"
	"time"

	"github.com/joshuap/frontapp-cli/internal/api"
	"github.com/joshuap/frontapp-cli/internal/envelope"
	"github.com/spf13/cobra"
)

const maxTextLength = 500

var readCmd = &cobra.Command{
	Use:   "read <conversation-id>",
	Short: "Read a conversation and its messages",
	Long:  "Fetch a conversation and its messages in one call. Shows plain text only (no HTML), truncated to 500 chars.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}

		conversationID := args[0]
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		type msgResult struct {
			resp *api.ListConversationMessagesResponse
			err  error
		}
		msgCh := make(chan msgResult, 1)
		go func() {
			msgLimit := defaultLimit
			resp, err := client.ListConversationMessagesWithResponse(ctx, conversationID, &api.ListConversationMessagesParams{
				Limit: &msgLimit,
			})
			msgCh <- msgResult{resp, err}
		}()

		convResp, err := client.GetConversationByIdWithResponse(ctx, conversationID)
		if err != nil {
			envelope.FprintError(cmd.OutOrStdout(), "front read", err.Error(), "TRANSPORT_ERROR", "Check network connectivity", nil)
			return ErrPrinted
		}
		if convResp.JSON200 == nil {
			msg, code, fix := envelope.ClassifyHTTPError(convResp.StatusCode(), convResp.Body)
			envelope.FprintError(cmd.OutOrStdout(), "front read", msg, code, fix, nil)
			return ErrPrinted
		}

		mr := <-msgCh
		if mr.err != nil {
			envelope.FprintError(cmd.OutOrStdout(), "front read", mr.err.Error(), "TRANSPORT_ERROR", "Check network connectivity", nil)
			return ErrPrinted
		}
		msgResp := mr.resp
		if msgResp.JSON200 == nil {
			msg, code, fix := envelope.ClassifyHTTPError(msgResp.StatusCode(), msgResp.Body)
			envelope.FprintError(cmd.OutOrStdout(), "front read", msg, code, fix, nil)
			return ErrPrinted
		}

		var rawMessages []api.MessageResponse
		if msgResp.JSON200.UnderscoreResults != nil {
			rawMessages = *msgResp.JSON200.UnderscoreResults
		}

		truncated := msgResp.JSON200.UnderscorePagination != nil && msgResp.JSON200.UnderscorePagination.Next != nil

		messages := make([]MessageSummary, len(rawMessages))
		for i, m := range rawMessages {
			messages[i] = mapMessage(m)
		}

		actions := []envelope.Action{
			{
				Command:     "front read <conversation-id>",
				Description: "Refresh this conversation",
				Params: map[string]envelope.ParamSpec{
					"conversation-id": {Value: conversationID, Description: "Conversation ID"},
				},
			},
			actionFor("inbox"),
			actionFor("inboxes"),
		}

		envelope.FprintResult(cmd.OutOrStdout(), "front read", ReadResult{
			Conversation: mapConversation(*convResp.JSON200),
			Messages:     messages,
			Truncated:    truncated,
		}, actions)
		return nil
	},
}

// mapMessage extracts essential fields from a MessageResponse.
// Drops HTML body, keeps text only, truncated to maxTextLength.
func mapMessage(m api.MessageResponse) MessageSummary {
	msg := MessageSummary{}

	if m.Id != nil {
		msg.ID = *m.Id
	}

	if m.Author != nil {
		name := m.Author.FirstName
		if m.Author.LastName != "" {
			name += " " + m.Author.LastName
		}
		msg.From = &ContactSummary{
			Name:  name,
			Email: m.Author.Email,
		}
	}

	if m.CreatedAt != nil {
		msg.Date = time.Unix(int64(*m.CreatedAt), 0).UTC().Format(time.RFC3339)
	}

	msg.IsInbound = m.IsInbound

	// Use text (plain), fall back to body (HTML) — truncate either way
	body := ""
	if m.Text != nil && *m.Text != "" {
		body = *m.Text
	} else if m.Body != nil && *m.Body != "" {
		body = *m.Body
	}
	if len(body) > maxTextLength {
		body = body[:maxTextLength] + "... [truncated]"
	}
	msg.Text = body

	return msg
}

func init() {
	rootCmd.AddCommand(readCmd)
	registerCommand("read", readCmd)
}
