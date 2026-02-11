// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1" //#nosec G505 -- compatibility for imported passwords
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/text"
	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/httpx"
)

func TestDefaultPasswordValidationStrategy(t *testing.T) {
	t.Parallel()

	// Tests are based on:
	// - https://www.troyhunt.com/passwords-evolved-authentication-guidance-for-the-modern-era/
	// - https://www.microsoft.com/en-us/research/wp-content/uploads/2016/06/Microsoft_Password_Guidance-1.pdf

	t.Run("default strategy", func(t *testing.T) {
		t.Parallel()

		_, reg := pkg.NewFastRegistryWithMocks(t)
		s, err := password.NewDefaultPasswordValidatorStrategy(reg)
		require.NoError(t, err)

		for k, tc := range []struct {
			id, pw      string
			expectedErr error
		}{
			{pw: "", expectedErr: text.NewErrorValidationPasswordMinLength(0, 0)},
			{pw: "12", expectedErr: text.NewErrorValidationPasswordMinLength(0, 0)},
			{pw: "1234", expectedErr: text.NewErrorValidationPasswordMinLength(0, 0)},
			{pw: "123456", expectedErr: text.NewErrorValidationPasswordMinLength(0, 0)},
			{pw: "12345678", expectedErr: text.NewErrorValidationPasswordTooManyBreaches(0)},
			{pw: "password", expectedErr: text.NewErrorValidationPasswordTooManyBreaches(0)},
			{pw: "1234567890", expectedErr: text.NewErrorValidationPasswordTooManyBreaches(0)},
			{pw: "qwertyui", expectedErr: text.NewErrorValidationPasswordTooManyBreaches(0)},
			{pw: "l3f9to", expectedErr: text.NewErrorValidationPasswordMinLength(0, 0)},
			{pw: "l3f9toh1uaf81n21"},
			{pw: "l3f9toh1uaf81n21", id: "l3f9toh1uaf81n21", expectedErr: text.NewErrorValidationPasswordIdentifierTooSimilar()},
			{pw: "l3f9toh1"},
			{pw: "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"},
			// simple permutation tests
			{id: "hello@example.com", pw: "hello@example.com12345", expectedErr: text.NewErrorValidationPasswordIdentifierTooSimilar()},
			{id: "hello@example.com", pw: "123hello@example.com123", expectedErr: text.NewErrorValidationPasswordIdentifierTooSimilar()},
			{id: "hello@example.com", pw: "hello@exam", expectedErr: text.NewErrorValidationPasswordIdentifierTooSimilar()},
			{id: "hello@example.com", pw: "HELLO@EXAMPLE.COM", expectedErr: text.NewErrorValidationPasswordIdentifierTooSimilar()},
			{id: "hello@example.com", pw: "HELLO@example.com", expectedErr: text.NewErrorValidationPasswordIdentifierTooSimilar()},
			{pw: "hello@example.com", id: "hello@exam", expectedErr: text.NewErrorValidationPasswordIdentifierTooSimilar()},
			{id: "hello@example.com", pw: "h3ll0@example", expectedErr: text.NewErrorValidationPasswordIdentifierTooSimilar()},
			{pw: "hello@example.com", id: "hello@exam", expectedErr: text.NewErrorValidationPasswordIdentifierTooSimilar()},
			{id: "abcd", pw: "9d3c8a1b"},
			{id: "a", pw: "kjOklafe"},
			{id: "ab", pw: "0000ab0000123"},
			// longest common substring with long password
			{id: "d4f6090b-5a84", pw: "d4f6090b-5a84-2184-4404-8d1b-8da3eb00ebbe"},
			{id: "asdflasdflasdf", pw: "asdflasdflpiuhefnciluaksdzuföfhg"},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				t.Parallel()

				err := s.Validate(t.Context(), tc.id, tc.pw)
				if tc.expectedErr == nil {
					require.NoErrorf(t, err, "id: %s, pw: %s", tc.id, tc.pw)
				} else {
					require.ErrorIsf(t, err, tc.expectedErr, "id: %s, pw: %s", tc.id, tc.pw)
				}
			})
		}
	})

	t.Run("failure cases", func(t *testing.T) {
		t.Parallel()

		_, reg := pkg.NewFastRegistryWithMocks(t)
		s, err := password.NewDefaultPasswordValidatorStrategy(reg)
		require.NoError(t, err)

		fakeClient := NewFakeHTTPClient()
		s.Client = httpx.NewResilientClient(
			httpx.ResilientClientWithMaxRetry(1),
			httpx.ResilientClientWithConnectionTimeout(time.Millisecond),
			httpx.ResilientClientWithMaxRetryWait(time.Millisecond))
		s.Client.HTTPClient = &fakeClient.Client

		t.Run("case=should send request to pwnedpasswords.com", func(t *testing.T) {
			ctx := contextx.WithConfigValue(t.Context(), config.ViperKeyIgnoreNetworkErrors, false)
			require.Error(t, s.Validate(ctx, "mohutdesub", "damrumukuh"))
			require.Contains(t, fakeClient.RequestedURLs(), "https://api.pwnedpasswords.com/range/BCBA9")
		})

		t.Run("case=should fail if request fails and ignoreNetworkErrors is not set", func(t *testing.T) {
			ctx := contextx.WithConfigValue(t.Context(), config.ViperKeyIgnoreNetworkErrors, false)
			fakeClient.RespondWithError("Network request failed")
			require.Error(t, s.Validate(ctx, "", "sumdarmetp"))
		})

		t.Run("case=should not fail if request fails and ignoreNetworkErrors is set", func(t *testing.T) {
			ctx := contextx.WithConfigValue(t.Context(), config.ViperKeyIgnoreNetworkErrors, true)
			fakeClient.RespondWithError("Network request failed")
			require.NoError(t, s.Validate(ctx, "", "pepegtawni"))
		})

		t.Run("case=should fail if response has non 200 code and ignoreNetworkErrors is not set", func(t *testing.T) {
			ctx := contextx.WithConfigValue(t.Context(), config.ViperKeyIgnoreNetworkErrors, false)
			fakeClient.RespondWith(http.StatusForbidden, "")
			require.Error(t, s.Validate(ctx, "", "jolhakowef"))
		})

		t.Run("case=should not fail if response has non 200 code code and ignoreNetworkErrors is set", func(t *testing.T) {
			ctx := contextx.WithConfigValue(t.Context(), config.ViperKeyIgnoreNetworkErrors, true)
			fakeClient.RespondWith(http.StatusInternalServerError, "")
			require.NoError(t, s.Validate(ctx, "", "jenuzuhjoj"))
		})
	})

	t.Run("max breaches", func(t *testing.T) {
		t.Parallel()

		const maxBreaches = 5
		_, reg := pkg.NewFastRegistryWithMocks(t, configx.WithValue(config.ViperKeyPasswordMaxBreaches, maxBreaches))
		s, err := password.NewDefaultPasswordValidatorStrategy(reg)
		require.NoError(t, err)

		hibpResp := make(chan string, 1)
		fakeClient := NewFakeHTTPClient()
		fakeClient.responder = func(req *http.Request) (*http.Response, error) {
			buffer := bytes.NewBufferString(<-hibpResp)
			return &http.Response{
				StatusCode:    http.StatusOK,
				Body:          io.NopCloser(buffer),
				ContentLength: int64(buffer.Len()),
				Request:       req,
			}, nil
		}
		s.Client = httpx.NewResilientClient(httpx.ResilientClientWithMaxRetry(1), httpx.ResilientClientWithConnectionTimeout(time.Millisecond))
		s.Client.HTTPClient = &fakeClient.Client

		hashPw := func(t *testing.T, pw string) string {
			//#nosec G401 -- sha1 is used for k-anonymity
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
				expectErr: herodot.ErrUpstreamError,
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
						maxBreaches,
						hashPw(t, randomPassword(t)),
						maxBreaches+1,
					)
				},
			},
			{
				name: "contains less than maxBreachesThreshold with a leading comma",
				res: func(t *testing.T, hash string) string {
					return fmt.Sprintf(
						"%s:%d\n%s:0,%d",
						hash,
						maxBreaches,
						hashPw(t, randomPassword(t)),
						maxBreaches+1,
					)
				},
			},
			{
				name: "contains more than maxBreachesThreshold",
				res: func(t *testing.T, hash string) string {
					return fmt.Sprintf("%s:%d", hash, maxBreaches+1)
				},
				expectErr: text.NewErrorValidationPasswordTooManyBreaches(
					int64(maxBreaches) + 1, // #nosec G115
				),
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
	t.Parallel()

	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	t.Cleanup(testServer.Close)
	testServerURL, err := url.Parse(testServer.URL)
	require.NoError(t, err)

	_, reg := pkg.NewFastRegistryWithMocks(t, configx.WithValue(config.ViperKeyPasswordHaveIBeenPwnedHost, testServerURL.Host))
	s, err := password.NewDefaultPasswordValidatorStrategy(reg)
	require.NoError(t, err)

	fakeClient := NewFakeHTTPClient()
	s.Client = httpx.NewResilientClient(httpx.ResilientClientWithMaxRetry(1), httpx.ResilientClientWithConnectionTimeout(time.Millisecond))
	s.Client.HTTPClient = &fakeClient.Client

	testServerExpectedCallURL := fmt.Sprintf("https://%s/range/BCBA9", testServerURL.Host)

	t.Run("case=should send request to test server", func(t *testing.T) {
		ctx := contextx.WithConfigValue(t.Context(), config.ViperKeyIgnoreNetworkErrors, false)
		require.Error(t, s.Validate(ctx, "mohutdesub", "damrumukuh"))
		require.Contains(t, fakeClient.RequestedURLs(), testServerExpectedCallURL)
	})
}

