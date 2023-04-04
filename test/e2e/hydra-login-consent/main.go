// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	client "github.com/ory/hydra-client-go/v2"

	"github.com/ory/x/osx"
	"github.com/ory/x/pointerx"
	"github.com/ory/x/urlx"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func checkReq(w http.ResponseWriter, err error) bool {
	if err != nil {
		http.Error(w, fmt.Sprintf("%+v", err), 500)
		return false
	}
	return true
}

func main() {
	router := httprouter.New()

	adminURL := urlx.ParseOrPanic(osx.GetenvDefault("HYDRA_ADMIN_URL", "http://localhost:4445"))
	cfg := client.NewConfiguration()
	cfg.Servers = client.ServerConfigurations{
		{URL: adminURL.String()},
	}
	hc := client.NewAPIClient(cfg)

	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte(`ok`))
	})
	router.GET("/login", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		res, _, err := hc.OAuth2Api.GetOAuth2LoginRequest(r.Context()).LoginChallenge(r.URL.Query().Get("login_challenge")).Execute()
		if !checkReq(w, err) {
			return
		}

		if res.Skip {
			res, _, err := hc.OAuth2Api.AcceptOAuth2LoginRequest(r.Context()).
				LoginChallenge(r.URL.Query().Get("login_challenge")).
				AcceptOAuth2LoginRequest(client.AcceptOAuth2LoginRequest{
					Remember:    pointerx.Bool(true),
					RememberFor: pointerx.Int64(3600),
					Subject:     res.Subject,
				}).Execute()
			if !checkReq(w, err) {
				return
			}
			http.Redirect(w, r, res.RedirectTo, http.StatusFound)
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

	router.POST("/login", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		check(r.ParseForm())
		remember := pointerx.Bool(r.Form.Get("remember") == "true")
		if r.Form.Get("action") == "accept" {
			res, _, err := hc.OAuth2Api.AcceptOAuth2LoginRequest(r.Context()).
				LoginChallenge(r.URL.Query().Get("login_challenge")).
				AcceptOAuth2LoginRequest(client.AcceptOAuth2LoginRequest{
					RememberFor: pointerx.Int64(3600),
					Remember:    remember,
					Subject:     r.Form.Get("username"),
				}).Execute()

			if !checkReq(w, err) {
				return
			}
			http.Redirect(w, r, res.RedirectTo, http.StatusFound)
			return
		}
		res, _, err := hc.OAuth2Api.RejectOAuth2LoginRequest(r.Context()).LoginChallenge(r.URL.Query().Get("login_challenge")).
			RejectOAuth2Request(client.RejectOAuth2Request{Error: pointerx.String("login rejected request")}).Execute()
		if !checkReq(w, err) {
			return
		}
		http.Redirect(w, r, res.RedirectTo, http.StatusFound)
	})

	router.GET("/consent", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		res, _, err := hc.OAuth2Api.GetOAuth2ConsentRequest(r.Context()).ConsentChallenge(r.URL.Query().
			Get("consent_challenge")).Execute()
		if !checkReq(w, err) {
			return
		}

		if *res.Skip {
			res, _, err := hc.OAuth2Api.AcceptOAuth2ConsentRequest(r.Context()).
				ConsentChallenge(r.URL.Query().Get("consent_challenge")).
				AcceptOAuth2ConsentRequest(client.AcceptOAuth2ConsentRequest{
					GrantScope: res.RequestedScope,
				}).Execute()
			if !checkReq(w, err) {
				return
			}
			http.Redirect(w, r, res.RedirectTo, http.StatusFound)
			return
		}

		checkoxes := ""
		for _, s := range res.RequestedScope {
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

	router.POST("/consent", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_ = r.ParseForm()
		remember := pointerx.Bool(r.Form.Get("remember") == "true")
		if r.Form.Get("action") == "accept" {
			res, _, err := hc.OAuth2Api.AcceptOAuth2ConsentRequest(r.Context()).
				ConsentChallenge(r.URL.Query().Get("consent_challenge")).
				AcceptOAuth2ConsentRequest(client.AcceptOAuth2ConsentRequest{
					Session: &client.AcceptOAuth2ConsentRequestSession{
						IdToken: map[string]interface{}{
							"website": r.Form.Get("website"),
						},
					},
					RememberFor: pointerx.Int64(3600),
					Remember:    remember,
					GrantScope:  r.Form["scope"]},
				).Execute()
			if !checkReq(w, err) {
				return
			}
			http.Redirect(w, r, res.RedirectTo, http.StatusFound)
			return
		}
		res, _, err := hc.OAuth2Api.RejectOAuth2ConsentRequest(r.Context()).
			ConsentChallenge(r.URL.Query().Get("consent_challenge")).
			RejectOAuth2Request(client.RejectOAuth2Request{Error: pointerx.String("consent rejected request")}).Execute()
		if !checkReq(w, err) {
			return
		}
		http.Redirect(w, r, res.RedirectTo, http.StatusFound)
	})

	addr := ":" + osx.GetenvDefault("PORT", "4446")
	//#nosec G112
	server := &http.Server{Addr: addr, Handler: router}
	fmt.Printf("Starting web server at %s\n", addr)
	check(server.ListenAndServe())
}
