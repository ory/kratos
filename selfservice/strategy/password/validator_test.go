package password_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ory/herodot"

	"github.com/stretchr/testify/require"

	"github.com/ory/x/httpx"

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
		s, _ := password.NewDefaultPasswordValidatorStrategy(reg)
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
			{pw: "l3f9to", pass: false},
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
			{id: "a", pw: "kjOklafe", pass: true},
			{id: "ab", pw: "0000ab0000", pass: true},
			// longest common substring with long password
			{id: "d4f6090b-5a84", pw: "d4f6090b-5a84-2184-4404-8d1b-8da3eb00ebbe", pass: true},
			{id: "asdflasdflasdf", pw: "asdflasdflpiuhefnciluaksdzuföfhg", pass: true},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				c := tc
				t.Parallel()

				err := s.Validate(context.Background(), c.id, c.pw)
				if c.pass {
					require.NoError(t, err, "err: %+v, id: %s, pw: %s", err, c.id, c.pw)
				} else {
					require.Error(t, err, "id: %s, pw: %s", c.id, c.pw)
				}
			})
		}

	})

	t.Run("failure cases", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		s, _ := password.NewDefaultPasswordValidatorStrategy(reg)
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
		s, err := password.NewDefaultPasswordValidatorStrategy(reg)
		require.NoError(t, err)

		hibpResp := make(chan string, 1)
		fakeClient := NewFakeHTTPClient()
		fakeClient.responder = func(req *http.Request) (*http.Response, error) {
			buffer := bytes.NewBufferString(<-hibpResp)
			return &http.Response{
				StatusCode:    http.StatusOK,
				Body:          ioutil.NopCloser(buffer),
				ContentLength: int64(buffer.Len()),
				Request:       req,
			}, nil
		}
		s.Client = httpx.NewResilientClient(httpx.ResilientClientWithClient(&fakeClient.Client), httpx.ResilientClientWithMaxRetry(1), httpx.ResilientClientWithConnectionTimeout(time.Millisecond))

		var hashPw = func(t *testing.T, pw string) string {
			/* #nosec G401 sha1 is used for k-anonymity */
			h := sha1.New()
			_, err := h.Write([]byte(pw))
			require.NoError(t, err)
			hpw := h.Sum(nil)
			return fmt.Sprintf("%X", hpw)[5:]
		}
		randomPassword := func(t *testing.T) string {
			pw := make([]byte, 10)
			_, err := rand.Read(pw)
			require.NoError(t, err)
			return fmt.Sprintf("%x", pw)
		}

		conf.MustSet(config.ViperKeyPasswordMaxBreaches, 5)
		for _, tc := range []struct {
			name      string
			res       func(t *testing.T, hash string) string
			expectErr error
		}{
			{
				name: "contains invalid data which is ignored",
				res: func(t *testing.T, hash string) string {
					return fmt.Sprintf("%s:2\ninvalid", hash)
				},
			},
			{
				name: "is missing a colon",
				res: func(t *testing.T, hash string) string {
					return hash
				},
			},
			{
				name: "contains invalid hash count",
				res: func(t *testing.T, hash string) string {
					return fmt.Sprintf("%s:text\n%s:2", hashPw(t, randomPassword(t)), hash)
				},
				expectErr: herodot.ErrInternalServerError,
			},
			{
				name: "is missing hash count",
				res: func(t *testing.T, hash string) string {
					return fmt.Sprintf("%s\n%s:2", hash, hashPw(t, randomPassword(t)))
				},
			},
			{
				name: "response contains no matches",
				res: func(t *testing.T, hash string) string {
					return fmt.Sprintf("%s:57", hashPw(t, randomPassword(t)))
				},
			},
			{
				name: "contains less than maxBreachesThreshold",
				res: func(t *testing.T, hash string) string {
					return fmt.Sprintf(
						"%s:%d\n%s:%d",
						hash,
						conf.PasswordPolicyConfig().MaxBreaches,
						hashPw(t, randomPassword(t)),
						conf.PasswordPolicyConfig().MaxBreaches+1,
					)
				},
			},
			{
				name: "contains more than maxBreachesThreshold",
				res: func(t *testing.T, hash string) string {
					return fmt.Sprintf("%s:%d", hash, conf.PasswordPolicyConfig().MaxBreaches+1)
				},
				expectErr: password.ErrTooManyBreaches,
			},
		} {
			t.Run(fmt.Sprintf("case=%s/expected err=%s", tc.name, tc.expectErr), func(t *testing.T) {
				pw := randomPassword(t)
				hash := hashPw(t, pw)
				hibpResp <- tc.res(t, hash)

				err := s.Validate(context.Background(), "", pw)
				assert.ErrorIs(t, err, tc.expectErr)
			})

			// verify the fetch was done, i.e. channel is empty
			select {
			case r := <-hibpResp:
				t.Logf("expected the validate step to fetch the response, but I still got %s", r)
				t.FailNow()
			default:
				// continue
			}
		}
	})
}

