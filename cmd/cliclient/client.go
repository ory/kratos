package cliclient

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/spf13/pflag"

	"github.com/ory/kratos/internal/httpclient/client"
	"github.com/ory/x/cmdx"
)

const (
	envKeyEndpoint = "KRATOS_ADMIN_ENDPOINT"
	FlagEndpoint   = "endpoint"
)

func NewClient(cmd *cobra.Command) *client.OryKratos {
	endpoint, err := cmd.Flags().GetString(FlagEndpoint)
	cmdx.Must(err, "flag access error: %s", err)

	if endpoint == "" {
		endpoint = os.Getenv(envKeyEndpoint)
	}

	if endpoint == "" {
		// no endpoint is set
		fmt.Fprintln(os.Stderr, "You have to set the remote endpoint, try --help for details.")
		os.Exit(1)
	}

	u, err := url.Parse(endpoint)
	cmdx.Must(err, `Could not parse the endpoint URL "%s".`, endpoint)

	return client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     u.Host,
		BasePath: u.Path,
		Schemes:  []string{u.Scheme},
	})
}

func RegisterClientFlags(flags *pflag.FlagSet) {
	flags.StringP(FlagEndpoint, FlagEndpoint[:1], "", fmt.Sprintf("The upstream admin endpoint URL. Alternatively set using the %s environmental variable.", envKeyEndpoint))
}
