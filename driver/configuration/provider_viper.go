package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/stringsx"
	"github.com/ory/x/tracing"

	"github.com/ory/viper"

	"github.com/ory/x/jsonx"

	"github.com/ory/x/urlx"
	"github.com/ory/x/viperx"
)

type ViperProvider struct {
	l   *logrusx.Logger
	dev bool
}

var _ Provider = new(ViperProvider)

const DefaultBrowserReturnURL = "default_browser_return_url"

const DefaultSQLiteMemoryDSN = "sqlite://:memory:?_fk=true"

const UnknownVersion = "unknown version"

const (
	ViperKeyDSN = "dsn"

	ViperKeyCourierSMTPURL       = "courier.smtp.connection_uri"
	ViperKeyCourierTemplatesPath = "courier.template_override_path"
	ViperKeyCourierSMTPFrom      = "courier.smtp.from_address"

	ViperKeySecretsDefault = "secrets.default"
	ViperKeySecretsCookie  = "secrets.cookie"

	ViperKeyPublicBaseURL = "serve.public.base_url"
	ViperKeyPublicPort    = "serve.public.port"
	ViperKeyPublicHost    = "serve.public.host"
	ViperKeyAdminBaseURL  = "serve.admin.base_url"
	ViperKeyAdminPort     = "serve.admin.port"
	ViperKeyAdminHost     = "serve.admin.host"

	ViperKeySessionLifespan         = "session.lifespan"
	ViperKeySessionSameSite         = "session.cookie.same_site"
	ViperKeySessionDomain           = "session.cookie.domain"
	ViperKeySessionPath             = "session.cookie.path"
	ViperKeySessionPersistentCookie = "session.cookie.persistent"

	ViperKeySelfServiceStrategyConfig = "selfservice.strategies"

	ViperKeySelfServiceBrowserDefaultReturnTo = "selfservice." + DefaultBrowserReturnURL
	ViperKeyURLsWhitelistedReturnToDomains    = "selfservice.whitelisted_return_urls"

	ViperKeySelfServiceRegistrationUI              = "selfservice.flows.registration.ui_url"
	ViperKeySelfServiceRegistrationRequestLifespan = "selfservice.flows.registration.request_lifespan"
	ViperKeySelfServiceRegistrationAfter           = "selfservice.flows.registration.after"
	ViperKeySelfServiceRegistrationBeforeHooks     = "selfservice.flows.registration.before.hooks"

	ViperKeySelfServiceLoginUI              = "selfservice.flows.login.ui_url"
	ViperKeySelfServiceLoginRequestLifespan = "selfservice.flows.login.request_lifespan"
	ViperKeySelfServiceLoginAfter           = "selfservice.flows.login.after"
	ViperKeySelfServiceLoginBeforeHooks     = "selfservice.flows.login.before.hooks"

	ViperKeySelfServiceErrorUI = "selfservice.flows.error.ui_url"

	ViperKeySelfServiceLogoutBrowserDefaultReturnTo = "selfservice.flows.logout.after." + DefaultBrowserReturnURL

	ViperKeySelfServiceSettingsURL                           = "selfservice.flows.settings.ui_url"
	ViperKeySelfServiceSettingsAfter                         = "selfservice.flows.settings.after"
	ViperKeySelfServiceSettingsRequestLifespan               = "selfservice.flows.settings.request_lifespan"
	ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter = "selfservice.flows.settings.privileged_session_max_age"

	ViperKeySelfServiceRecoveryEnabled                = "selfservice.flows.recovery.enabled"
	ViperKeySelfServiceRecoveryUI                     = "selfservice.flows.recovery.ui_url"
	ViperKeySelfServiceRecoveryRequestLifespan        = "selfservice.flows.recovery.request_lifespan"
	ViperKeySelfServiceRecoveryBrowserDefaultReturnTo = "selfservice.flows.recovery.after." + DefaultBrowserReturnURL

	ViperKeySelfServiceVerificationEnabled                = "selfservice.flows.verification.enabled"
	ViperKeySelfServiceVerificationUI                     = "selfservice.flows.verification.ui_url"
	ViperKeySelfServiceVerificationRequestLifespan        = "selfservice.flows.verification.request_lifespan"
	ViperKeySelfServiceVerificationBrowserDefaultReturnTo = "selfservice.flows.verification.after." + DefaultBrowserReturnURL

	ViperKeyDefaultIdentitySchemaURL = "identity.default_schema_url"
	ViperKeyIdentitySchemas          = "identity.schemas"

	ViperKeyHasherArgon2ConfigMemory      = "hashers.argon2.memory"
	ViperKeyHasherArgon2ConfigIterations  = "hashers.argon2.iterations"
	ViperKeyHasherArgon2ConfigParallelism = "hashers.argon2.parallelism"
	ViperKeyHasherArgon2ConfigSaltLength  = "hashers.argon2.salt_length"
	ViperKeyHasherArgon2ConfigKeyLength   = "hashers.argon2.key_length"

	ViperKeyVersion = "version"
)

