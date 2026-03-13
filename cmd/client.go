package cmd

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/joshuap/frontapp-cli/internal/api"
	"github.com/joshuap/frontapp-cli/internal/envelope"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const defaultBaseURL = "https://api2.frontapp.com"

func newClient(cmd *cobra.Command) (*api.ClientWithResponses, error) {
	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		token = os.Getenv("FRONT_API_TOKEN")
	}
	if token == "" {
		return nil, errNoToken
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}

	auth := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "front-cli/"+Version)
		return nil
	}

	return api.NewClientWithResponses(defaultBaseURL,
		api.WithHTTPClient(httpClient),
		api.WithRequestEditorFn(auth))
}

type noTokenError struct{}

func (e noTokenError) Error() string {
	return "no API token provided"
}

var errNoToken = noTokenError{}

// addCurrentFlags adds all explicitly-set flags to a pagination params map,
// preserving their current values so the next-page action carries them forward.
// Skips flags annotated with "internal"="true" (e.g. token, page-token).
func addCurrentFlags(cmd *cobra.Command, params map[string]envelope.ParamSpec) {
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if vals, ok := f.Annotations["internal"]; ok && len(vals) > 0 && vals[0] == "true" {
			return
		}
		key := "--" + f.Name
		if _, exists := params[key]; exists {
			return
		}
		params[key] = envelope.ParamSpec{
			Description: f.Usage,
			Value:       f.Value.String(),
		}
	})
}
