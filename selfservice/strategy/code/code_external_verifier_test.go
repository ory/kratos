package code_test

import (
	"bytes"
	"context"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/httpx"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"testing"
)

type MockDependencies struct {
	mock.Mock
	t              *testing.T
	mockHTTPClient *retryablehttp.Client
}

func NewMockDependencies(t *testing.T, mockHTTPClient *retryablehttp.Client) *MockDependencies {
	return &MockDependencies{t: t, mockHTTPClient: mockHTTPClient}
}

func (m *MockDependencies) Config() *config.Config {
	return config.MustNew(
		m.t,
		logrusx.New("kratos", "test"),
		os.Stderr,
		&contextx.Default{},
		configx.WithConfigFiles("../../../test/e2e/profiles/code/.kratos.yml"),
		configx.SkipValidation(),
	)
}

func (m *MockDependencies) Logger() *logrusx.Logger {
	return logrusx.New("kratos", "test")
}

func (m *MockDependencies) Audit() *logrusx.Logger {
	return logrusx.New("kratos", "test")
}

func (m *MockDependencies) Tracer(ctx context.Context) *otelx.Tracer {
	return otelx.NewNoop(nil, nil)
}

func (m *MockDependencies) HTTPClient(ctx context.Context, options ...httpx.ResilientOptions) *retryablehttp.Client {
	return m.mockHTTPClient
}

func (m *MockDependencies) JsonnetVM(ctx context.Context) (jsonnetsecure.VM, error) {
	return jsonnetsecure.NewTestProvider(m.t).JsonnetVM(ctx)
}

type MockSMSTemplate struct {
	mock.Mock
	marshalJson string
}

func (m *MockSMSTemplate) MarshalJSON() ([]byte, error) {
	return []byte(m.marshalJson), nil
}

func (m *MockSMSTemplate) SMSBody(ctx context.Context) (string, error) {
	return "sms body", nil
}

func (m *MockSMSTemplate) TemplateType() template.TemplateType {
	return "sms"
}

func (m *MockSMSTemplate) PhoneNumber() (string, error) {
	return "1234567890", nil
}

func TestVerificationStart(t *testing.T) {
	ctx := context.Background()

	mockHTTPClient := new(retryablehttp.Client)
	mockHTTPClient.CheckRetry = driver.NoRetryOnRateLimitPolicy
	mockDeps := NewMockDependencies(t, mockHTTPClient)

	mockSMSTemplate := new(MockSMSTemplate)
	mockSMSTemplate.marshalJson = `{"To":"12345678"}`

	externalVerifier := code.NewExternalVerifier(mockDeps)

	t.Run("method=VerificationStart", func(t *testing.T) {
		t.Run("case=returns no error for 2xx response", func(t *testing.T) {
			mockHTTPClient.HTTPClient = &http.Client{
				Transport: &mockTransport{
					response: &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewReader([]byte("OK"))),
					},
				},
			}

			err := externalVerifier.VerificationStart(ctx, mockSMSTemplate)
			require.NoError(t, err)
		})

		t.Run("case=returns error for 4xx response", func(t *testing.T) {
			mockHTTPClient.HTTPClient = &http.Client{
				Transport: &mockTransport{
					response: &http.Response{
						StatusCode: 400,
						Body:       io.NopCloser(bytes.NewReader([]byte("Bad Request"))),
					},
				},
			}

			err := externalVerifier.VerificationStart(ctx, mockSMSTemplate)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "upstream server replied with status code 400 and body Bad Request")
		})

		t.Run("case=returns error for 5xx response", func(t *testing.T) {
			mockHTTPClient.HTTPClient = &http.Client{
				Transport: &mockTransport{
					response: &http.Response{
						StatusCode: 500,
						Body:       io.NopCloser(bytes.NewReader([]byte("Internal Server Error"))),
					},
				},
			}

			err := externalVerifier.VerificationStart(ctx, mockSMSTemplate)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "giving up after 1 attempt(s)")
		})
	})
}

func TestVerificationCheck(t *testing.T) {
	ctx := context.Background()

	mockHTTPClient := new(retryablehttp.Client)
	mockHTTPClient.CheckRetry = driver.NoRetryOnRateLimitPolicy
	mockDeps := NewMockDependencies(t, mockHTTPClient)

	mockSMSTemplate := new(MockSMSTemplate)
	mockSMSTemplate.marshalJson = `{"To":"12345678", "Code":"1234"}`

	externalVerifier := code.NewExternalVerifier(mockDeps)

	t.Run("method=VerificationCheck", func(t *testing.T) {
		t.Run("case=returns no error for 2xx response", func(t *testing.T) {
			mockHTTPClient.HTTPClient = &http.Client{
				Transport: &mockTransport{
					response: &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewReader([]byte("OK"))),
					},
				},
			}

			err := externalVerifier.VerificationCheck(ctx, mockSMSTemplate)
			require.NoError(t, err)
		})

		t.Run("case=returns error for 4xx response", func(t *testing.T) {
			mockHTTPClient.HTTPClient = &http.Client{
				Transport: &mockTransport{
					response: &http.Response{
						StatusCode: 400,
						Body:       io.NopCloser(bytes.NewReader([]byte("Bad Request"))),
					},
				},
			}

			err := externalVerifier.VerificationCheck(ctx, mockSMSTemplate)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "The requested resource could not be found")
		})

		t.Run("case=returns error for 5xx response", func(t *testing.T) {
			mockHTTPClient.HTTPClient = &http.Client{
				Transport: &mockTransport{
					response: &http.Response{
						StatusCode: 500,
						Body:       io.NopCloser(bytes.NewReader([]byte("Internal Server Error"))),
					},
				},
			}

			err := externalVerifier.VerificationCheck(ctx, mockSMSTemplate)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "giving up after 1 attempt(s)")
		})
	})
}

type mockTransport struct {
	response *http.Response
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.response, nil
}