func TestDisableHaveIBeenPwnedValidationHost(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t, configx.WithValue(config.ViperKeyPasswordHaveIBeenPwnedEnabled, false))
	s, err := password.NewDefaultPasswordValidatorStrategy(reg)
	require.NoError(t, err)

	fakeClient := NewFakeHTTPClient()
	s.Client = httpx.NewResilientClient(httpx.ResilientClientWithMaxRetry(1), httpx.ResilientClientWithConnectionTimeout(time.Millisecond))
	s.Client.HTTPClient = &fakeClient.Client

	t.Run("case=should not send request to test server", func(t *testing.T) {
		require.NoError(t, s.Validate(t.Context(), "mohutdesub", "damrumukuh"))
		require.Empty(t, fakeClient.RequestedURLs())
	})
}

func TestChangeMinPasswordLength(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t, configx.WithValue(config.ViperKeyPasswordMinLength, 10))
	s, err := password.NewDefaultPasswordValidatorStrategy(reg)
	require.NoError(t, err)

	t.Run("case=should not fail if password is longer than min length", func(t *testing.T) {
		require.NoError(t, s.Validate(t.Context(), "", "kuobahcaas"))
	})

	t.Run("case=should fail if password is shorter than min length", func(t *testing.T) {
		require.Error(t, s.Validate(t.Context(), "", "rfqyfjied"))
	})
}

func TestChangeIdentifierSimilarityCheckEnabled(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t)
	s, err := password.NewDefaultPasswordValidatorStrategy(reg)
	require.NoError(t, err)

	t.Run("case=should not fail if password is similar to identifier", func(t *testing.T) {
		ctx := contextx.WithConfigValue(t.Context(), config.ViperKeyPasswordIdentifierSimilarityCheckEnabled, false)
		require.NoError(t, s.Validate(ctx, "bosqwfaxee", "bosqwfaxee"))
	})

	t.Run("case=should fail if password is similar to identifier", func(t *testing.T) {
		ctx := contextx.WithConfigValue(t.Context(), config.ViperKeyPasswordIdentifierSimilarityCheckEnabled, true)
		require.Error(t, s.Validate(ctx, "bosqwfaxee", "bosqwfaxee"))
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
			Body:          io.NopCloser(buffer),
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
		_ = request.Body.Close()
	}
	return c.responder(request)
}

type fakeRoundTripper struct {
	client *fakeHttpClient
}

func (rt *fakeRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return rt.client.handle(request)
}
