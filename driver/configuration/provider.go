package configuration

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/x/tracing"
)

type HasherArgon2Config struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type SelfServiceHook struct {
	Name   string          `json:"hook"`
	Config json.RawMessage `json:"config"`
}

type SelfServiceStrategy struct {
	Enabled bool            `json:"enabled"`
	Config  json.RawMessage `json:"config"`
}

type SchemaConfig struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type SchemaConfigs []SchemaConfig

func (s SchemaConfigs) FindSchemaByID(id string) (*SchemaConfig, error) {
	for _, sc := range s {
		if sc.ID == id {
			return &sc, nil
		}
	}

	return nil, errors.Errorf("could not find schema with id \"%s\"", id)
}

const DefaultIdentityTraitsSchemaID = "default"

type Provider interface {
	IsInsecureDevMode() bool
	DisableAPIFlowEnforcement() bool

	DSN() string

	AdminListenOn() string
	PublicListenOn() string

	SelfServiceBrowserDefaultReturnTo() *url.URL
	SelfServiceBrowserWhitelistedReturnToDomains() []url.URL

	SelfPublicURL() *url.URL
	SelfAdminURL() *url.URL

	SecretsDefault() [][]byte
	SecretsSession() [][]byte
	SessionLifespan() time.Duration
	SessionPersistentCookie() bool
	SessionSameSiteMode() http.SameSite
	SessionDomain() string
	SessionPath() string

	SelfServiceStrategy(strategy string) *SelfServiceStrategy

	SelfServiceFlowLoginUI() *url.URL
	SelfServiceFlowLoginBeforeHooks() []SelfServiceHook
	SelfServiceFlowLoginAfterHooks(strategy string) []SelfServiceHook
	SelfServiceFlowLoginReturnTo(strategy string) *url.URL
	SelfServiceFlowLoginRequestLifespan() time.Duration

	SelfServiceFlowRegistrationUI() *url.URL
	SelfServiceFlowRegistrationBeforeHooks() []SelfServiceHook
	SelfServiceFlowRegistrationAfterHooks(strategy string) []SelfServiceHook
	SelfServiceFlowRegistrationReturnTo(strategy string) *url.URL
	SelfServiceFlowRegistrationRequestLifespan() time.Duration

	SelfServiceFlowSettingsUI() *url.URL
	SelfServiceFlowSettingsPrivilegedSessionMaxAge() time.Duration
	SelfServiceFlowSettingsAfterHooks(strategy string) []SelfServiceHook
	SelfServiceFlowSettingsReturnTo(strategy string, defaultReturnTo *url.URL) *url.URL
	SelfServiceFlowSettingsFlowLifespan() time.Duration

	SelfServiceFlowVerificationEnabled() bool
	SelfServiceFlowVerificationUI() *url.URL
	SelfServiceFlowVerificationReturnTo(defaultReturnTo *url.URL) *url.URL
	SelfServiceFlowVerificationRequestLifespan() time.Duration

	SelfServiceFlowRecoveryEnabled() bool
	SelfServiceFlowRecoveryUI() *url.URL
	SelfServiceFlowRecoveryRequestLifespan() time.Duration

	SelfServiceFlowErrorURL() *url.URL

	SelfServiceFlowLogoutRedirectURL() *url.URL

	CourierSMTPFrom() string
	CourierSMTPURL() *url.URL
	CourierTemplatesRoot() string

	DefaultIdentityTraitsSchemaURL() *url.URL
	IdentityTraitsSchemas() SchemaConfigs

	HasherArgon2() *HasherArgon2Config

	TracingServiceName() string
	TracingProvider() string
	TracingJaegerConfig() *tracing.JaegerConfig

	ConfigVersion() string
}