func HookStrategyKey(key, strategy string) string {
	return fmt.Sprintf("%s.%s.hooks", key, strategy)
}

func NewViperProvider(l *logrusx.Logger, dev bool) *ViperProvider {
	return &ViperProvider{l: l, dev: dev}
}

func (p *ViperProvider) SessionDomain() string {
	return viperx.GetString(p.l, ViperKeySessionDomain, "")
}

func (p *ViperProvider) SessionPath() string {
	return viperx.GetString(p.l, ViperKeySessionPath, "")
}

func (p *ViperProvider) HasherArgon2() *HasherArgon2Config {
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
	return mustParseURLFromViper(p.l, ViperKeyDefaultIdentitySchemaURL)
}

func (p *ViperProvider) IdentityTraitsSchemas() SchemaConfigs {
	ds := SchemaConfig{
		ID:  DefaultIdentityTraitsSchemaID,
		URL: p.DefaultIdentityTraitsSchemaURL().String(),
	}
	var b bytes.Buffer
	var ss SchemaConfigs
	raw := viper.Get(ViperKeyIdentitySchemas)

	if raw == nil {
		return SchemaConfigs{ds}
	}

	if err := json.NewEncoder(&b).Encode(raw); err != nil {
		p.l.WithError(err).Fatalf("Unable to encode values from %s.", ViperKeyIdentitySchemas)
	}

	if err := jsonx.NewStrictDecoder(&b).Decode(&ss); err != nil {
		p.l.WithError(err).Fatalf("Unable to decode values from %s.", ViperKeyIdentitySchemas)
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
		return DefaultSQLiteMemoryDSN
	}

	if len(dsn) > 0 {
		return dsn
	}

	p.l.Fatal("dsn must be set")
	return ""
}

func (p *ViperProvider) SelfServiceFlowVerificationEnabled() bool {
	return viperx.GetBool(p.l, ViperKeySelfServiceVerificationEnabled, false)
}

func (p *ViperProvider) SelfServiceFlowRecoveryEnabled() bool {
	return viperx.GetBool(p.l, ViperKeySelfServiceRecoveryEnabled, false)
}

func (p *ViperProvider) SelfServiceFlowLoginBeforeHooks() []SelfServiceHook {
	return p.selfServiceHooks(ViperKeySelfServiceLoginBeforeHooks)
}

func (p *ViperProvider) SelfServiceFlowRegistrationBeforeHooks() []SelfServiceHook {
	return p.selfServiceHooks(ViperKeySelfServiceRegistrationBeforeHooks)
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

func (p *ViperProvider) SelfServiceFlowLoginAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(HookStrategyKey(ViperKeySelfServiceLoginAfter, strategy))
}

func (p *ViperProvider) SelfServiceFlowSettingsAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(HookStrategyKey(ViperKeySelfServiceSettingsAfter, strategy))
}

func (p *ViperProvider) SelfServiceFlowRegistrationAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(HookStrategyKey(ViperKeySelfServiceRegistrationAfter, strategy))
}

func (p *ViperProvider) SelfServiceStrategy(strategy string) *SelfServiceStrategy {
	var s SelfServiceStrategy
	if err := viperx.UnmarshalKey(ViperKeySelfServiceStrategyConfig+"."+strategy, &s); err != nil {
		p.l.WithError(errors.WithStack(err)).Fatalf("Unable to encode values from configuration for strategy %s", strategy)
	}

	// FIXME The default value can for some reason not be set from the JSON Schema. This is a workaround.
	// FIXME see https://github.com/ory/x/issues/169
	if viper.Get(fmt.Sprintf("%s.%s.enabled", ViperKeySelfServiceStrategyConfig, strategy)) == nil {
		switch strategy {
		case "password":
			fallthrough
		case "profile":
			fallthrough
		case "link":
			s.Enabled = true
		}
	}

	if len(s.Config) == 0 {
		s.Config = json.RawMessage("{}")
	}

	return &s
}

