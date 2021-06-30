package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"

	"github.com/ory/kratos/x"

	"github.com/ory/kratos/internal/testhelpers"

	ory "github.com/ory/kratos-client-go"
)

func PrintJSONPretty(v interface{}) {
	out, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(out))
}

func TestClient(t *testing.T) *ory.APIClient {
	publicURL, _ := testhelpers.StartE2EServer(t, "../../pkg/stub/kratos.yaml")
	return NewSDKForSelfHosted(publicURL)
}

func NewSDK(project string) *ory.APIClient {
	return NewSDKForSelfHosted(fmt.Sprintf("https://%s.projects.oryapis.com/api/kratos/public", project))
}

func NewSDKForSelfHosted(endpoint string) *ory.APIClient {
	conf := ory.NewConfiguration()
	conf.Servers = ory.ServerConfigurations{{URL: endpoint}}
	cj, _ := cookiejar.New(nil)
	conf.HTTPClient = &http.Client{Jar: cj}
	return ory.NewAPIClient(conf)
}

func ExitOnError(err error) {
	if err == nil {
		return
	}
	out, _ := json.MarshalIndent(err, "", "  ")
	fmt.Printf("%s\n\nAn error occurred: %+v\n", out, err)
	os.Exit(1)
}

func SDKExitOnError(err error, res *http.Response) {
	if err == nil {
		return
	}
	body := x.MustReadAll(res.Body)
	out, _ := json.MarshalIndent(err, "", "  ")
	fmt.Printf("%s\n\nAn error occurred: %+v\nbody: %s\n", out, err, body)
	os.Exit(1)
}
