package cliclient

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/pflag"

	"github.com/ory/kratos/internal/httpclient/client"
	"github.com/ory/x/cmdx"
)

const (
	envKeyEndpoint = "KRATOS_ADMIN_ENDPOINT"
)

var endpoint string

func NewClient() *client.OryKratos {
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
	flags.StringVarP(&endpoint, "endpoint", "e", "", fmt.Sprintf("The upstream admin endpoint URL. Alternatively set using the %s environmental variable.", envKeyEndpoint))
}
