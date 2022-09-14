package hydraclient

import (
	"context"
	"fmt"
	"github.com/pkg/errors"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"
	hydraclientgo "github.com/ory/hydra-client-go"
	"github.com/ory/kratos/session"
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
		return nil, errors.WithStack(herodot.ErrInternalServerError.
			WithError(err.Error()).
			WithDetail("status_code", r.StatusCode))
	}
	return resp, nil
}

func AcceptHydraLoginRequest(hydra_admin_url string, hlc uuid.UUID, sub string, remember bool, remember_for int64, amr session.AuthenticationMethods) (*hydraclientgo.CompletedRequest, error) {
	adminClient := GetHydraAdminAPIClient(hydra_admin_url)
	alr := hydraclientgo.NewAcceptLoginRequest(sub)
	alr.Remember = &remember
	alr.RememberFor = &remember_for
	alr.Amr = []string{}
	for _, r := range amr {
		alr.Amr = append(alr.Amr, string(r.Method))
	}

	resp, r, err := adminClient.AcceptLoginRequest(context.Background()).LoginChallenge(fmt.Sprintf("%x", hlc)).AcceptLoginRequest(*alr).Execute()
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.
			WithError(err.Error()).
			WithDetail("status_code", r.StatusCode))
	}
	return resp, nil
}
