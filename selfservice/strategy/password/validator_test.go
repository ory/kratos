package password_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/ory/x/httpx"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/password"
)

func TestDefaultPasswordValidationStrategy(t *testing.T) {
	// Tests are based on:
	// - https://www.troyhunt.com/passwords-evolved-authentication-guidance-for-the-modern-era/
	// - https://www.microsoft.com/en-us/research/wp-content/uploads/2016/06/Microsoft_Password_Guidance-1.pdf

	t.Run("default strategy", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		s := password.NewDefaultPasswordValidatorStrategy(reg)
		for k, tc := range []struct {
			id   string
			pw   string
			pass bool
		}{
			{pw: "", pass: false},
			{pw: "12", pass: false},
			{pw: "1234", pass: false},
			{pw: "123456", pass: false},
			{pw: "12345678", pass: false},
			{pw: "password", pass: false},
			{pw: "1234567890", pass: false},
			{pw: "qwertyui", pass: false},
			{pw: "l3f9toh1uaf81n21", pass: true},
			{pw: "l3f9toh1uaf81n21", id: "l3f9toh1uaf81n21", pass: false},
			{pw: "l3f9toh1", pass: true},
			{pw: "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", pass: true},
			// simple permutation tests
			{id: "hello@example.com", pw: "hello@example.com12345", pass: false},
			{id: "hello@example.com", pw: "123hello@example.com123", pass: false},
			{id: "hello@example.com", pw: "hello@exam", pass: false},
			{id: "hello@example.com", pw: "HELLO@EXAMPLE.COM", pass: false},
			{id: "hello@example.com", pw: "HELLO@example.com", pass: false},
			{pw: "hello@example.com", id: "hello@exam", pass: false},
			{id: "hello@example.com", pw: "h3ll0@example", pass: false},
			{pw: "hello@example.com", id: "hello@exam", pass: false},
			{id: "abcd", pw: "9d3c8a1b", pass: true},
			{id: "a", pw: "kjOkla", pass: true},
			{id: "ab", pw: "0000ab0000", pass: true},
			// longest common substring with long password
			{id: "d4f6090b-5a84", pw: "d4f6090b-5a84-2184-4404-8d1b-8da3eb00ebbe", pass: true},
			{id: "asdflasdflasdf", pw: "asdflasdflpiuhefnciluaksdzuföfhg", pass: true},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				t.Parallel()
				err := s.Validate(context.Background(), tc.id, tc.pw)
				if tc.pass {
					require.NoError(t, err, "err: %+v, id: %s, pw: %s", err, tc.id, tc.pw)
				} else {
					require.Error(t, err, "id: %s, pw: %s", tc.id, tc.pw)
				}
			})
		}

	})

	t.Run("failure cases", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		s := password.NewDefaultPasswordValidatorStrategy(reg)
		fakeClient := NewFakeHTTPClient()
		s.Client = httpx.NewResilientClient(httpx.ResilientClientWithClient(&fakeClient.Client), httpx.ResilientClientWithMaxRetry(1), httpx.ResilientClientWithConnectionTimeout(time.Millisecond), httpx.ResilientClientWithMaxRetryWait(time.Millisecond))

		t.Run("case=should send request to pwnedpasswords.com", func(t *testing.T) {
			conf.MustSet(config.ViperKeyIgnoreNetworkErrors, false)
			require.Error(t, s.Validate(context.Background(), "mohutdesub", "damrumukuh"))
			require.Contains(t, fakeClient.RequestedURLs(), "https://api.pwnedpasswords.com/range/BCBA9")
		})

		t.Run("case=should fail if request fails and ignoreNetworkErrors is not set", func(t *testing.T) {
			conf.MustSet(config.ViperKeyIgnoreNetworkErrors, false)
			fakeClient.RespondWithError("Network request failed")
			require.Error(t, s.Validate(context.Background(), "", "sumdarmetp"))
		})

		t.Run("case=should not fail if request fails and ignoreNetworkErrors is set", func(t *testing.T) {
			conf.MustSet(config.ViperKeyIgnoreNetworkErrors, true)
			fakeClient.RespondWithError("Network request failed")
			require.NoError(t, s.Validate(context.Background(), "", "pepegtawni"))
		})

		t.Run("case=should fail if response has non 200 code and ignoreNetworkErrors is not set", func(t *testing.T) {
			conf.MustSet(config.ViperKeyIgnoreNetworkErrors, false)
			fakeClient.RespondWith(http.StatusForbidden, "")
			require.Error(t, s.Validate(context.Background(), "", "jolhakowef"))
		})

		t.Run("case=should not fail if response has non 200 code code and ignoreNetworkErrors is set", func(t *testing.T) {
			conf.MustSet(config.ViperKeyIgnoreNetworkErrors, true)
			fakeClient.RespondWith(http.StatusInternalServerError, "")
			require.NoError(t, s.Validate(context.Background(), "", "jenuzuhjoj"))
		})
	})

	t.Run("max breaches", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		s := password.NewDefaultPasswordValidatorStrategy(reg)
		fakeClient := NewFakeHTTPClient()
		s.Client = httpx.NewResilientClient(httpx.ResilientClientWithClient(&fakeClient.Client), httpx.ResilientClientWithMaxRetry(1), httpx.ResilientClientWithConnectionTimeout(time.Millisecond))

		conf.MustSet(config.ViperKeyPasswordMaxBreaches, 5)
		for _, tc := range []struct {
			cs   string
			pw   string
			res  string
			pass bool
		}{
			{
				cs:   "contains invalid data",
				pw:   "lufsokpugo",
				res:  "0225BDB8F106B1B4A5DF4C31B80AC695874:2\ninvalid",
				pass: false,
			},
			{
				cs:   "contains invalid hash count",
				pw:   "gimekvizec",
				res:  "0248B3D6077106761CC84F4B9CF680C6D84:text\n1A34C526A9D14832C6ACFEAE90261ED78F8:2",
				pass: false,
			},
			{
				cs:   "is missing hash count",
				pw:   "bofulosasm",
				res:  "1D29CF237A57F6FEA8F29E8D907DCF1EBBA\n026364A8EE59DEDCF9E2DC80B9D7BAB7389:2",
				pass: false,
			},
			{
				cs:   "response contains no matches",
				pw:   "lizrafakha",
				res:  "0D6CF6289C9CA71B47D2167EB7FE89690E7:57",
				pass: true,
			},
			{
				cs:   "contains less than maxBreachesThreshold",
				pw:   "tafpabdopa",
				res:  fmt.Sprintf("280915F3B572F94217D86F1D63BED53F66A:%d\n0F76A7D21E7C3E653E98236897AD7888937:%d", conf.PasswordPolicyConfig().MaxBreaches, conf.PasswordPolicyConfig().MaxBreaches+1),
				pass: true,
			},
			{
				cs:   "contains more than maxBreachesThreshold",
				pw:   "hicudsumla",
				res:  fmt.Sprintf("5656812AA72561AAA6663E486A46D5711BE:%d", conf.PasswordPolicyConfig().MaxBreaches+1),
				pass: false,
			},
		} {
			fakeClient.RespondWith(http.StatusOK, tc.res)
			format := "case=should not fail if response %s"
			if !tc.pass {
				format = "case=should fail if response %s"
			}
			t.Run(fmt.Sprintf(format, tc.cs), func(t *testing.T) {
				err := s.Validate(context.Background(), "", tc.pw)
				if tc.pass {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			})
		}
	})
}

