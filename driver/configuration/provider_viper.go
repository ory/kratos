package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	l  logrus.FieldLogger
	ss [][]byte
}

var _ Provider = new(ViperProvider)

const (
	ViperKeyDSN = "dsn"

	ViperKeySecretsSession = "secrets.session"

	ViperKeyURLsDefaultReturnTo            = "urls.default_return_to"
	ViperKeyURLsSelfPublic                 = "urls.self.public"
	ViperKeyURLsSelfAdmin                  = "urls.self.admin"
	ViperKeyURLsLogin                      = "urls.login_ui"
	ViperKeyURLsError                      = "urls.error_ui"
	ViperKeyURLsMFA                        = "urls.mfa_ui"
	ViperKeyURLsRegistration               = "urls.registration_ui"
	ViperKeyURLsWhitelistedReturnToDomains = "urls.whitelisted_return_to_domains"

	ViperKeyLifespanSession = "ttl.session"

	ViperKeySelfServiceStrategyConfig              = "selfservice.strategies"
	ViperKeySelfServiceRegistrationBeforeConfig    = "selfservice.registration.before"
	ViperKeySelfServiceRegistrationAfterConfig     = "selfservice.registration.after"
	ViperKeySelfServiceLifespanRegistrationRequest = "selfservice.registration.request_lifespan"
	ViperKeySelfServiceLoginBeforeConfig           = "selfservice.login.before"
	ViperKeySelfServiceLoginAfterConfig            = "selfservice.login.after"
	ViperKeySelfServiceLifespanLoginRequest        = "selfservice.login.request_lifespan"
	ViperKeySelfServiceLogoutRedirectURL           = "selfservice.logout.redirect_to"

	ViperKeyDefaultIdentityTraitsSchemaURL = "identity.traits.default_schema_url"

	ViperKeyHasherArgon2ConfigMemory      = "hashers.argon2.memory"
	ViperKeyHasherArgon2ConfigIterations  = "hashers.argon2.iterations"
	ViperKeyHasherArgon2ConfigParallelism = "hashers.argon2.parallelism"
	ViperKeyHasherArgon2ConfigSaltLength  = "hashers.argon2.salt_length"
	ViperKeyHasherArgon2ConfigKeyLength   = "hashers.argon2.key_length"
)

func BindEnvs() {
	if err := viper.BindEnv(
		ViperKeyURLsSelfPublic,
		ViperKeyURLsSelfAdmin,
		ViperKeyURLsLogin,
		ViperKeyURLsError,
		ViperKeyURLsMFA,
		ViperKeyURLsRegistration,
		ViperKeyURLsWhitelistedReturnToDomains,
		ViperKeyDSN,
		ViperKeySecretsSession,
		ViperKeySelfServiceLifespanRegistrationRequest,
		ViperKeySelfServiceLifespanLoginRequest,
		ViperKeyLifespanSession,
		ViperKeySelfServiceStrategyConfig,
		ViperKeyDefaultIdentityTraitsSchemaURL,
		ViperKeyHasherArgon2ConfigMemory,
		ViperKeyHasherArgon2ConfigIterations,
		ViperKeyHasherArgon2ConfigParallelism,
		ViperKeyHasherArgon2ConfigSaltLength,
		ViperKeyHasherArgon2ConfigKeyLength,
		ViperKeySelfServiceLoginBeforeConfig,
		ViperKeySelfServiceRegistrationBeforeConfig,
		ViperKeySelfServiceRegistrationAfterConfig,
		ViperKeySelfServiceLoginAfterConfig,
		ViperKeySelfServiceLogoutRedirectURL,
		ViperKeyURLsDefaultReturnTo,
	); err != nil {
		panic(err.Error())
	}
}

func NewViperProvider(l logrus.FieldLogger) *ViperProvider {
	return &ViperProvider{
		l: l,
	}
}

func (c *ViperProvider) HashersArgon2() *HasherArgon2Config {
	return &HasherArgon2Config{
		Memory:      uint32(viperx.GetInt(c.l, ViperKeyHasherArgon2ConfigMemory, 4*1024*1024)),
		Iterations:  uint32(viperx.GetInt(c.l, ViperKeyHasherArgon2ConfigIterations, 4)),
		Parallelism: uint8(viperx.GetInt(c.l, ViperKeyHasherArgon2ConfigParallelism, runtime.NumCPU()*2)),
		SaltLength:  uint32(viperx.GetInt(c.l, ViperKeyHasherArgon2ConfigSaltLength, 16)),
		KeyLength:   uint32(viperx.GetInt(c.l, ViperKeyHasherArgon2ConfigKeyLength, 32)),
	}
}

