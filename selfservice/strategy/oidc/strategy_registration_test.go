// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"net/http"
	"reflect"
	"testing"

	"golang.org/x/oauth2"

	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/x/decoderx"
)

func TestStrategy_processRegistration(t *testing.T) {
	type fields struct {
		d         dependencies
		validator *schema.Validator
		dec       *decoderx.HTTP
	}
	type args struct {
		w         http.ResponseWriter
		r         *http.Request
		a         *registration.Flow
		token     *oauth2.Token
		claims    *Claims
		provider  Provider
		container *authCodeContainer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *login.Flow
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Strategy{
				d:         tt.fields.d,
				validator: tt.fields.validator,
				dec:       tt.fields.dec,
			}
			got, err := s.processRegistration(tt.args.w, tt.args.r, tt.args.a, tt.args.token, tt.args.claims, tt.args.provider, tt.args.container)
			if (err != nil) != tt.wantErr {
				t.Errorf("Strategy.processRegistration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Strategy.processRegistration() = %v, want %v", got, tt.want)
			}
		})
	}
}
