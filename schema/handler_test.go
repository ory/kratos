package schema_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gobuffalo/httptest"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/schema"
	"github.com/ory/viper"
)

func RegisterDefaultSchema(ts *httptest.Server) {
	var b bytes.Buffer
	defaultSchema := &schema.Schema{
		URL: viper.GetString(configuration.ViperKeyDefaultIdentityTraitsSchemaURL),
	}
	_ = json.NewEncoder(&b).Encode(defaultSchema)
	req, _ := http.NewRequest("PUT", ts.URL+"/schemas", &b)
	_, _ = ts.Client().Do(req)
}

func TestHandler(t *testing.T) {

}