func (c *ViperProvider) listenOn(key string) string {
	fb := 4433
	if key == "admin" {
		fb = 4434
	}

	port := viperx.GetInt(c.l, "serve."+key+".port", fb)
	if port < 1 {
		c.l.Fatalf("serve.%s.port can not be zero or negative", key)
	}

	return fmt.Sprintf("%s:%d", viper.GetString("serve."+key+".host"), port)
}

func (c *ViperProvider) DefaultIdentityTraitsSchemaURL() *url.URL {
	return mustParseURLFromViper(c.l, ViperKeyDefaultIdentityTraitsSchemaURL)
}

func (c *ViperProvider) AdminListenOn() string {
	return c.listenOn("admin")
}

func (c *ViperProvider) PublicListenOn() string {
	return c.listenOn("public")
}

func (c *ViperProvider) DSN() string {
	if dsn := viperx.GetString(c.l, ViperKeyDSN, ""); len(dsn) > 0 {
		return dsn
	}

	c.l.Fatal("dsn must be set")
	return ""
}

func (c *ViperProvider) SelfServiceLoginBeforeHooks() []SelfServiceHook {
	return c.selfServiceHooks(ViperKeySelfServiceLoginBeforeConfig)
}

func (c *ViperProvider) SelfServiceRegistrationBeforeHooks() []SelfServiceHook {
	return c.selfServiceHooks(ViperKeySelfServiceRegistrationBeforeConfig)
}

func (c *ViperProvider) selfServiceHooks(key string) []SelfServiceHook {
	var b bytes.Buffer
	var hooks []SelfServiceHook
	raw := viper.Get(key)

	if raw == nil {
		return []SelfServiceHook{}
	}

	if err := json.NewEncoder(&b).Encode(raw); err != nil {
		c.l.WithError(err).Fatalf("Unable to decode values from configuration key: %s", key)
	}

	if err := jsonx.NewStrictDecoder(&b).Decode(&hooks); err != nil {
		c.l.WithError(err).Fatalf("Unable to encode values from configuration key: %s", key)
	}

	for k := range hooks {
		if len(hooks[k].Config) == 0 {
			hooks[k].Config = json.RawMessage("{}")
		}
	}

	return hooks
}

func (c *ViperProvider) SelfServiceLoginAfterHooks(strategy string) []SelfServiceHook {
	return c.selfServiceHooks(ViperKeySelfServiceLoginAfterConfig + "." + strategy)
}

func (c *ViperProvider) SelfServiceRegistrationAfterHooks(strategy string) []SelfServiceHook {
	return c.selfServiceHooks(ViperKeySelfServiceRegistrationAfterConfig + "." + strategy)
}

func (c *ViperProvider) SelfServiceStrategy(strategy string) *SelfServiceStrategy {
	configs := viper.GetStringMap(ViperKeySelfServiceStrategyConfig)
	config, ok := configs[strategy]
	if !ok || config == `null` || config == "" || config == nil {
		config = map[string]interface{}{}
	}

	var b bytes.Buffer
	var s SelfServiceStrategy
	if err := json.NewEncoder(&b).Encode(config); err != nil {
		c.l.WithError(errors.WithStack(err)).WithField("configs", fmt.Sprintf("%s", configs)).WithField("config", fmt.Sprintf("%s", config)).Fatalf("Unable to encode values from configuration for strategy %s", strategy)
	} else if err := jsonx.NewStrictDecoder(&b).Decode(&s); err != nil {
		c.l.WithError(errors.WithStack(err)).WithField("configs", fmt.Sprintf("%s", configs)).WithField("config", fmt.Sprintf("%s", config)).Fatalf("Unable to decode values from configuration for strategy %s", strategy)
	}

	if len(s.Config) == 0 {
		s.Config = json.RawMessage("{}")
	}

	return &s
}

func (c *ViperProvider) SessionSecrets() [][]byte {
	secrets := viperx.GetStringSlice(c.l, ViperKeySecretsSession, []string{})

	if len(secrets) == 0 {
		if c.ss == nil {
			c.ss = [][]byte{[]byte(uuid.New().String())}
		}
		return c.ss
	}

	result := make([][]byte, len(secrets))
	for k, v := range secrets {
		result[k] = []byte(v)
	}

	return result
}

