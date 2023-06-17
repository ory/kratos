// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hydra_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

func requestFromChallenge(s string) *http.Request {
	return &http.Request{URL: urlx.ParseOrPanic("https://hydra?login_challenge=" + s)}
}

func TestGetLoginChallengeID(t *testing.T) {
	uuidChallenge := "b346a452-e8fb-4828-8ef8-a4dbc98dc23a"
	blobChallenge := "1337deadbeefcafe"
	defaultConfig := config.MustNew(t, logrusx.New("", ""), os.Stderr, configx.SkipValidation())
	configWithHydra := config.MustNew(
		t,
		logrusx.New("", ""),
		os.Stderr,
		configx.SkipValidation(),
		configx.WithValues(map[string]interface{}{
			config.ViperKeyOAuth2ProviderURL: "https://hydra",
		}),
	)

	type args struct {
		conf *config.Config
		r    *http.Request
	}
	tests := []struct {
		name      string
		args      args
		want      string
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name: "no login challenge; hydra is not configured",
			args: args{
				conf: defaultConfig,
				r:    &http.Request{URL: urlx.ParseOrPanic("https://hydra")},
			},
			want:      "",
			assertErr: assert.NoError,
		},
		{
			name: "no login challenge; hydra is configured",
			args: args{
				conf: configWithHydra,
				r:    &http.Request{URL: urlx.ParseOrPanic("https://hydra")},
			},
			want:      "",
			assertErr: assert.NoError,
		},
		{
			name: "empty login challenge; hydra is configured",
			args: args{
				conf: configWithHydra,
				r:    requestFromChallenge(""),
			},
			want:      "",
			assertErr: assert.Error,
		},
		{
			name: "login_challenge is present; Hydra is not configured",
			args: args{
				conf: defaultConfig,
				r:    requestFromChallenge(uuidChallenge),
			},
			want:      "",
			assertErr: assert.Error,
		},
		{
			name: "login_challenge is present; hydra is configured",
			args: args{
				conf: configWithHydra,
				r:    requestFromChallenge(uuidChallenge),
			},
			want:      uuidChallenge,
			assertErr: assert.NoError,
		},
		{
			name: "login_challenge is present & non-uuid; hydra is configured",
			args: args{
				conf: configWithHydra,
				r:    requestFromChallenge(blobChallenge),
			},
			want:      blobChallenge,
			assertErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hydra.GetLoginChallengeID(tt.args.conf, tt.args.r)
			tt.assertErr(t, err)
			assert.Equal(t, sqlxx.NullString(tt.want), got)
		})
	}
}
