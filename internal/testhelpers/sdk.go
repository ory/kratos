// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	kratos "github.com/ory/kratos/internal/httpclient"
)

func NewSDKClient(ts *httptest.Server) *kratos.APIClient {
	return NewSDKClientFromURL(ts.URL)
}

func NewSDKCustomClient(ts *httptest.Server, client *http.Client) *kratos.APIClient {
	conf := kratos.NewConfiguration()
	conf.Servers = kratos.ServerConfigurations{{URL: ts.URL}}
	conf.HTTPClient = client
	return kratos.NewAPIClient(conf)
}

func NewSDKClientFromURL(u string) *kratos.APIClient {
	conf := kratos.NewConfiguration()
	conf.Servers = kratos.ServerConfigurations{{URL: u}}
	return kratos.NewAPIClient(conf)
}

func SDKFormFieldsToURLValues(ff []kratos.UiNode) url.Values {
	values := url.Values{}
	for _, f := range ff {
		attr := f.Attributes.UiNodeInputAttributes
		if attr == nil {
			continue
		}

		val := attr.Value
		if val == nil {
			continue
		}

		switch v := val.(type) {
		case bool:
			values.Set(attr.Name, fmt.Sprintf("%v", v))
		case string:
			values.Set(attr.Name, fmt.Sprintf("%v", v))
		case float32:
			values.Set(attr.Name, fmt.Sprintf("%v", v))
		case float64:
			values.Set(attr.Name, fmt.Sprintf("%v", v))
		}
	}
	return values
}
