package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/joshuap/frontapp-cli/internal/api"
	"github.com/joshuap/frontapp-cli/internal/envelope"
	"github.com/spf13/cobra"
)

const defaultLimit = 25

var inboxCmd = &cobra.Command{
	Use:   "inbox [inbox-id]",
	Short: "Search conversations",
	Long: `Search conversations via the Front search API. Defaults to "is:open is:unassigned".
Optionally pass an inbox ID to scope results to a single inbox.

Search syntax supports: is:open, is:archived, is:assigned, is:unassigned,
inbox:<inbox_id>, from:<handle>, to:<handle>, tag:<tag_id>, assignee:<team_id>,
before:<unix_ts>, after:<unix_ts>, contact:<contact_id>, and free text.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}

		query, err := buildSearchQuery(cmd, args)
		if err != nil {
			envelope.FprintError(cmd.OutOrStdout(), "front inbox", err.Error(), "INVALID_INPUT", "Use YYYY-MM-DD format for --before and --after flags", nil)
			return ErrPrinted
		}

		params := &api.SearchConversationsParams{}
		limit, _ := cmd.Flags().GetInt("limit")
		params.Limit = &limit
		if v, _ := cmd.Flags().GetString("page-token"); v != "" {
			params.PageToken = &v
		}

		resp, err := client.SearchConversationsWithResponse(cmd.Context(), query, params)
		if err != nil {
			envelope.FprintError(cmd.OutOrStdout(), "front inbox", err.Error(), "TRANSPORT_ERROR", "Check network connectivity", nil)
			return ErrPrinted
		}
		if resp.JSON200 == nil {
			msg, code, fix := envelope.ClassifyHTTPError(resp.StatusCode(), resp.Body)
			envelope.FprintError(cmd.OutOrStdout(), "front inbox", msg, code, fix, nil)
			return ErrPrinted
		}

		var items []api.ConversationResponse
		if resp.JSON200.UnderscoreResults != nil {
			items = *resp.JSON200.UnderscoreResults
		}

		var total int
		if resp.JSON200.UnderscoreTotal != nil {
			total = *resp.JSON200.UnderscoreTotal
		}

		var nextToken string
		hasNext := resp.JSON200.UnderscorePagination != nil && resp.JSON200.UnderscorePagination.Next != nil
		if hasNext {
			nextToken = envelope.NextPageToken(*resp.JSON200.UnderscorePagination.Next)
		}

		conversations := make([]ConversationSummary, len(items))
		for i, c := range items {
			conversations[i] = mapConversation(c)
		}

		actions := []envelope.Action{}
		if hasNext && nextToken != "" {
			paginationCmd := "front inbox"
			if len(args) > 0 {
				paginationCmd = "front inbox " + args[0]
			}
			paginationParams := map[string]envelope.ParamSpec{
				"--page-token": {Description: "Next page token", Value: nextToken},
			}
			if cmd.Flags().Changed("query") {
				queryVal, _ := cmd.Flags().GetString("query")
				paginationParams["--query"] = envelope.ParamSpec{Description: "Search query", Value: queryVal}
			}
			addCurrentFlags(cmd, paginationParams)
			actions = append(actions, envelope.Action{
				Command:     paginationCmd,
				Description: "Next page of results",
				Params:      paginationParams,
			})
		}
		if len(conversations) > 0 {
			actions = append(actions, envelope.Action{
				Command:     "front read <conversation-id>",
				Description: "Read conversation and messages",
				Params: map[string]envelope.ParamSpec{
					"conversation-id": {Value: items[0].Id, Description: "Conversation ID"},
				},
			})
		}
		actions = append(actions, actionFor("inboxes"))

		envelope.FprintResult(cmd.OutOrStdout(), "front inbox", InboxResult{
			Total:         total,
			Showing:       len(conversations),
			Query:         query,
			Conversations: conversations,
			NextPageToken: nextToken,
		}, actions)
		return nil
	},
}

// buildSearchQuery composes the Front search query from args and flags.
func buildSearchQuery(cmd *cobra.Command, args []string) (string, error) {
	query, _ := cmd.Flags().GetString("query")
	parts := []string{query}

	if len(args) > 0 {
		parts = append(parts, "inbox:"+args[0])
	}
	if v, _ := cmd.Flags().GetString("from"); v != "" {
		parts = append(parts, "from:"+v)
	}
	if v, _ := cmd.Flags().GetString("before"); v != "" {
		ts, err := dateToUnix(v)
		if err != nil {
			return "", err
		}
		parts = append(parts, "before:"+ts)
	}
	if v, _ := cmd.Flags().GetString("after"); v != "" {
		ts, err := dateToUnix(v)
		if err != nil {
			return "", err
		}
		parts = append(parts, "after:"+ts)
	}

	return strings.TrimSpace(strings.Join(parts, " ")), nil
}

// dateToUnix converts a YYYY-MM-DD string to a unix timestamp string.
func dateToUnix(s string) (string, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return "", fmt.Errorf("invalid date %q, expected YYYY-MM-DD format", s)
	}
	return fmt.Sprintf("%d", t.Unix()), nil
}

// mapConversation extracts the essential fields from a ConversationResponse.
func mapConversation(c api.ConversationResponse) ConversationSummary {
	s := ConversationSummary{
		ID:      c.Id,
		Subject: c.Subject,
		Status:  string(c.Status),
		From:    ContactSummary{Handle: c.Recipient.Handle},
	}

	if c.Recipient.Name != nil && *c.Recipient.Name != "" {
		s.From.Name = *c.Recipient.Name
	}

	if c.UpdatedAt != nil {
		s.Date = time.Unix(int64(*c.UpdatedAt), 0).UTC().Format(time.RFC3339)
	} else if c.CreatedAt != nil {
		s.Date = time.Unix(int64(*c.CreatedAt), 0).UTC().Format(time.RFC3339)
	}

	if len(c.Tags) > 0 {
		s.Tags = make([]string, len(c.Tags))
		for i, t := range c.Tags {
			s.Tags[i] = t.Name
		}
	}

	return s
}

func init() {
	inboxCmd.Flags().String("query", "is:open is:unassigned", "Search query (Front search syntax)")
	inboxCmd.Flags().String("from", "", "Filter by sender handle (shortcut for from:<handle> in query)")
	inboxCmd.Flags().String("before", "", "Before date, YYYY-MM-DD (shortcut for before:<ts> in query)")
	inboxCmd.Flags().String("after", "", "After date, YYYY-MM-DD (shortcut for after:<ts> in query)")
	inboxCmd.Flags().Int("limit", defaultLimit, "Maximum number of results to return")
	inboxCmd.Flags().String("page-token", "", "Pagination token for next page")
	inboxCmd.Flags().SetAnnotation("page-token", "internal", []string{"true"})

	rootCmd.AddCommand(inboxCmd)
	registerCommand("inbox", inboxCmd)
}