func (c *ViperProvider) DefaultReturnToURL() *url.URL {
	return mustParseURLFromViper(c.l, ViperKeyURLsDefaultReturnTo)
}

func (c *ViperProvider) SelfPublicURL() *url.URL {
	return mustParseURLFromViper(c.l, ViperKeyURLsSelfPublic)
}

func (c *ViperProvider) SelfAdminURL() *url.URL {
	return mustParseURLFromViper(c.l, ViperKeyURLsSelfAdmin)
}

func (c *ViperProvider) LoginURL() *url.URL {
	return mustParseURLFromViper(c.l, ViperKeyURLsLogin)
}

func (c *ViperProvider) ErrorURL() *url.URL {
	return mustParseURLFromViper(c.l, ViperKeyURLsError)
}

func (c *ViperProvider) MultiFactorURL() *url.URL {
	return mustParseURLFromViper(c.l, ViperKeyURLsMFA)
}

func (c *ViperProvider) RegisterURL() *url.URL {
	return mustParseURLFromViper(c.l, ViperKeyURLsRegistration)
}

func (c *ViperProvider) SessionLifespan() time.Duration {
	return viperx.GetDuration(c.l, ViperKeyLifespanSession, time.Hour)
}

func (c *ViperProvider) WhitelistedReturnToDomains() (us []url.URL) {
	src := viperx.GetStringSlice(c.l, ViperKeyURLsWhitelistedReturnToDomains, []string{})
	for _, u := range src {
		if len(u) > 0 {
			us = append(us, *urlx.ParseOrFatal(c.l, u))
		}
	}

	return us
}

func (c *ViperProvider) SelfServiceLoginRequestLifespan() time.Duration {
	return viperx.GetDuration(c.l, ViperKeySelfServiceLifespanLoginRequest, time.Minute*15)
}

func (c *ViperProvider) SelfServiceRegistrationRequestLifespan() time.Duration {
	return viperx.GetDuration(c.l, ViperKeySelfServiceLifespanRegistrationRequest, time.Minute*15)
}

func (c *ViperProvider) SelfServiceLogoutRedirectURL() *url.URL {
	return mustParseURLFromViper(c.l, ViperKeySelfServiceLogoutRedirectURL)
}

func mustParseURLFromViper(l logrus.FieldLogger, key string) *url.URL {
	u, err := url.ParseRequestURI(viper.GetString(key))
	if err != nil {
		l.WithError(err).WithField("stack", fmt.Sprintf("%+v", errors.WithStack(err))).Fatalf("Configuration value from key %s is not a valid URL: %s", key, viper.GetString(key))
	}
	return u
}

func (v *ViperProvider) TracingServiceName() string {
	return viperx.GetString(v.l, "tracing.service_name", "ORY Hydra")
}

func (v *ViperProvider) TracingProvider() string {
	return viperx.GetString(v.l, "tracing.provider", "", "TRACING_PROVIDER")
}

func (v *ViperProvider) TracingJaegerConfig() *tracing.JaegerConfig {
	return &tracing.JaegerConfig{
		LocalAgentHostPort: viperx.GetString(v.l, "tracing.providers.jaeger.local_agent_address", "", "TRACING_PROVIDER_JAEGER_LOCAL_AGENT_ADDRESS"),
		SamplerType:        viperx.GetString(v.l, "tracing.providers.jaeger.sampling.type", "const", "TRACING_PROVIDER_JAEGER_SAMPLING_TYPE"),
		SamplerValue:       viperx.GetFloat64(v.l, "tracing.providers.jaeger.sampling.value", float64(1), "TRACING_PROVIDER_JAEGER_SAMPLING_VALUE"),
		SamplerServerURL:   viperx.GetString(v.l, "tracing.providers.jaeger.sampling.server_url", "", "TRACING_PROVIDER_JAEGER_SAMPLING_SERVER_URL"),
		Propagation: stringsx.Coalesce(
			viper.GetString("JAEGER_PROPAGATION"), // Standard Jaeger client config
			viperx.GetString(v.l, "tracing.providers.jaeger.propagation", "", "TRACING_PROVIDER_JAEGER_PROPAGATION"),
		),
	}
}
