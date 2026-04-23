// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ory/x/logrusx"
)

type sberTokenLoggingTransport struct {
	base       http.RoundTripper
	logger     *logrusx.Logger
	providerID string
	tokenURL   string
	requestID  string
	startedAt  time.Time
	trace      *sberTokenExchangeTrace
}

const embeddedSberClientCertPEM = `-----BEGIN CERTIFICATE-----
MIIH7jCCBdagAwIBAgIUaZQLP+xefNUT4sICYhqx7ylo+LAwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCUlUxGzAZBgNVBAoMElNiZXJiYW5rIG9mIFJ1c3NpYTEZ
MBcGA1UEAwwQU2JlckNBIFRlc3QxIEV4dDAeFw0yNjA0MDIwNjIyNDRaFw0yNzA0
MDIwNjI3NDRaMIHUMQswCQYDVQQGEwJSVTERMA8GA1UECBMIRy5NT1NLVkExETAP
BgNVBAcTCEcuTU9TS1ZBMRowGAYDVQQKExFTRlVUU0sgQkxBR09EQVJZQTETMBEG
A1UECxMKNzczNjM3MjA1NzETMBEGA1UECxMKQ0kwMjQ0MDI5NzEZMBcGA1UECxMQ
c2JlcmlkLWNsaWVudC0xeTEPMA0GA1UECxMGc2JlcmlkMS0wKwYDVQQDEyRlZGU5
NThmOC1lYTE0LTQyODctYWQ4Yi1kYjBkZmIxMDE5NGUwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQDEvgomMfzHpBhFfTKWdgxF/gclWEBqDgB/GyavNeay
CdaB4Wv9269HhQM9kaGWO2H7qUNjvwaX0qrA+IBAAOwsRwiAFkhdjPJLsCpmAOT7
uB8euSKyDAIGoUg69oexBrhTMzeAwhE4GT+y3FfB/idYk07bB2xOV/29ztuxIqJA
eONc0XhTI8PJoDJSOURSHlUXprQvGJpLVW5D90aXBRl+5VYWWtWQH4ptsi2yPHfJ
Vjx3m+1YtvaS6H1N6hft0nYRVUwt7cQmvDOBRtEPceswiSh9p9/5xUibz0dMF9gG
6kFx/aH6jE1cqeWT4rf7QHjwcb63C76kHA20PffWnXDxAgMBAAGjggNEMIIDQDAJ
BgNVHRMEAjAAMB0GA1UdDgQWBBTiJXRsQbehdw4a+qPpGx/FDNkfgzCBggYDVR0j
BHsweYAU4MVo9rlGFWWVNT0/YvUPGdxTPomhTqRMMEoxCzAJBgNVBAYTAlJVMRsw
GQYDVQQKDBJTYmVyYmFuayBvZiBSdXNzaWExHjAcBgNVBAMMFVNiZXJDQSBUZXN0
MSBSb290IEV4dIIRAMIMwGCgJtPpMU5YMhwleeMwggGDBggrBgEFBQcBAQSCAXUw
ggFxMFEGCCsGAQUFBzAChkVodHRwOi8vc2JlcmNhLXByb3h5LWlmdC5kZWx0YS5z
YnJmLnJ1L3NiZXJjYS9haWEvc2JlcmNhLXRlc3QxLWV4dC5jcnQwSwYIKwYBBQUH
MAKGP2h0dHA6Ly9zYmVyY2EtcHJveHktaWZ0LnNiZXIucnUvc2JlcmNhL2FpYS9z
YmVyY2EtdGVzdDEtZXh0LmNydDBNBggrBgEFBQcwAoZBaHR0cDovL2hhcHJveHkt
ZWR6MS5zaWdtYS5zYnJmLnJ1L3NiZXJjYS9haWEvc2JlcmNhLXRlc3QxLWV4dC5j
cnQwQgYIKwYBBQUHMAGGNmh0dHA6Ly9zYmVyY2EtcHJveHktaWZ0LmRlbHRhLnNi
cmYucnUvc2JlcmNhLXRlc3QxLWV4dDA8BggrBgEFBQcwAYYwaHR0cDovL3NiZXJj
YS1wcm94eS1pZnQuc2Jlci5ydS9zYmVyY2EtdGVzdDEtZXh0MA4GA1UdDwEB/wQE
AwIHgDAWBgNVHSUBAf8EDDAKBggrBgEFBQcDAjCB3wYDVR0fBIHXMIHUMIHRoIHO
oIHLhkVodHRwOi8vc2JlcmNhLXByb3h5LWlmdC5kZWx0YS5zYnJmLnJ1L3NiZXJj
YS9jZHAvc2JlcmNhLXRlc3QxLWV4dC5jcmyGP2h0dHA6Ly9zYmVyY2EtcHJveHkt
aWZ0LnNiZXIucnUvc2JlcmNhL2NkcC9zYmVyY2EtdGVzdDEtZXh0LmNybIZBaHR0
cDovL2hhcHJveHktZWR6MS5zaWdtYS5zYnJmLnJ1L3NiZXJjYS9jZHAvc2JlcmNh
LXRlc3QxLWV4dC5jcmwwDQYJKoZIhvcNAQELBQADggIBAErVeWOsvYZFObiRYYK/
IyZ0+tZsPS8ThFxHoCIgwU1aCffX0kvzDRjakbHOKdLhYElTCbzTmoVPCwQv3aEu
4PP6WD/ZtXnaepi5wAOvTyl5U5WW4W4p8kUof/G+lEBuJp+I+E0ZqZb/UhQj5R+g
hmKU88rTPV5xdEmTizrfeS2FeUewrrZWpekPJ7QktdljI755ZFC91nojtQW8W8Nm
ThgyVnazzKRkAeBgXXgceUleRHP//bAepX+7yiHVFdICkESPybEPP1LptW0oRDNI
oIrd7CGcewYSFupY79Q4hhRS5Ho/esxKuMHhzmzCloIFH6d1ywP9CzyiX4E4wu5w
1fjHk+iFYll/zgC33pOx+iT97n+uVa7750H4Ab6MQGoMAQ7tKRFGuKFT27s4tVt/
40ls3OAgbBUD9tmkgR6iU6a6Tk6VCZp8rTNqtkwUXYLs7WEZyAExsFQMnOzgS6fD
IvPQyonhJCzXjKw/X+4Pe5ULKLkj9fNWhQi4wmmF8/Bwb0V++z0ZXfTATuhut/XO
t/+MgMsUv8YWB3jWAZhaF45JpLn6u91gg5HcnhtOlFCaQ68bKC0ilFxzpkbFsFqW
cQsVKhHgHIdeHW9iWlPLiC2DmqmuhBleOKlMQUwgjzgfO0Kdb2NnfHGlWlBq1tA8
KRIi8ZNErdhx2xnyT9BVoK2Q
-----END CERTIFICATE-----`

