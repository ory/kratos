package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ory/x/stringsx"
	"github.com/ory/x/tracing"

	"github.com/ory/viper"

	"github.com/ory/x/jsonx"

	"github.com/ory/x/urlx"
	"github.com/ory/x/viperx"
)

type ViperProvider struct {
	l   logrus.FieldLogger
	ss  [][]byte
	dev bool
}

var _ Provider = new(ViperProvider)

const (
	ViperKeyDSN = "dsn"

	ViperKeyCourierSMTPURL       = "courier.smtp.connection_uri"
	ViperKeyCourierTemplatesPath = "courier.template_override_path"
	ViperKeyCourierSMTPFrom      = "courier.smtp.from_address"

	ViperKeySecretsSession = "secrets.session"

	ViperKeyURLsDefaultReturnTo            = "urls.default_return_to"
	ViperKeyURLsSelfPublic                 = "urls.self.public"
	ViperKeyURLsSelfAdmin                  = "urls.self.admin"
	ViperKeyURLsLogin                      = "urls.login_ui"
	ViperKeyURLsError                      = "urls.error_ui"
	ViperKeyURLsVerification               = "urls.verify_ui"
	ViperKeyURLsProfile                    = "urls.profile_ui"
	ViperKeyURLsMFA                        = "urls.mfa_ui"
	ViperKeyURLsRegistration               = "urls.registration_ui"
	ViperKeyURLsWhitelistedReturnToDomains = "urls.whitelisted_return_to_domains"

	ViperKeyLifespanSession = "ttl.session"

	ViperKeySessionSameSite = "security.session.cookie.same_site"

	ViperKeySelfServiceStrategyConfig                = "selfservice.strategies"
	ViperKeySelfServiceRegistrationBeforeConfig      = "selfservice.registration.before"
	ViperKeySelfServiceRegistrationAfterConfig       = "selfservice.registration.after"
	ViperKeySelfServiceLifespanRegistrationRequest   = "selfservice.registration.request_lifespan"
	ViperKeySelfServiceLoginBeforeConfig             = "selfservice.login.before"
	ViperKeySelfServiceLoginAfterConfig              = "selfservice.login.after"
	ViperKeySelfServiceProfileManagementAfterConfig  = "selfservice.profile_management.after"
	ViperKeySelfServiceLifespanLoginRequest          = "selfservice.login.request_lifespan"
	ViperKeySelfServiceLogoutRedirectURL             = "selfservice.logout.redirect_to"
	ViperKeySelfServiceLifespanProfileRequest        = "selfservice.profile.request_lifespan"
	ViperKeySelfServicePrivilegedAuthenticationAfter = "selfservice.profile.privileged_session_max_age"
	ViperKeySelfServiceLifespanLink                  = "selfservice.profile.link_lifespan"
	ViperKeySelfServiceLifespanVerificationRequest   = "selfservice.verify.request_lifespan"
	ViperKeySelfServiceVerifyReturnTo                = "selfservice.verify.return_to"

	ViperKeyDefaultIdentityTraitsSchemaURL = "identity.traits.default_schema_url"
	ViperKeyIdentityTraitsSchemas          = "identity.traits.schemas"

	ViperKeyHasherArgon2ConfigMemory      = "hashers.argon2.memory"
	ViperKeyHasherArgon2ConfigIterations  = "hashers.argon2.iterations"
	ViperKeyHasherArgon2ConfigParallelism = "hashers.argon2.parallelism"
	ViperKeyHasherArgon2ConfigSaltLength  = "hashers.argon2.salt_length"
	ViperKeyHasherArgon2ConfigKeyLength   = "hashers.argon2.key_length"
)

func NewViperProvider(l logrus.FieldLogger, dev bool) *ViperProvider {
	return &ViperProvider{
		l:   l,
		dev: dev,
	}
}

func (p *ViperProvider) HashersArgon2() *HasherArgon2Config {
	return &HasherArgon2Config{
		Memory:      uint32(viperx.GetInt(p.l, ViperKeyHasherArgon2ConfigMemory, 4*1024*1024)),
		Iterations:  uint32(viperx.GetInt(p.l, ViperKeyHasherArgon2ConfigIterations, 4)),
		Parallelism: uint8(viperx.GetInt(p.l, ViperKeyHasherArgon2ConfigParallelism, runtime.NumCPU()*2)),
		SaltLength:  uint32(viperx.GetInt(p.l, ViperKeyHasherArgon2ConfigSaltLength, 16)),
		KeyLength:   uint32(viperx.GetInt(p.l, ViperKeyHasherArgon2ConfigKeyLength, 32)),
	}
}

func (p *ViperProvider) listenOn(key string) string {
	fb := 4433
	if key == "admin" {
		fb = 4434
	}

	port := viperx.GetInt(p.l, "serve."+key+".port", fb)
	if port < 1 {
		p.l.Fatalf("serve.%s.port can not be zero or negative", key)
	}

	return fmt.Sprintf("%s:%d", viper.GetString("serve."+key+".host"), port)
}