func (p *ViperProvider) SecretsDefault() [][]byte {
	secrets := viperx.GetStringSlice(p.l, ViperKeySecretsDefault, nil)
	if len(secrets) == 0 {
		secrets = []string{uuid.New().String()}
		viper.Set(ViperKeySecretsDefault, secrets)
	}

	result := make([][]byte, len(secrets))
	for k, v := range secrets {
		result[k] = []byte(v)
	}

	return result
}

func (p *ViperProvider) SecretsSession() [][]byte {
	secrets := viperx.GetStringSlice(p.l, ViperKeySecretsCookie, nil)
	if len(secrets) == 0 {
		return p.SecretsDefault()
	}

	result := make([][]byte, len(secrets))
	for k, v := range secrets {
		result[k] = []byte(v)
	}

	return result
}

func (p *ViperProvider) SelfServiceBrowserDefaultReturnTo() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeySelfServiceBrowserDefaultReturnTo)
}

func (p *ViperProvider) guessBaseURL(keyHost, keyPort string, defaultPort int) *url.URL {
	port := viperx.GetInt(p.l, keyPort, defaultPort)

	host := viper.GetString(keyHost)
	if host == "0.0.0.0" || len(host) == 0 {
		var err error
		host, err = os.Hostname()
		if err != nil {
			p.l.WithError(err).Warn("Unable to get hostname from system, falling back to 127.0.0.1.")
			host = "127.0.0.1"
		}
	}

	guess := url.URL{Host: fmt.Sprintf("%s:%d", host, port), Scheme: "https", Path: "/"}
	if p.IsInsecureDevMode() {
		guess.Scheme = "http"
	}

	return &guess
}

func (p *ViperProvider) baseURL(keyURL, keyHost, keyPort string, defaultPort int) *url.URL {
	baseURL := viper.GetString(keyURL)
	if baseURL == "" {
		p.l.Warnf("Configuration key %s was left empty. Optimistically guessing the server's base URL. Please set a value to avoid problems with redirects and cookies.", ViperKeyPublicBaseURL)
		return p.guessBaseURL(keyHost, keyPort, defaultPort)
	}

	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil {
		p.l.WithError(err).Errorf("Configuration key %s is not a valid URL. Falling back to optimistically guessing the server's base URL. Please set a value to avoid problems with redirects and cookies.", ViperKeyPublicBaseURL)
		return p.guessBaseURL(keyHost, keyPort, defaultPort)
	}

	return parsed
}

func (p *ViperProvider) SelfPublicURL() *url.URL {
	return p.baseURL(ViperKeyPublicBaseURL, ViperKeyPublicHost, ViperKeyPublicPort, 4433)
}

func (p *ViperProvider) SelfAdminURL() *url.URL {
	return p.baseURL(ViperKeyAdminBaseURL, ViperKeyAdminHost, ViperKeyAdminPort, 4434)
}

func (p *ViperProvider) CourierSMTPURL() *url.URL {
	u, err := url.ParseRequestURI(viper.GetString(ViperKeyCourierSMTPURL))
	if err != nil {
		return new(url.URL)
	}
	return u
}

func (p *ViperProvider) SelfServiceFlowLoginUI() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeySelfServiceLoginUI)
}

func (p *ViperProvider) SelfServiceFlowSettingsUI() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeySelfServiceSettingsURL)
}

func (p *ViperProvider) SelfServiceFlowErrorURL() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeySelfServiceErrorUI)
}

func (p *ViperProvider) SelfServiceFlowRegistrationUI() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeySelfServiceRegistrationUI)
}

func (p *ViperProvider) SelfServiceFlowRecoveryUI() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeySelfServiceRecoveryUI)
}

// SessionLifespan returns nil when the value is not set.
func (p *ViperProvider) SessionLifespan() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySessionLifespan, time.Hour*24)
}

func (p *ViperProvider) SessionPersistentCookie() bool {
	return viper.GetBool(ViperKeySessionPersistentCookie)
}

func (p *ViperProvider) SelfServiceBrowserWhitelistedReturnToDomains() (us []url.URL) {
	src := viperx.GetStringSlice(p.l, ViperKeyURLsWhitelistedReturnToDomains, []string{})
	for _, u := range src {
		if len(u) > 0 {
			us = append(us, *urlx.ParseOrFatal(p.l, u))
		}
	}

	return us
}