type fakeHttpClient struct {
	http.Client

	requestedURLs []string
	responder     func(*http.Request) (*http.Response, error)
}

func NewFakeHTTPClient() *fakeHttpClient {
	client := fakeHttpClient{
		responder: func(*http.Request) (*http.Response, error) {
			return nil, errors.New("No responder defined in fake HTTP client")
		},
	}
	client.Client = http.Client{
		Transport: &fakeRoundTripper{&client},
		Timeout:   time.Second,
	}
	return &client
}

func (c *fakeHttpClient) RespondWith(status int, body string) {
	c.responder = func(request *http.Request) (*http.Response, error) {
		buffer := bytes.NewBufferString(body)
		return &http.Response{
			StatusCode:    status,
			Body:          ioutil.NopCloser(buffer),
			ContentLength: int64(buffer.Len()),
			Request:       request,
		}, nil
	}
}

func (c *fakeHttpClient) RespondWithError(err string) {
	c.responder = func(*http.Request) (*http.Response, error) {
		return nil, errors.New(err)
	}
}

func (c *fakeHttpClient) Reset() {
	c.requestedURLs = nil
}

func (c *fakeHttpClient) RequestedURLs() []string {
	return c.requestedURLs
}

func (c *fakeHttpClient) handle(request *http.Request) (*http.Response, error) {
	c.requestedURLs = append(c.requestedURLs, request.URL.String())
	if request.Body != nil {
		request.Body.Close()
	}
	return c.responder(request)
}

type fakeRoundTripper struct {
	client *fakeHttpClient
}

func (rt *fakeRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return rt.client.handle(request)
}