func (p *ViperProvider) DefaultIdentityTraitsSchemaURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeyDefaultIdentityTraitsSchemaURL)
}

func (p *ViperProvider) IdentityTraitsSchemas() SchemaConfigs {
	ds := SchemaConfig{
		ID:  DefaultIdentityTraitsSchemaID,
		URL: p.DefaultIdentityTraitsSchemaURL().String(),
	}
	var b bytes.Buffer
	var ss SchemaConfigs
	raw := viper.Get(ViperKeyIdentityTraitsSchemas)

	if raw == nil {
		return SchemaConfigs{ds}
	}

	if err := json.NewEncoder(&b).Encode(raw); err != nil {
		p.l.WithError(err).Fatalf("Unable to encode values from %s.", ViperKeyIdentityTraitsSchemas)
	}

	if err := jsonx.NewStrictDecoder(&b).Decode(&ss); err != nil {
		p.l.WithError(err).Fatalf("Unable to decode values from %s.", ViperKeyIdentityTraitsSchemas)
	}

	return append(ss, ds)
}

func (p *ViperProvider) AdminListenOn() string {
	return p.listenOn("admin")
}

func (p *ViperProvider) PublicListenOn() string {
	return p.listenOn("public")
}

func (p *ViperProvider) DSN() string {
	dsn := viperx.GetString(p.l, ViperKeyDSN, "")

	if dsn == "memory" {
		return "sqlite://mem.db?mode=memory&_fk=true&cache=shared"
	}

	if len(dsn) > 0 {
		return dsn
	}

	p.l.Fatal("dsn must be set")
	return ""
}

func (p *ViperProvider) SelfServiceLoginBeforeHooks() []SelfServiceHook {
	return p.selfServiceHooks(ViperKeySelfServiceLoginBeforeConfig)
}

func (p *ViperProvider) SelfServiceRegistrationBeforeHooks() []SelfServiceHook {
	return p.selfServiceHooks(ViperKeySelfServiceRegistrationBeforeConfig)
}

func (p *ViperProvider) selfServiceHooks(key string) []SelfServiceHook {
	var b bytes.Buffer
	var hooks []SelfServiceHook
	raw := viper.Get(key)

	if raw == nil {
		return []SelfServiceHook{}
	}

	if err := json.NewEncoder(&b).Encode(raw); err != nil {
		p.l.WithError(err).Fatalf("Unable to decode values from configuration key: %s", key)
	}

	if err := jsonx.NewStrictDecoder(&b).Decode(&hooks); err != nil {
		p.l.WithError(err).Fatalf("Unable to encode values from configuration key: %s", key)
	}

	for k := range hooks {
		if len(hooks[k].Config) == 0 {
			hooks[k].Config = json.RawMessage("{}")
		}
	}

	return hooks
}

func (p *ViperProvider) SelfServiceLoginAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(ViperKeySelfServiceLoginAfterConfig + "." + strategy)
}

func (p *ViperProvider) SelfServiceProfileManagementAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(ViperKeySelfServiceProfileManagementAfterConfig + "." + strategy)
}

func (p *ViperProvider) SelfServiceRegistrationAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(ViperKeySelfServiceRegistrationAfterConfig + "." + strategy)
}

func (p *ViperProvider) SelfServiceStrategy(strategy string) *SelfServiceStrategy {
	configs := viper.GetStringMap(ViperKeySelfServiceStrategyConfig)
	config, ok := configs[strategy]
	if !ok || config == `null` || config == "" || config == nil {
		config = map[string]interface{}{}
	}

	var b bytes.Buffer
	var s SelfServiceStrategy
	if err := json.NewEncoder(&b).Encode(config); err != nil {
		p.l.WithError(errors.WithStack(err)).WithField("configs", fmt.Sprintf("%s", configs)).WithField("config", fmt.Sprintf("%s", config)).Fatalf("Unable to encode values from configuration for strategy %s", strategy)
	} else if err := jsonx.NewStrictDecoder(&b).Decode(&s); err != nil {
		p.l.WithError(errors.WithStack(err)).WithField("configs", fmt.Sprintf("%s", configs)).WithField("config", fmt.Sprintf("%s", config)).Fatalf("Unable to decode values from configuration for strategy %s", strategy)
	}

	if len(s.Config) == 0 {
		s.Config = json.RawMessage("{}")
	}

	return &s
}

func (p *ViperProvider) SessionSecrets() [][]byte {
	secrets := viperx.GetStringSlice(p.l, ViperKeySecretsSession, []string{})

	if len(secrets) == 0 {
		if p.ss == nil {
			p.ss = [][]byte{[]byte(uuid.New().String())}
		}
		return p.ss
	}

	result := make([][]byte, len(secrets))
	for k, v := range secrets {
		result[k] = []byte(v)
	}

	return result
}

func (p *ViperProvider) DefaultReturnToURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeyURLsDefaultReturnTo)
}

func (p *ViperProvider) SelfPublicURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeyURLsSelfPublic)
}

