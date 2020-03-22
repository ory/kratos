package testhelpers

import (
	"fmt"
	"net/http/httptest"
	"net/url"

	"github.com/ory/x/pointerx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/internal/httpclient/client"
	"github.com/ory/kratos/internal/httpclient/models"
)

func NewSDKClient(ts *httptest.Server) *client.OryKratos {
	return NewSDKClientFromURL(ts.URL)
}

func NewSDKClientFromURL(u string) *client.OryKratos {
	return client.NewHTTPClientWithConfig(nil,
		&client.TransportConfig{Host: urlx.ParseOrPanic(u).Host, BasePath: "/", Schemes: []string{"http"}})
}

func SDKFormFieldsToURLValues(ff models.FormFields) url.Values {
	values := url.Values{}
	for _, f := range ff {
		values.Set(pointerx.StringR(f.Name), fmt.Sprintf("%v", f.Value))
	}
	return values
}