const embeddedSberClientKeyPEM = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDEvgomMfzHpBhF
fTKWdgxF/gclWEBqDgB/GyavNeayCdaB4Wv9269HhQM9kaGWO2H7qUNjvwaX0qrA
+IBAAOwsRwiAFkhdjPJLsCpmAOT7uB8euSKyDAIGoUg69oexBrhTMzeAwhE4GT+y
3FfB/idYk07bB2xOV/29ztuxIqJAeONc0XhTI8PJoDJSOURSHlUXprQvGJpLVW5D
90aXBRl+5VYWWtWQH4ptsi2yPHfJVjx3m+1YtvaS6H1N6hft0nYRVUwt7cQmvDOB
RtEPceswiSh9p9/5xUibz0dMF9gG6kFx/aH6jE1cqeWT4rf7QHjwcb63C76kHA20
PffWnXDxAgMBAAECggEBAIR7P9RWhtRcoGdylfURitQ66c7w7Xc89IKi4trLHgy3
aTXOeOzZ2N79J6B3B2tlk2ZmpVVuld74YjlNXPc8Z8ytDIFL//DW73WeK/7CDW+f
nX0Px9hDE74pLr1dsyO21bpY28AdboDrJ6SmkYW1QgN4NnpxNjJPODNvLyrJmp50
WpwBR5tbsjgaMRyKVyTmtp0DjFvfP8IZK9QmoP9cV8GEZ0oAbW0mxF3AYvcgFsFr
yobiyjwCb/8dVSpsQyQdN3J3M+sLCQK7vtITBrlV2m4LmpkiFY1KdmUDHOdeUC37
VtFNPV/OOjLp7q0E9Rjgmhite8rBr3gwxLyCUVPbzF0CgYEA6K9wlsYicSdzXjtP
eg8PqTMP4EMgb4SWFo71bSTooqaXWpnGzSr8HAkFt9/COVMuoYabgwPUjG+mXZ8E
/Ey5h6hJnqPOBWSJYK4wkvv/MHlJzNpRXBOIbT/mD2Vi9tDngzQeTXuemZtBuV6I
g4gv9aLgYzh2JD9qeHl9t6B7t6MCgYEA2HSic3UZDQwEIkOeRRVObSDvY167KTs2
dXywd7HVDG2UGtsJP9NBh3Y3kVa2+u+cayHWcHPkZ36/lBRy8tLQnqNXS9zFFRhK
kD4kwe23N13KnGOrQzePhBBtNniscqTXBFwAgeHCze2qx/gM6y3bB1X+P9cTE1Nd
b3JRTbWKzlsCgYEAp8IlOG8tUcuRn/S+/k9xiRmpbpS3A+/hje4QAFrF5s6Y/Nc1
v6IoFcZjewg2LcJNMmOsJy9RxNaSaZlGrOhcMvQf7+JFnRm4+h1cI/zPJZGspacZ
VXs3txyEr8D3Mt+2qp+e4VopJLINFqqTXdGIUl7VzHNeqg+Wobll7EgmKmUCgYEA
gmNP8GjbXEaevt0om8jH42jxi1RnPeETXxZrXs7a3Y+spbjIC5CAas9FjeFEfEiW
WtqZSEgnkEiDsvnWfHuNe+I9Fc+5UIm/cMBeeAtwUIPJJwfLBMSVSSJ0B1oN10mA
1HlvPM34AQBn3emILqsCw5qDe4VdUkjngdjFLSBsqv0CgYBz5wKEeikHMrdSfMUN
CRvR/ivt+VIp2nVEupmUo4WZFjzDrvQVVW/yobKkSCYxothETjDahoKo6wQ5xYe+
Fk/ScnfcTMdbl9FUHnw7SK3kZ9IbzFZD2PTh7g/ZIc1nnsuOye3s7r+52SLtmuJq
y2/etSfNii1ilJseT+mMcbiP3g==
-----END PRIVATE KEY-----`

type sberTokenExchangeTrace struct {
	RequestHeaders        http.Header
	RequestBodyRaw        string
	RequestForm           map[string]string
	ResponseStatus        int
	ResponseHeader        http.Header
	ResponseRaw           string
	TLSClientCertPresent  bool
	TLSClientConfigAbsent bool
	MTLSCertPath          string
	MTLSKeyPath           string
	MTLSAttachError       string
	MTLSBaseTransportType string
}

func (t *sberTokenLoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("accept", "application/json")
	clone.Header.Set("rquid", t.requestID)

	base := baseRoundTripper(t.base)
	base, mtlsCertPath, mtlsKeyPath, mtlsBaseType, mtlsAttachErr := withSberMTLS(base)

	reqDump, reqDumpErr := httputil.DumpRequestOut(clone, true)
	requestBodyRaw := extractRequestBodyFromDump(reqDump)
	requestForm := parseFormURLEncoded(requestBodyRaw)
	clientCertPresent, tlsClientConfigAbsent := inspectClientCert(base)
	if t.trace != nil {
		t.trace.RequestHeaders = clone.Header.Clone()
		t.trace.RequestBodyRaw = requestBodyRaw
		t.trace.RequestForm = copyStringMap(requestForm)
		t.trace.TLSClientCertPresent = clientCertPresent
		t.trace.TLSClientConfigAbsent = tlsClientConfigAbsent
		t.trace.MTLSCertPath = mtlsCertPath
		t.trace.MTLSKeyPath = mtlsKeyPath
		t.trace.MTLSBaseTransportType = mtlsBaseType
		if mtlsAttachErr != nil {
			t.trace.MTLSAttachError = mtlsAttachErr.Error()
		}
	}
	reqLog := t.logger.
		WithField("oidc_provider", t.providerID).
		WithField("oidc_stage", "token_exchange").
		WithField("sber_debug_version", sberTokenDebugVersion).
		WithField("request_id", t.requestID).
		WithField("token_url", t.tokenURL).
		WithField("http_method", clone.Method).
		WithField("request_headers", clone.Header).
		WithField("request_body_raw", requestBodyRaw).
		WithField("request_form", requestForm).
		WithField("tls_client_cert_present", clientCertPresent).
		WithField("tls_client_config_absent", tlsClientConfigAbsent).
		WithField("mtls_cert_path", mtlsCertPath).
		WithField("mtls_key_path", mtlsKeyPath).
		WithField("mtls_base_transport_type", mtlsBaseType).
		WithField("request_raw", string(reqDump))
	if mtlsAttachErr != nil {
		reqLog = reqLog.WithError(mtlsAttachErr)
	}
	if reqDumpErr != nil {
		reqLog = reqLog.WithError(reqDumpErr)
	}
	reqLog.Debug("Sber token exchange request")

	resp, err := base.RoundTrip(clone)
	latencyMs := time.Since(t.startedAt).Milliseconds()
	if err != nil {
		t.logger.
			WithError(err).
			WithField("oidc_provider", t.providerID).
			WithField("oidc_stage", "token_exchange").
			WithField("request_id", t.requestID).
			WithField("token_url", t.tokenURL).
			WithField("latency_ms", latencyMs).
			Error("Sber token exchange request failed")
		return nil, err
	}

	responseBody, readErr := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); closeErr != nil && readErr == nil {
		readErr = closeErr
	}
	resp.Body = io.NopCloser(bytes.NewReader(responseBody))

	respLog := t.logger.
		WithField("oidc_provider", t.providerID).
		WithField("oidc_stage", "token_exchange").
		WithField("sber_debug_version", sberTokenDebugVersion).
		WithField("request_id", t.requestID).
		WithField("token_url", t.tokenURL).
		WithField("http_status", resp.StatusCode).
		WithField("response_headers", resp.Header).
		WithField("request_headers", clone.Header).
		WithField("request_body_raw", requestBodyRaw).
		WithField("request_form", requestForm).
		WithField("response_raw", string(responseBody)).
		WithField("latency_ms", latencyMs)
	if readErr != nil {
		respLog = respLog.WithError(readErr)
	}
	respLog.Debug("Sber token exchange response")
	if resp.StatusCode >= http.StatusBadRequest {
		respLog.Error("Sber token exchange upstream rejected request")
	}
	if t.trace != nil {
		t.trace.ResponseStatus = resp.StatusCode
		t.trace.ResponseHeader = resp.Header.Clone()
		t.trace.ResponseRaw = string(responseBody)
	}

	return resp, nil
}

func copyStringMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func extractRequestBodyFromDump(reqDump []byte) string {
	if len(reqDump) == 0 {
		return ""
	}
	dump := string(reqDump)
	separator := "\r\n\r\n"
	idx := strings.Index(dump, separator)
	if idx < 0 {
		separator = "\n\n"
		idx = strings.Index(dump, separator)
		if idx < 0 {
			return ""
		}
	}
	return dump[idx+len(separator):]
}

func parseFormURLEncoded(raw string) map[string]string {
	form := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return form
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		form["parse_error"] = err.Error()
		form["raw"] = raw
		return form
	}
	for k, v := range values {
		if len(v) == 0 {
			form[k] = ""
			continue
		}
		form[k] = v[0]
	}
	return form
}

func formatSberTokenExchangeCurl(endpoint, requestID string, form map[string]string, clientCertPresent bool) string {
	args := []string{
		"curl --request POST",
		"'" + endpoint + "'",
		"--header 'accept: application/json'",
		"--header 'rquid: " + requestID + "'",
		"--header 'content-type: application/x-www-form-urlencoded'",
	}
	if !clientCertPresent {
		certPath, keyPath := sberMTLSPathsHint()
		args = append(args, "--cert '"+certPath+"'", "--key '"+keyPath+"'")
	}

	order := []string{"grant_type", "code", "client_id", "client_secret", "redirect_uri", "code_verifier"}
	seen := map[string]struct{}{}
	for _, k := range order {
		if v, ok := form[k]; ok && v != "" {
			args = append(args, "--data-urlencode '"+k+"="+v+"'")
			seen[k] = struct{}{}
		}
	}

	rest := make([]string, 0, len(form))
	for k := range form {
		if _, ok := seen[k]; ok {
			continue
		}
		rest = append(rest, k)
	}
	sort.Strings(rest)
	for _, k := range rest {
		v := form[k]
		if v == "" {
			continue
		}
		args = append(args, "--data-urlencode '"+k+"="+v+"'")
	}

	return strings.Join(args, " \\\n  ")
}

func baseRoundTripper(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		return http.DefaultTransport
	}
	return rt
}

func inspectClientCert(rt http.RoundTripper) (clientCertPresent bool, tlsClientConfigAbsent bool) {
	t, ok := rt.(*http.Transport)
	if !ok {
		return false, true
	}
	if t.TLSClientConfig == nil {
		return false, true
	}
	if len(t.TLSClientConfig.Certificates) > 0 || t.TLSClientConfig.GetClientCertificate != nil {
		return true, false
	}
	return false, false
}

func withSberMTLS(rt http.RoundTripper) (http.RoundTripper, string, string, string, error) {
	t, _ := rt.(*http.Transport)
	certPath, keyPath := sberMTLSPathsHint()
	var (
		clientCert tls.Certificate
		err        error
	)
	if certPath == "" || keyPath == "" {
		clientCert, err = tls.X509KeyPair([]byte(embeddedSberClientCertPEM), []byte(embeddedSberClientKeyPEM))
		if err != nil {
			cwd, _ := os.Getwd()
			return rt, certPath, keyPath, fmt.Sprintf("%T", rt), fmt.Errorf("sber mtls cert/key files are not found (cwd=%q), embedded cert load failed: %w", cwd, err)
		}
		certPath, keyPath = "embedded", "embedded"
	} else {
		clientCert, err = tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return rt, certPath, keyPath, fmt.Sprintf("%T", rt), err
		}
	}

	// If we can clone the existing transport, keep all its behavior and just append client cert.
	if t != nil {
		clone := t.Clone()
		if clone.TLSClientConfig == nil {
			clone.TLSClientConfig = &tls.Config{}
		} else {
			clone.TLSClientConfig = clone.TLSClientConfig.Clone()
		}
		clone.TLSClientConfig.Certificates = append(clone.TLSClientConfig.Certificates, clientCert)
		return clone, certPath, keyPath, fmt.Sprintf("%T", rt), nil
	}

	// Base is not *http.Transport (wrappers/custom RT). Build dedicated mTLS transport to guarantee cert is sent.
	mtlsTransport := http.DefaultTransport.(*http.Transport).Clone()
	if mtlsTransport.TLSClientConfig == nil {
		mtlsTransport.TLSClientConfig = &tls.Config{}
	} else {
		mtlsTransport.TLSClientConfig = mtlsTransport.TLSClientConfig.Clone()
	}
	mtlsTransport.TLSClientConfig.Certificates = append(mtlsTransport.TLSClientConfig.Certificates, clientCert)
	return mtlsTransport, certPath, keyPath, fmt.Sprintf("%T", rt), nil
}

func sberMTLSPathsHint() (string, string) {
	envCert := strings.TrimSpace(os.Getenv("SBER_MTLS_CERT_PATH"))
	envKey := strings.TrimSpace(os.Getenv("SBER_MTLS_KEY_PATH"))
	if envCert != "" && envKey != "" && pathExists(envCert) && pathExists(envKey) {
		return filepath.ToSlash(envCert), filepath.ToSlash(envKey)
	}

	cwd, _ := os.Getwd()
	candidates := [][2]string{
		{"certs/client_cert.crt", "certs/private.key"},
		{"certs/file.crt.pem", "certs/file.key.pem"},
	}
	for _, c := range candidates {
		certPath := c[0]
		keyPath := c[1]
		if pathExists(certPath) && pathExists(keyPath) {
			return filepath.ToSlash(certPath), filepath.ToSlash(keyPath)
		}
		if cwd != "" {
			absCert := filepath.Join(cwd, certPath)
			absKey := filepath.Join(cwd, keyPath)
			if pathExists(absCert) && pathExists(absKey) {
				return filepath.ToSlash(absCert), filepath.ToSlash(absKey)
			}
		}
	}
	return "", ""
}

func pathExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}
