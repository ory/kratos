// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cliclient

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/spf13/cobra"

	"github.com/spf13/pflag"

	kratos "github.com/ory/kratos/internal/httpclient"
)

const (
	envKeyEndpoint = "KRATOS_ADMIN_URL"
	FlagEndpoint   = "endpoint"
)

type ContextKey int

const (
	ClientContextKey ContextKey = iota + 1
)

type ClientContext struct {
	Endpoint   string
	HTTPClient *http.Client
}

func NewClient(cmd *cobra.Command) (*kratos.APIClient, error) {
	if f, ok := cmd.Context().Value(ClientContextKey).(func(cmd *cobra.Command) (*ClientContext, error)); ok {
		cc, err := f(cmd)
		if err != nil {
			return nil, err
		}

		conf := kratos.NewConfiguration()
		conf.HTTPClient = cc.HTTPClient
		conf.Servers = kratos.ServerConfigurations{{URL: cc.Endpoint}}
		return kratos.NewAPIClient(conf), nil
	} else if f != nil {
		return nil, errors.Errorf("ClientContextKey was expected to be *client.OryKratos but it contained an invalid type %T ", f)
	}

	endpoint, err := cmd.Flags().GetString(FlagEndpoint)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if endpoint == "" {
		endpoint = os.Getenv(envKeyEndpoint)
	}

	if endpoint == "" {
		return nil, errors.Errorf("you have to set the remote endpoint, try --help for details")
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrapf(err, `could not parse the endpoint URL "%s"`, endpoint)
	}

	conf := kratos.NewConfiguration()
	conf.HTTPClient = retryablehttp.NewClient().StandardClient()
	conf.HTTPClient.Timeout = time.Second * 10
	conf.Servers = kratos.ServerConfigurations{{URL: u.String()}}
	return kratos.NewAPIClient(conf), nil
}

func RegisterClientFlags(flags *pflag.FlagSet) {
	flags.StringP(FlagEndpoint, FlagEndpoint[:1], "", fmt.Sprintf("The URL of Ory Kratos' Admin API. Alternatively set using the %s environmental variable.", envKeyEndpoint))
}
