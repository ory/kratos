package hydraclient

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"
	hydraclientgo "github.com/ory/hydra-client-go"
)

func GetHydraAdminAPIClient(hydra_admin_url string) hydraclientgo.AdminApi {
	configuration := hydraclientgo.NewConfiguration()
	configuration.Servers = hydraclientgo.ServerConfigurations{{
		URL: hydra_admin_url,
	}}
	return hydraclientgo.NewAPIClient(configuration).AdminApi
}

func GetHydraLoginRequest(hydra_admin_url string, hlc uuid.UUID) (*hydraclientgo.LoginRequest, error) {
	resp, r, err := GetHydraAdminAPIClient(hydra_admin_url).GetLoginRequest(context.Background()).LoginChallenge(fmt.Sprintf("%x", hlc)).Execute()
	if err != nil {
		return nil, herodot.ErrInternalServerError.
			WithError(err.Error()).
			WithDetail("status_code", r.StatusCode).
			WithDebugf("hydra_admin_url=%s", hydra_admin_url)
	}
	return resp, nil
}

func AcceptHydraLoginRequest(hydra_admin_url string, hlc uuid.UUID, sub string) (*hydraclientgo.CompletedRequest, error) {
	adminClient := GetHydraAdminAPIClient(hydra_admin_url)
	// TODO remember...
	alr := hydraclientgo.NewAcceptLoginRequest(sub)
	resp, r, err := adminClient.AcceptLoginRequest(context.Background()).LoginChallenge(fmt.Sprintf("%x", hlc)).AcceptLoginRequest(*alr).Execute()
	if err != nil {
		return nil, herodot.ErrInternalServerError.
			WithError(err.Error()).
			WithDetail("status_code", r.StatusCode).
			WithDebugf("hydra_admin_url=%s", hydra_admin_url)
	}
	return resp, nil
}
