package cmd

import (
	"os"

	"github.com/joshuap/frontapp-cli/internal/api"
	"github.com/joshuap/frontapp-cli/internal/envelope"
	"github.com/spf13/cobra"
)

var inboxesCmd = &cobra.Command{
	Use:   "inboxes",
	Short: "List all inboxes",
	Long: `List all inboxes accessible with the current API token.

Set FRONT_USER to your email address to list only your inboxes
(uses the teammate-scoped endpoint with alt:email: resource alias).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}

		var (
			listResult *api.ListOfInboxes
			statusCode int
			body       []byte
			user       string
		)

		if email := os.Getenv("FRONT_USER"); email != "" {
			user = email
			teammateID := "alt:email:" + email
			resp, err := client.ListTeammateInboxesWithResponse(cmd.Context(), teammateID)
			if err != nil {
				envelope.FprintError(cmd.OutOrStdout(), "front inboxes", err.Error(), "TRANSPORT_ERROR", "Check network connectivity", nil)
				return ErrPrinted
			}
			listResult = resp.JSON200
			statusCode = resp.StatusCode()
			body = resp.Body
		} else {
			resp, err := client.ListInboxesWithResponse(cmd.Context())
			if err != nil {
				envelope.FprintError(cmd.OutOrStdout(), "front inboxes", err.Error(), "TRANSPORT_ERROR", "Check network connectivity", nil)
				return ErrPrinted
			}
			listResult = resp.JSON200
			statusCode = resp.StatusCode()
			body = resp.Body
		}

		if listResult == nil {
			msg, code, fix := envelope.ClassifyHTTPError(statusCode, body)
			envelope.FprintError(cmd.OutOrStdout(), "front inboxes", msg, code, fix, nil)
			return ErrPrinted
		}

		var items []api.InboxResponse
		if listResult.UnderscoreResults != nil {
			items = *listResult.UnderscoreResults
		}

		inboxes := make([]InboxSummary, len(items))
		for i, inbox := range items {
			if inbox.Id != nil {
				inboxes[i].ID = *inbox.Id
			}
			if inbox.Name != nil {
				inboxes[i].Name = *inbox.Name
			}
		}

		actions := []envelope.Action{}
		if len(items) > 0 && items[0].Id != nil {
			actions = append(actions, envelope.Action{
				Command:     "front inbox <inbox-id>",
				Description: "Search conversations in this inbox",
				Params: map[string]envelope.ParamSpec{
					"inbox-id": {Value: *items[0].Id, Description: "Inbox ID"},
				},
			})
		}
		actions = append(actions, actionFor("inbox"))

		envelope.FprintResult(cmd.OutOrStdout(), "front inboxes", InboxesResult{
			User:    user,
			Count:   len(inboxes),
			Inboxes: inboxes,
		}, actions)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inboxesCmd)
	registerCommand("inboxes", inboxesCmd)
}