func (p *ViperProvider) SelfAdminURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeyURLsSelfAdmin)
}

func (p *ViperProvider) CourierSMTPURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeyCourierSMTPURL)
}

func (p *ViperProvider) LoginURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeyURLsLogin)
}

func (p *ViperProvider) ProfileURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeyURLsProfile)
}

func (p *ViperProvider) ErrorURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeyURLsError)
}

func (p *ViperProvider) MultiFactorURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeyURLsMFA)
}

func (p *ViperProvider) RegisterURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeyURLsRegistration)
}

func (p *ViperProvider) SessionLifespan() time.Duration {
	return viperx.GetDuration(p.l, ViperKeyLifespanSession, time.Hour)
}

func (p *ViperProvider) WhitelistedReturnToDomains() (us []url.URL) {
	src := viperx.GetStringSlice(p.l, ViperKeyURLsWhitelistedReturnToDomains, []string{})
	for _, u := range src {
		if len(u) > 0 {
			us = append(us, *urlx.ParseOrFatal(p.l, u))
		}
	}

	return us
}

func (p *ViperProvider) SelfServiceLoginRequestLifespan() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySelfServiceLifespanLoginRequest, time.Hour)
}

func (p *ViperProvider) SelfServiceProfileRequestLifespan() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySelfServiceLifespanProfileRequest, time.Hour)
}

func (p *ViperProvider) SelfServiceRegistrationRequestLifespan() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySelfServiceLifespanRegistrationRequest, time.Hour)
}

func (p *ViperProvider) SelfServiceLogoutRedirectURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeySelfServiceLogoutRedirectURL)
}

func (p *ViperProvider) CourierSMTPFrom() string {
	return viperx.GetString(p.l, ViperKeyCourierSMTPFrom, "noreply@kratos.ory.sh")
}

func (p *ViperProvider) CourierTemplatesRoot() string {
	return viperx.GetString(p.l, ViperKeyCourierTemplatesPath, "")
}

func mustParseURLFromViper(l logrus.FieldLogger, key string) *url.URL {
	u, err := url.ParseRequestURI(viper.GetString(key))
	if err != nil {
		l.WithError(err).WithField("stack", fmt.Sprintf("%+v", errors.WithStack(err))).Fatalf("Configuration value from key %s is not a valid URL: %s", key, viper.GetString(key))
	}
	return u
}

func (p *ViperProvider) TracingServiceName() string {
	return viperx.GetString(p.l, "tracing.service_name", "ORY Hydra")
}

func (p *ViperProvider) TracingProvider() string {
	return viperx.GetString(p.l, "tracing.provider", "", "TRACING_PROVIDER")
}

func (p *ViperProvider) TracingJaegerConfig() *tracing.JaegerConfig {
	return &tracing.JaegerConfig{
		LocalAgentHostPort: viperx.GetString(p.l, "tracing.providers.jaeger.local_agent_address", "", "TRACING_PROVIDER_JAEGER_LOCAL_AGENT_ADDRESS"),
		SamplerType:        viperx.GetString(p.l, "tracing.providers.jaeger.sampling.type", "const", "TRACING_PROVIDER_JAEGER_SAMPLING_TYPE"),
		SamplerValue:       viperx.GetFloat64(p.l, "tracing.providers.jaeger.sampling.value", float64(1), "TRACING_PROVIDER_JAEGER_SAMPLING_VALUE"),
		SamplerServerURL:   viperx.GetString(p.l, "tracing.providers.jaeger.sampling.server_url", "", "TRACING_PROVIDER_JAEGER_SAMPLING_SERVER_URL"),
		Propagation: stringsx.Coalesce(
			viper.GetString("JAEGER_PROPAGATION"), // Standard Jaeger client config
			viperx.GetString(p.l, "tracing.providers.jaeger.propagation", "", "TRACING_PROVIDER_JAEGER_PROPAGATION"),
		),
	}
}

func (p *ViperProvider) IsInsecureDevMode() bool {
	return p.dev
}

func (p *ViperProvider) VerificationURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeyURLsVerification)
}

// SelfServiceVerificationRequestLifespan defines the lifespan of a verification request (the ui interaction). This
// does not specify the lifespan of a verification code!
func (p *ViperProvider) SelfServiceVerificationRequestLifespan() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySelfServiceLifespanVerificationRequest, time.Hour)
}

func (p *ViperProvider) SelfServiceVerificationLinkLifespan() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySelfServiceLifespanLink, time.Hour*24)
}

func (p *ViperProvider) SelfServiceVerificationReturnTo() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeySelfServiceVerifyReturnTo)
}

func (p *ViperProvider) SelfServicePrivilegedSessionMaxAge() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySelfServicePrivilegedAuthenticationAfter, time.Hour)
}

func (p *ViperProvider) SessionSameSiteMode() http.SameSite {
	switch viperx.GetString(p.l, ViperKeySessionSameSite, "Lax") {
	case "Lax":
		return http.SameSiteLaxMode
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	}
	return http.SameSiteDefaultMode
}
