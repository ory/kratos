// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ory/hydra-client-go/client"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
	kratos "github.com/ory/kratos-client-go"
	"github.com/ory/x/osx"
	"github.com/ory/x/pointerx"
	"github.com/ory/x/urlx"
)

func checkReq(w http.ResponseWriter, err error) bool {
	if err != nil {
		http.Error(w, fmt.Sprintf("%+v", err), 500)
		return false
	}
	return true
}

func main() {
	kratosPublicURL := urlx.ParseOrPanic(osx.GetenvDefault("KRATOS_PUBLIC_URL", "http://localhost:4433"))
	adminURL := urlx.ParseOrPanic(osx.GetenvDefault("HYDRA_ADMIN_URL", "http://localhost:4445"))
	hc := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{Schemes: []string{adminURL.Scheme}, Host: adminURL.Host, BasePath: adminURL.Path})

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`ok`))
	})
	http.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		res, err := hc.Admin.GetLoginRequest(admin.NewGetLoginRequestParams().
			WithLoginChallenge(r.URL.Query().Get("login_challenge")))
		if !checkReq(w, err) {
			return
		}
		if *res.Payload.Skip {
			res, err := hc.Admin.AcceptLoginRequest(admin.NewAcceptLoginRequestParams().
				WithLoginChallenge(r.URL.Query().Get("login_challenge")).
				WithBody(&models.AcceptLoginRequest{Remember: true, RememberFor: 3600,
					Subject: res.Payload.Subject}))
			if !checkReq(w, err) {
				return
			}
			http.Redirect(w, r, *res.Payload.RedirectTo, http.StatusFound)
			return
		}

		challenge := r.URL.Query().Get("login_challenge")
		_, _ = fmt.Fprintf(w, `<html>
<body>
	<form action="/login?login_challenge=%s" method="post">
		<input type="text" name="username" id="username" />
		<input type="checkbox" name="remember" id="remember" value="true"/> Remember me
		<button type="submit" name="action" value="accept" id="accept">login</button>
		<button type="submit" name="action" value="reject" id="reject">reject</button>
	</form>
</body>
</html>`, challenge)
	})

	http.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		if !checkReq(w, r.ParseForm()) {
			return
		}
		if r.Form.Get("action") == "accept" {
			res, err := hc.Admin.AcceptLoginRequest(admin.NewAcceptLoginRequestParams().
				WithLoginChallenge(r.URL.Query().Get("login_challenge")).
				WithBody(&models.AcceptLoginRequest{
					RememberFor: 3600, Remember: r.Form.Get("remember") == "true",
					Subject: pointerx.String(r.Form.Get("username"))}))
			if !checkReq(w, err) {
				return
			}
			http.Redirect(w, r, *res.Payload.RedirectTo, http.StatusFound)
			return
		}
		res, err := hc.Admin.RejectLoginRequest(admin.NewRejectLoginRequestParams().
			WithLoginChallenge(r.URL.Query().Get("login_challenge")).
			WithBody(&models.RejectRequest{Error: "login rejected request"}))
		if !checkReq(w, err) {
			return
		}
		http.Redirect(w, r, *res.Payload.RedirectTo, http.StatusFound)
	})

	http.HandleFunc("GET /consent", func(w http.ResponseWriter, r *http.Request) {
		res, err := hc.Admin.GetConsentRequest(admin.NewGetConsentRequestParams().
			WithConsentChallenge(r.URL.Query().Get("consent_challenge")))
		if !checkReq(w, err) {
			return
		}
		if res.Payload.Skip {
			res, err := hc.Admin.AcceptConsentRequest(admin.NewAcceptConsentRequestParams().
				WithConsentChallenge(r.URL.Query().Get("consent_challenge")).
				WithBody(&models.AcceptConsentRequest{GrantScope: res.Payload.RequestedScope}))
			if !checkReq(w, err) {
				return
			}
			http.Redirect(w, r, *res.Payload.RedirectTo, http.StatusFound)
			return
		}

		checkoxes := ""
		for _, s := range res.Payload.RequestedScope {
			checkoxes += fmt.Sprintf(`<li><input type="checkbox" name="scope" value="%s" id="%s"/>%s</li>`, s, s, s)
		}

		challenge := r.URL.Query().Get("consent_challenge")
		_, _ = fmt.Fprintf(w, `<html>
<body>
	<form action="/consent?consent_challenge=%s" method="post">
		<ul>
		%s
		</ul>
		<input type="text" name="website" id="website" />
		<input type="checkbox" name="remember" id="remember" value="true"/> Remember me
		<button type="submit" name="action" value="accept" id="accept">login</button>
		<button type="submit" name="action" value="reject" id="reject">reject</button>
	</form>
</body>
</html>`, challenge, checkoxes)
	})

	http.HandleFunc("POST /consent", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("action") == "accept" {
			kratosConfig := kratos.NewConfiguration()
			kratosConfig.Servers = kratos.ServerConfigurations{{URL: kratosPublicURL.String()}}
			kratosClient := kratos.NewAPIClient(kratosConfig)
			session, _, err := kratosClient.V0alpha2Api.ToSession(r.Context()).Cookie(r.Header.Get("Cookie")).Execute()
			if err != nil {
				panic(err)
			}
			traitMap, ok := session.Identity.Traits.(map[string]interface{})
			if !ok {
				panic("type assertion failed")
			}
			idToken := map[string]interface{}{}
			// Populate ID token claims with values found in the session's traits
			for _, scope := range r.Form["scope"] {
				if v, ok := traitMap[scope]; ok {
					idToken[scope] = v
				}
			}

			res, err := hc.Admin.AcceptConsentRequest(admin.NewAcceptConsentRequestParams().
				WithConsentChallenge(r.URL.Query().Get("consent_challenge")).
				WithBody(&models.AcceptConsentRequest{
					Session:  &models.ConsentRequestSession{IDToken: idToken},
					Remember: r.Form.Get("remember") == "true", RememberFor: 3600,
					GrantScope: r.Form["scope"]}))
			if !checkReq(w, err) {
				return
			}
			http.Redirect(w, r, *res.Payload.RedirectTo, http.StatusFound)
			return
		}
		res, err := hc.Admin.RejectConsentRequest(admin.NewRejectConsentRequestParams().
			WithConsentChallenge(r.URL.Query().Get("consent_challenge")).
			WithBody(&models.RejectRequest{Error: "consent rejected request"}))
		if !checkReq(w, err) {
			return
		}
		http.Redirect(w, r, *res.Payload.RedirectTo, http.StatusFound)
	})

	addr := ":" + osx.GetenvDefault("PORT", "4746")
	fmt.Printf("Starting web server at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
