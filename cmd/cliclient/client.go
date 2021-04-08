package cliclient

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/spf13/pflag"

	"github.com/ory/kratos-client-go/client"
	"github.com/ory/x/cmdx"
)

const (
	envKeyEndpoint = "KRATOS_ADMIN_URL"
	FlagEndpoint   = "endpoint"
)

type ContextKey int

const (
	ClientContextKey ContextKey = iota + 1
	HTTPClientContextKey
)

func NewHTTPClient(cmd *cobra.Command) *http.Client {
	if f, ok := cmd.Context().Value(HTTPClientContextKey).(func(cmd *cobra.Command) *http.Client); ok {
		return f(cmd)
	} else if f != nil {
		panic(fmt.Sprintf("ClientContextKey was expected to be *http.Client but it contained an invalid type %T ", f))
	}
	return &http.Client{Transport: http.DefaultTransport}
}

func NewClient(cmd *cobra.Command) *client.OryKratos {
	if f, ok := cmd.Context().Value(ClientContextKey).(func(cmd *cobra.Command) *client.OryKratos); ok {
		return f(cmd)
	} else if f != nil {
		panic(fmt.Sprintf("ClientContextKey was expected to be *client.OryKratos but it contained an invalid type %T ", f))
	}

	endpoint, err := cmd.Flags().GetString(FlagEndpoint)
	cmdx.Must(err, "flag access error: %s", err)

	if endpoint == "" {
		endpoint = os.Getenv(envKeyEndpoint)
	}

	if endpoint == "" {
		// no endpoint is set
		_, _ = fmt.Fprintln(os.Stderr, "You have to set the remote endpoint, try --help for details.")
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
	flags.StringP(FlagEndpoint, FlagEndpoint[:1], "", fmt.Sprintf("The URL of ORY Kratos' Admin API. Alternatively set using the %s environmental variable.", envKeyEndpoint))
}