func (p *ViperProvider) SelfServiceFlowLoginRequestLifespan() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySelfServiceLoginRequestLifespan, time.Hour)
}

func (p *ViperProvider) SelfServiceFlowSettingsFlowLifespan() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySelfServiceSettingsRequestLifespan, time.Hour)
}

func (p *ViperProvider) SelfServiceFlowRegistrationRequestLifespan() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySelfServiceRegistrationRequestLifespan, time.Hour)
}

func (p *ViperProvider) SelfServiceFlowLogoutRedirectURL() *url.URL {
	redir, err := url.ParseRequestURI(
		viperx.GetString(p.l, ViperKeySelfServiceLogoutBrowserDefaultReturnTo, ""))
	if err != nil {
		return p.SelfServiceBrowserDefaultReturnTo()
	}
	return redir
}

func (p *ViperProvider) CourierSMTPFrom() string {
	return viperx.GetString(p.l, ViperKeyCourierSMTPFrom, "noreply@kratos.ory.sh")
}

func (p *ViperProvider) CourierTemplatesRoot() string {
	return viperx.GetString(p.l, ViperKeyCourierTemplatesPath, "/courier/template/templates")
}

func mustParseURLFromViper(l *logrusx.Logger, key string) *url.URL {
	u, err := url.ParseRequestURI(viper.GetString(key))
	if err != nil {
		l.WithError(errors.WithStack(err)).
			Fatalf("Configuration value from key %s is not a valid URL: %s", key, viper.GetString(key))
	}
	return u
}

func (p *ViperProvider) TracingServiceName() string {
	return viperx.GetString(p.l, "tracing.service_name", "ORY Kratos")
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

func (p *ViperProvider) SelfServiceFlowVerificationUI() *url.URL {
	return mustParseURLFromViper(p.l, ViperKeySelfServiceVerificationUI)
}

func (p *ViperProvider) SelfServiceFlowVerificationRequestLifespan() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySelfServiceVerificationRequestLifespan, time.Hour)
}

func (p *ViperProvider) SelfServiceFlowVerificationReturnTo() *url.URL {
	redir, err := url.ParseRequestURI(
		viperx.GetString(p.l, ViperKeySelfServiceVerificationBrowserDefaultReturnTo, ""))
	if err != nil {
		return p.SelfServiceBrowserDefaultReturnTo()
	}
	return redir
}

func (p *ViperProvider) SelfServiceFlowRecoveryReturnTo() *url.URL {
	redir, err := url.ParseRequestURI(
		viperx.GetString(p.l, ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, ""))
	if err != nil {
		return p.SelfServiceBrowserDefaultReturnTo()
	}
	return redir
}

func (p *ViperProvider) SelfServiceFlowRecoveryRequestLifespan() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySelfServiceRecoveryRequestLifespan, time.Hour)
}

func (p *ViperProvider) SelfServiceFlowSettingsPrivilegedSessionMaxAge() time.Duration {
	return viperx.GetDuration(p.l, ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Hour)
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

func (p *ViperProvider) SelfServiceFlowLoginReturnTo(strategy string) *url.URL {
	return p.selfServiceReturnTo(ViperKeySelfServiceLoginAfter, strategy)
}

func (p *ViperProvider) SelfServiceFlowRegistrationReturnTo(strategy string) *url.URL {
	return p.selfServiceReturnTo(ViperKeySelfServiceRegistrationAfter, strategy)
}

func (p *ViperProvider) SelfServiceFlowSettingsReturnTo(strategy string, defaultReturnTo *url.URL) *url.URL {
	redir, err := url.ParseRequestURI(
		viperx.GetString(p.l, ViperKeySelfServiceSettingsAfter+"."+strategy+"."+DefaultBrowserReturnURL,
			viperx.GetString(p.l, ViperKeySelfServiceSettingsAfter+"."+DefaultBrowserReturnURL, "")))
	if err != nil {
		return defaultReturnTo
	}
	return redir
}

func (p *ViperProvider) selfServiceReturnTo(key string, strategy string) *url.URL {
	redir, err := url.ParseRequestURI(
		viperx.GetString(p.l, key+"."+strategy+"."+DefaultBrowserReturnURL, viperx.GetString(p.l, key+"."+DefaultBrowserReturnURL, "")))
	if err != nil {
		return p.SelfServiceBrowserDefaultReturnTo()
	}
	return redir
}

func (p *ViperProvider) ConfigVersion() string {
	return viperx.GetString(p.l, ViperKeyVersion, UnknownVersion)
}
