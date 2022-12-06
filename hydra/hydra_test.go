// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hydra_test

import (
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/urlx"
)

func TestGetLoginChallengeID(t *testing.T) {
	validChallenge := "https://hydra?login_challenge=b346a452-e8fb-4828-8ef8-a4dbc98dc23a"
	invalidChallenge := "https://hydra?login_challenge=invalid"
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
		name    string
		args    args
		want    uuid.NullUUID
		wantErr bool
	}{
		{
			name: "no login challenge; hydra is not configured",
			args: args{
				conf: defaultConfig,
				r:    &http.Request{URL: urlx.ParseOrPanic("https://hydra")},
			},
			want:    uuid.NullUUID{Valid: false},
			wantErr: false,
		},
		{
			name: "no login challenge; hydra is configured",
			args: args{
				conf: configWithHydra,
				r:    &http.Request{URL: urlx.ParseOrPanic("https://hydra")},
			},
			want:    uuid.NullUUID{Valid: false},
			wantErr: false,
		},
		{
			name: "login_challenge is present; Hydra is not configured",
			args: args{
				conf: defaultConfig,
				r:    &http.Request{URL: urlx.ParseOrPanic(validChallenge)},
			},
			want:    uuid.NullUUID{Valid: false},
			wantErr: true,
		},
		{
			name: "login_challenge is present; hydra is configured",
			args: args{
				conf: configWithHydra,
				r:    &http.Request{URL: urlx.ParseOrPanic(validChallenge)},
			},
			want:    uuid.NullUUID{Valid: true, UUID: uuid.FromStringOrNil("b346a452-e8fb-4828-8ef8-a4dbc98dc23a")},
			wantErr: false,
		},
		{
			name: "login_challenge is invalid; hydra is configured",
			args: args{
				conf: configWithHydra,
				r:    &http.Request{URL: urlx.ParseOrPanic(invalidChallenge)},
			},
			want:    uuid.NullUUID{Valid: false},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hydra.GetLoginChallengeID(tt.args.conf, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLoginChallengeID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLoginChallengeID() = %v, want %v", got, tt.want)
			}
		})
	}
}