func TestChangeHaveIBeenPwnedValidationHost(t *testing.T) {
	testServer := httptest.NewUnstartedServer(&fakeValidatorAPI{})
	defer testServer.Close()
	testServer.StartTLS()
	testServerURL, _ := url.Parse(testServer.URL)
	conf, reg := internal.NewFastRegistryWithMocks(t)
	s, _ := password.NewDefaultPasswordValidatorStrategy(reg)
	conf.MustSet(config.ViperKeyPasswordHaveIBeenPwnedHost, testServerURL.Host)

	fakeClient := NewFakeHTTPClient()
	s.Client = httpx.NewResilientClient(httpx.ResilientClientWithClient(&fakeClient.Client), httpx.ResilientClientWithMaxRetry(1), httpx.ResilientClientWithConnectionTimeout(time.Millisecond))

	testServerExpectedCallURL := fmt.Sprintf("https://%s/range/BCBA9", testServerURL.Host)

	t.Run("case=should send request to test server", func(t *testing.T) {
		conf.MustSet(config.ViperKeyIgnoreNetworkErrors, false)
		require.Error(t, s.Validate(context.Background(), "mohutdesub", "damrumukuh"))
		require.Contains(t, fakeClient.RequestedURLs(), testServerExpectedCallURL)
	})
}

func TestDisableHaveIBeenPwnedValidationHost(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	s, _ := password.NewDefaultPasswordValidatorStrategy(reg)
	conf.MustSet(config.ViperKeyPasswordHaveIBeenPwnedEnabled, false)

	fakeClient := NewFakeHTTPClient()
	s.Client = httpx.NewResilientClient(httpx.ResilientClientWithClient(&fakeClient.Client), httpx.ResilientClientWithMaxRetry(1), httpx.ResilientClientWithConnectionTimeout(time.Millisecond))

	t.Run("case=should not send request to test server", func(t *testing.T) {
		require.NoError(t, s.Validate(context.Background(), "mohutdesub", "damrumukuh"))
		require.Empty(t, fakeClient.RequestedURLs())
	})
}

func TestChangeMinPasswordLength(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	s, _ := password.NewDefaultPasswordValidatorStrategy(reg)
	conf.MustSet(config.ViperKeyPasswordMinLength, 10)

	t.Run("case=should not fail if password is longer than min length", func(t *testing.T) {
		require.NoError(t, s.Validate(context.Background(), "", "kuobahcaas"))
	})

	t.Run("case=should fail if password is shorter than min length", func(t *testing.T) {
		require.Error(t, s.Validate(context.Background(), "", "rfqyfjied"))
	})
}

func TestChangeIdentifierSimilarityCheckEnabled(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	s, _ := password.NewDefaultPasswordValidatorStrategy(reg)

	t.Run("case=should not fail if password is similar to identifier", func(t *testing.T) {
		conf.MustSet(config.ViperKeyPasswordIdentifierSimilarityCheckEnabled, false)
		require.NoError(t, s.Validate(context.Background(), "bosqwfaxee", "bosqwfaxee"))
	})

	t.Run("case=should fail if password is similar to identifier", func(t *testing.T) {
		conf.MustSet(config.ViperKeyPasswordIdentifierSimilarityCheckEnabled, true)
		require.Error(t, s.Validate(context.Background(), "bosqwfaxee", "bosqwfaxee"))
	})
}

type fakeValidatorAPI struct{}

func (api *fakeValidatorAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
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
