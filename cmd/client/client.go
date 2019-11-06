package client

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/ory/viper"

	"github.com/ory/kratos/sdk/go/kratos/client"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
)

func c(cmd *cobra.Command, key string) *client.OryKratos {
	e := flagx.MustGetString(cmd, key)
	if e == "" {
		e = viper.GetString("kratos.urls.admin")
	}
	u, err := url.ParseRequestURI(e)
	cmdx.Must(err, `Unable to parse endpoint URL "%s": %s`, e, err)

	return client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     u.Host,
		BasePath: u.Path,
		Schemes:  []string{u.Scheme},
	})
}
