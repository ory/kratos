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
	flagEndpoint   = "endpoint"
)

func NewClient(cmd *cobra.Command) *client.OryKratos {
	endpoint, err := cmd.Flags().GetString(flagEndpoint)
	cmdx.Must(err, "flag access error: %s", err)

	if endpoint == "" {
		endpoint = os.Getenv(envKeyEndpoint)
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
	flags.StringP(flagEndpoint, flagEndpoint[:1], "", fmt.Sprintf("The upstream admin endpoint URL. Alternatively set using the %s environmental variable.", envKeyEndpoint))
}
