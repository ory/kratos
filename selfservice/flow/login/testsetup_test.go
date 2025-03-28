// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var returnToServer *httptest.Server

func TestMain(m *testing.M) {
	returnToServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	os.Exit(m.Run())
}
