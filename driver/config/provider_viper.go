package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/inhies/go-bytesize"

	"github.com/markbates/pkger"
	"github.com/rs/cors"
	"github.com/tidwall/gjson"

	"github.com/ory/x/configx"
	"github.com/ory/x/jsonx"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/tracing"

	kjson "github.com/knadh/koanf/parsers/json"
)

const (
	DefaultIdentityTraitsSchemaID                                   = "default"
	DefaultBrowserReturnURL                                         = "default_browser_return_url"
	DefaultSQLiteMemoryDSN                                          = "sqlite://:memory:?_fk=true"
	UnknownVersion                                                  = "unknown version"
	ViperKeyDSN                                                     = "dsn"
	ViperKeyCourierSMTPURL                                          = "courier.smtp.connection_uri"
	ViperKeyCourierTemplatesPath                                    = "courier.template_override_path"
	ViperKeyCourierSMTPFrom                                         = "courier.smtp.from_address"
	ViperKeySecretsDefault                                          = "secrets.default"
	ViperKeySecretsCookie                                           = "secrets.cookie"
	ViperKeyPublicBaseURL                                           = "serve.public.base_url"
	ViperKeyPublicPort                                              = "serve.public.port"
	ViperKeyPublicHost                                              = "serve.public.host"
	ViperKeyAdminBaseURL                                            = "serve.admin.base_url"
	ViperKeyAdminPort                                               = "serve.admin.port"
	ViperKeyAdminHost                                               = "serve.admin.host"
	ViperKeySessionLifespan                                         = "session.lifespan"
	ViperKeySessionSameSite                                         = "session.cookie.same_site"
	ViperKeySessionDomain                                           = "session.cookie.domain"
	ViperKeySessionPath                                             = "session.cookie.path"
	ViperKeySessionPersistentCookie                                 = "session.cookie.persistent"
	ViperKeySelfServiceStrategyConfig                               = "selfservice.methods"
	ViperKeySelfServiceBrowserDefaultReturnTo                       = "selfservice." + DefaultBrowserReturnURL
	ViperKeyURLsWhitelistedReturnToDomains                          = "selfservice.whitelisted_return_urls"
	ViperKeySelfServiceRegistrationUI                               = "selfservice.flows.registration.ui_url"
	ViperKeySelfServiceRegistrationRequestLifespan                  = "selfservice.flows.registration.lifespan"
	ViperKeySelfServiceRegistrationAfter                            = "selfservice.flows.registration.after"
	ViperKeySelfServiceRegistrationBeforeHooks                      = "selfservice.flows.registration.before.hooks"
	ViperKeySelfServiceLoginUI                                      = "selfservice.flows.login.ui_url"
	ViperKeySelfServiceLoginRequestLifespan                         = "selfservice.flows.login.lifespan"
	ViperKeySelfServiceLoginAfter                                   = "selfservice.flows.login.after"
	ViperKeySelfServiceLoginBeforeHooks                             = "selfservice.flows.login.before.hooks"
	ViperKeySelfServiceErrorUI                                      = "selfservice.flows.error.ui_url"
	ViperKeySelfServiceLogoutBrowserDefaultReturnTo                 = "selfservice.flows.logout.after." + DefaultBrowserReturnURL
	ViperKeySelfServiceSettingsURL                                  = "selfservice.flows.settings.ui_url"
	ViperKeySelfServiceSettingsAfter                                = "selfservice.flows.settings.after"
	ViperKeySelfServiceSettingsRequestLifespan                      = "selfservice.flows.settings.lifespan"
	ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter        = "selfservice.flows.settings.privileged_session_max_age"
	ViperKeySelfServiceRecoveryEnabled                              = "selfservice.flows.recovery.enabled"
	ViperKeySelfServiceRecoveryUI                                   = "selfservice.flows.recovery.ui_url"
	ViperKeySelfServiceRecoveryRequestLifespan                      = "selfservice.flows.recovery.lifespan"
	ViperKeySelfServiceRecoveryBrowserDefaultReturnTo               = "selfservice.flows.recovery.after." + DefaultBrowserReturnURL
	ViperKeySelfServiceVerificationEnabled                          = "selfservice.flows.verification.enabled"
	ViperKeySelfServiceVerificationUI                               = "selfservice.flows.verification.ui_url"
	ViperKeySelfServiceVerificationRequestLifespan                  = "selfservice.flows.verification.lifespan"
	ViperKeySelfServiceVerificationBrowserDefaultReturnTo           = "selfservice.flows.verification.after." + DefaultBrowserReturnURL
	ViperKeyDefaultIdentitySchemaURL                                = "identity.default_schema_url"
	ViperKeyIdentitySchemas                                         = "identity.schemas"
	ViperKeyHasherArgon2ConfigMemory                                = "hashers.argon2.memory"
	ViperKeyHasherArgon2ConfigIterations                            = "hashers.argon2.iterations"
	ViperKeyHasherArgon2ConfigParallelism                           = "hashers.argon2.parallelism"
	ViperKeyHasherArgon2ConfigSaltLength                            = "hashers.argon2.salt_length"
	ViperKeyHasherArgon2ConfigKeyLength                             = "hashers.argon2.key_length"
	ViperKeyHasherArgon2ConfigMinimalDuration                       = "hashers.argon2.minimal_duration"
	ViperKeyHasherArgon2ConfigExpectedDeviation                     = "hashers.argon2.expected_deviation"
	ViperKeyHasherArgon2ConfigDedicatedMemory                       = "hashers.argon2.dedicated_memory"
	ViperKeyPasswordMaxBreaches                                     = "password.max_breaches"
	ViperKeyIgnoreNetworkErrors                                     = "password.ignore_network_errors"
	ViperKeyVersion                                                 = "version"
	Argon2DefaultMemory                                             = uint32(1 * bytesize.GB / bytesize.KB)
	Argon2DefaultIterations                                  uint32 = 4
	Argon2DefaultSaltLength                                  uint32 = 16
	Argon2DefaultKeyLength                                   uint32 = 32
	Argon2DefaultDuration                                           = 500 * time.Millisecond
	Argon2DefaultDeviation                                          = 500 * time.Millisecond
	Argon2DefaultDedicatedMemory                                    = 4 * bytesize.GB
)

type (
	HasherArgon2Config struct {
		Memory            uint32            `json:"memory"`
		Iterations        uint32            `json:"iterations"`
		Parallelism       uint8             `json:"parallelism"`
		SaltLength        uint32            `json:"salt_length"`
		KeyLength         uint32            `json:"key_length"`
		MinimalDuration   time.Duration     `json:"minimal_duration"`
		ExpectedDeviation time.Duration     `json:"expected_deviation"`
		DedicatedMemory   bytesize.ByteSize `json:"dedicated_memory"`
	}
	SelfServiceHook struct {
		Name   string          `json:"hook"`
		Config json.RawMessage `json:"config"`
	}
	SelfServiceStrategy struct {
		Enabled bool            `json:"enabled"`
		Config  json.RawMessage `json:"config"`
	}
	SchemaConfig struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	PasswordPolicyConfig struct {
		MaxBreaches         uint `json:"max_breaches"`
		IgnoreNetworkErrors bool `json:"ignore_network_errors"`
	}
	SchemaConfigs []SchemaConfig
	Provider      struct {
		l *logrusx.Logger
		p *configx.Provider
	}

	Providers interface {
		Configuration(ctx context.Context) *Provider
	}
)

func (c *HasherArgon2Config) MarshalJSON() ([]byte, error) {
	type encoded struct {
		Memory            uint32 `json:"memory"`
		Iterations        uint32 `json:"iterations"`
		Parallelism       uint8  `json:"parallelism"`
		SaltLength        uint32 `json:"salt_length"`
		KeyLength         uint32 `json:"key_length"`
		MinimalDuration   string `json:"minimal_duration"`
		ExpectedDeviation string `json:"expected_deviation"`
		DedicatedMemory   string `json:"dedicated_memory"`
	}

	return json.Marshal(&encoded{
		Memory:            c.Memory,
		Iterations:        c.Iterations,
		Parallelism:       c.Parallelism,
		SaltLength:        c.SaltLength,
		KeyLength:         c.KeyLength,
		MinimalDuration:   c.MinimalDuration.String(),
		ExpectedDeviation: c.ExpectedDeviation.String(),
		DedicatedMemory:   c.DedicatedMemory.String(),
	})
}

var Argon2DefaultParallelism = uint8(runtime.NumCPU() * 2)

func HookStrategyKey(key, strategy string) string {
	return fmt.Sprintf("%s.%s.hooks", key, strategy)
}

func (s SchemaConfigs) FindSchemaByID(id string) (*SchemaConfig, error) {
	for _, sc := range s {
		if sc.ID == id {
			return &sc, nil
		}
	}

	return nil, errors.Errorf("could not find schema with id \"%s\"", id)
}

func MustNew(l *logrusx.Logger, opts ...configx.OptionModifier) *Provider {
	p, err := New(l, opts...)
	if err != nil {
		l.WithError(err).Fatalf("Unable to load config.")
	}
	return p
}

func New(l *logrusx.Logger, opts ...configx.OptionModifier) (*Provider, error) {
	f, err := pkger.Open("github.com/ory/kratos:/.schema/config.schema.json")
	if err != nil {
		return nil, err
	}

	schema, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	opts = append([]configx.OptionModifier{
		configx.WithStderrValidationReporter(),
		configx.OmitKeysFromTracing("dsn", "secrets.default", "secrets.cookie", "client_secret"),
		configx.WithImmutables("serve", "profiling", "log"),
		configx.WithLogrusWatcher(l),
	}, opts...)

	p, err := configx.New(schema, opts...)
	if err != nil {
		return nil, err
	}

	l.UseConfig(p)
	return &Provider{l: l, p: p}, nil
}

func (p *Provider) Source() *configx.Provider {
	return p.p
}

func (p *Provider) CORS(iface string) (cors.Options, bool) {
	switch iface {
	case "admin":
		return p.cors("serve.admin")
	case "public":
		return p.cors("serve.public")
	default:
		panic(fmt.Sprintf("Received unexpected CORS interface: %s", iface))
	}
}

func (p *Provider) cors(prefix string) (cors.Options, bool) {
	return p.p.CORS(prefix, cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Cookie"},
		ExposedHeaders:   []string{"Content-Type", "Set-Cookie"},
		AllowCredentials: true,
	})
}

func (p *Provider) Set(key string, value interface{}) error {
	return p.p.Set(key, value)
}

func (p *Provider) MustSet(key string, value interface{}) {
	if err := p.p.Set(key, value); err != nil {
		p.l.WithError(err).Fatalf("Unable to set \"%s\" to \"%s\".", key, value)
	}
}

func (p *Provider) SessionDomain() string {
	return p.p.String(ViperKeySessionDomain)
}

func (p *Provider) SessionPath() string {
	return p.p.String(ViperKeySessionPath)
}

func (p *Provider) HasherArgon2() (*HasherArgon2Config, error) {
	// warn about usage of default values and point to the docs
	// warning will require https://github.com/ory/viper/issues/19
	c := &HasherArgon2Config{
		Memory:            uint32(p.p.IntF(ViperKeyHasherArgon2ConfigMemory, int(Argon2DefaultMemory))),
		Iterations:        uint32(p.p.IntF(ViperKeyHasherArgon2ConfigIterations, int(Argon2DefaultIterations))),
		Parallelism:       uint8(p.p.IntF(ViperKeyHasherArgon2ConfigParallelism, int(Argon2DefaultParallelism))),
		SaltLength:        uint32(p.p.IntF(ViperKeyHasherArgon2ConfigSaltLength, int(Argon2DefaultSaltLength))),
		KeyLength:         uint32(p.p.IntF(ViperKeyHasherArgon2ConfigKeyLength, int(Argon2DefaultKeyLength))),
		MinimalDuration:   p.p.DurationF(ViperKeyHasherArgon2ConfigMinimalDuration, Argon2DefaultDuration),
		ExpectedDeviation: p.p.DurationF(ViperKeyHasherArgon2ConfigExpectedDeviation, Argon2DefaultDeviation),
		DedicatedMemory:   Argon2DefaultDedicatedMemory,
	}

	switch m := p.p.Get(ViperKeyHasherArgon2ConfigDedicatedMemory).(type) {
	case string:
		if err := c.DedicatedMemory.Set(m); err != nil {
			return nil, errors.WithStack(err)
		}
	case bytesize.ByteSize:
		c.DedicatedMemory = m
	}

	return c, nil
}

func (p *Provider) listenOn(key string) string {
	fb := 4433
	if key == "admin" {
		fb = 4434
	}

	port := p.p.IntF("serve."+key+".port", fb)
	if port < 1 {
		p.l.Fatalf("serve.%s.port can not be zero or negative", key)
	}

	return fmt.Sprintf("%s:%d", p.p.String("serve."+key+".host"), port)
}

func (p *Provider) DefaultIdentityTraitsSchemaURL() *url.URL {
	return p.parseURIOrFail(ViperKeyDefaultIdentitySchemaURL)
}

func (p *Provider) IdentityTraitsSchemas() SchemaConfigs {
	ds := SchemaConfig{
		ID:  DefaultIdentityTraitsSchemaID,
		URL: p.DefaultIdentityTraitsSchemaURL().String(),
	}

	if !p.p.Exists(ViperKeyIdentitySchemas) {
		return SchemaConfigs{ds}
	}

	var ss SchemaConfigs
	out, err := p.p.Marshal(kjson.Parser())
	if err != nil {
		p.l.WithError(err).Fatalf("Unable to dencode values from %s.", ViperKeyIdentitySchemas)
		return SchemaConfigs{ds}
	}

	config := gjson.GetBytes(out, ViperKeyIdentitySchemas).Raw
	if len(config) == 0 {
		return SchemaConfigs{ds}
	}

	if err := json.NewDecoder(bytes.NewBufferString(config)).Decode(&ss); err != nil {
		p.l.WithError(err).Fatalf("Unable to encode values from %s.", ViperKeyIdentitySchemas)
		return SchemaConfigs{ds}
	}

	return append(ss, ds)
}

func (p *Provider) AdminListenOn() string {
	return p.listenOn("admin")
}

func (p *Provider) PublicListenOn() string {
	return p.listenOn("public")
}

func (p *Provider) DSN() string {
	dsn := p.p.String(ViperKeyDSN)

	if dsn == "memory" {
		return DefaultSQLiteMemoryDSN
	}

	if len(dsn) > 0 {
		return dsn
	}

	p.l.Fatal("dsn must be set")
	return ""
}

func (p *Provider) DisableAPIFlowEnforcement() bool {
	if p.IsInsecureDevMode() && os.Getenv("DEV_DISABLE_API_FLOW_ENFORCEMENT") == "true" {
		p.l.Warn("Because \"DEV_DISABLE_API_FLOW_ENFORCEMENT=true\" and the \"--dev\" flag are set, self-service API flows will no longer check if the interaction is actually a browser flow. This is very dangerous as it allows bypassing of anti-CSRF measures, leaving the deployment highly vulnerable. This option should only be used for automated testing and never come close to real user data anywhere.")
		return true
	}
	return false
}

func (p *Provider) SelfServiceFlowVerificationEnabled() bool {
	return p.p.Bool(ViperKeySelfServiceVerificationEnabled)
}

func (p *Provider) SelfServiceFlowRecoveryEnabled() bool {
	return p.p.Bool(ViperKeySelfServiceRecoveryEnabled)
}

func (p *Provider) SelfServiceFlowLoginBeforeHooks() []SelfServiceHook {
	return p.selfServiceHooks(ViperKeySelfServiceLoginBeforeHooks)
}

func (p *Provider) SelfServiceFlowRegistrationBeforeHooks() []SelfServiceHook {
	return p.selfServiceHooks(ViperKeySelfServiceRegistrationBeforeHooks)
}

func (p *Provider) selfServiceHooks(key string) []SelfServiceHook {
	var hooks []SelfServiceHook
	if !p.p.Exists(key) {
		return []SelfServiceHook{}
	}

	out, err := p.p.Marshal(kjson.Parser())
	if err != nil {
		p.l.WithError(err).Fatalf("Unable to decode values from configuration key: %s", key)
	}

	config := gjson.GetBytes(out, key).Raw
	if len(config) == 0 {
		return []SelfServiceHook{}
	}

	if err := jsonx.NewStrictDecoder(bytes.NewBufferString(config)).Decode(&hooks); err != nil {
		p.l.WithError(err).Fatalf("Unable to encode value \"%s\" from configuration key: %s", config, key)
	}

	for k := range hooks {
		if len(hooks[k].Config) == 0 {
			hooks[k].Config = json.RawMessage("{}")
		}
	}

	return hooks
}

func (p *Provider) SelfServiceFlowLoginAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(HookStrategyKey(ViperKeySelfServiceLoginAfter, strategy))
}

func (p *Provider) SelfServiceFlowSettingsAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(HookStrategyKey(ViperKeySelfServiceSettingsAfter, strategy))
}

func (p *Provider) SelfServiceFlowRegistrationAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(HookStrategyKey(ViperKeySelfServiceRegistrationAfter, strategy))
}

func (p *Provider) SelfServiceStrategy(strategy string) *SelfServiceStrategy {
	config := "{}"
	out, err := p.p.Marshal(kjson.Parser())
	if err != nil {
		p.l.WithError(err).Warn("Unable to marshal self service strategy configuration.")
	} else if c := gjson.GetBytes(out,
		fmt.Sprintf("%s.%s.config", ViperKeySelfServiceStrategyConfig, strategy)).Raw; len(c) > 0 {
		config = c
	}

	enabledKey := fmt.Sprintf("%s.%s.enabled", ViperKeySelfServiceStrategyConfig, strategy)
	s := &SelfServiceStrategy{
		Enabled: p.p.Bool(enabledKey),
		Config:  json.RawMessage(config),
	}

	// The default value can easily be overwritten by setting e.g. `{"selfservice": "null"}` which means that
	// we need to forcibly set these values here:
	if !p.p.Exists(enabledKey) {
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

	return s
}

func (p *Provider) SecretsDefault() [][]byte {
	secrets := p.p.Strings(ViperKeySecretsDefault)

	if len(secrets) == 0 {
		secrets = []string{uuid.New().String()}
		p.MustSet(ViperKeySecretsDefault, secrets)
	}

	result := make([][]byte, len(secrets))
	for k, v := range secrets {
		result[k] = []byte(v)
	}

	return result
}

func (p *Provider) SecretsSession() [][]byte {
	secrets := p.p.Strings(ViperKeySecretsCookie)
	if len(secrets) == 0 {
		return p.SecretsDefault()
	}

	result := make([][]byte, len(secrets))
	for k, v := range secrets {
		result[k] = []byte(v)
	}

	return result
}

func (p *Provider) SelfServiceBrowserDefaultReturnTo() *url.URL {
	return p.parseURIOrFail(ViperKeySelfServiceBrowserDefaultReturnTo)
}

func (p *Provider) guessBaseURL(keyHost, keyPort string, defaultPort int) *url.URL {
	port := p.p.IntF(keyPort, defaultPort)

	host := p.p.String(keyHost)
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

func (p *Provider) baseURL(keyURL, keyHost, keyPort string, defaultPort int) *url.URL {
	switch t := p.p.Get(keyURL).(type) {
	case *url.URL:
		return t
	case url.URL:
		return &t
	case string:
		parsed, err := url.ParseRequestURI(t)
		if err != nil {
			p.l.WithError(err).Errorf("Configuration key %s is not a valid URL. Falling back to optimistically guessing the server's base URL. Please set a value to avoid problems with redirects and cookies.", ViperKeyPublicBaseURL)
			return p.guessBaseURL(keyHost, keyPort, defaultPort)
		}
		return parsed
	}

	p.l.Warnf("Configuration key %s was left empty. Optimistically guessing the server's base URL. Please set a value to avoid problems with redirects and cookies.", ViperKeyPublicBaseURL)
	return p.guessBaseURL(keyHost, keyPort, defaultPort)
}

func (p *Provider) SelfPublicURL() *url.URL {
	return p.baseURL(ViperKeyPublicBaseURL, ViperKeyPublicHost, ViperKeyPublicPort, 4433)
}

func (p *Provider) SelfAdminURL() *url.URL {
	return p.baseURL(ViperKeyAdminBaseURL, ViperKeyAdminHost, ViperKeyAdminPort, 4434)
}

func (p *Provider) CourierSMTPURL() *url.URL {
	return p.parseURIOrFail(ViperKeyCourierSMTPURL)
}

func (p *Provider) SelfServiceFlowLoginUI() *url.URL {
	return p.parseURIOrFail(ViperKeySelfServiceLoginUI)
}

func (p *Provider) SelfServiceFlowSettingsUI() *url.URL {
	return p.parseURIOrFail(ViperKeySelfServiceSettingsURL)
}

func (p *Provider) SelfServiceFlowErrorURL() *url.URL {
	return p.parseURIOrFail(ViperKeySelfServiceErrorUI)
}

func (p *Provider) SelfServiceFlowRegistrationUI() *url.URL {
	return p.parseURIOrFail(ViperKeySelfServiceRegistrationUI)
}

func (p *Provider) SelfServiceFlowRecoveryUI() *url.URL {
	return p.parseURIOrFail(ViperKeySelfServiceRecoveryUI)
}

// SessionLifespan returns nil when the value is not set.
func (p *Provider) SessionLifespan() time.Duration {
	return p.p.DurationF(ViperKeySessionLifespan, time.Hour*24)
}

func (p *Provider) SessionPersistentCookie() bool {
	return p.p.Bool(ViperKeySessionPersistentCookie)
}

func (p *Provider) SelfServiceBrowserWhitelistedReturnToDomains() (us []url.URL) {
	src := p.p.Strings(ViperKeyURLsWhitelistedReturnToDomains)
	for k, u := range src {
		if len(u) == 0 {
			continue
		}

		parsed, err := url.ParseRequestURI(u)
		if err != nil {
			p.l.WithError(err).Warnf("Ignoring URL \"%s\" from configuration key \"%s.%d\".", u, ViperKeyURLsWhitelistedReturnToDomains, k)
			continue
		}

		us = append(us, *parsed)
	}

	return us
}

func (p *Provider) SelfServiceFlowLoginRequestLifespan() time.Duration {
	return p.p.DurationF(ViperKeySelfServiceLoginRequestLifespan, time.Hour)
}

func (p *Provider) SelfServiceFlowSettingsFlowLifespan() time.Duration {
	return p.p.DurationF(ViperKeySelfServiceSettingsRequestLifespan, time.Hour)
}

func (p *Provider) SelfServiceFlowRegistrationRequestLifespan() time.Duration {
	return p.p.DurationF(ViperKeySelfServiceRegistrationRequestLifespan, time.Hour)
}

func (p *Provider) SelfServiceFlowLogoutRedirectURL() *url.URL {
	return p.p.RequestURIF(ViperKeySelfServiceLogoutBrowserDefaultReturnTo, p.SelfServiceBrowserDefaultReturnTo())
}

func (p *Provider) CourierSMTPFrom() string {
	return p.p.StringF(ViperKeyCourierSMTPFrom, "noreply@kratos.ory.sh")
}

func (p *Provider) CourierTemplatesRoot() string {
	return p.p.StringF(ViperKeyCourierTemplatesPath, "/courier/template/templates")
}

func (p *Provider) parseURIOrFail(key string) *url.URL {
	u, err := url.ParseRequestURI(p.p.String(key))
	if err != nil {
		p.l.WithError(errors.WithStack(err)).
			Fatalf("Configuration value from key %s is not a valid URL: %s", key, p.p.String(key))
	}
	return u
}

func (p *Provider) Tracing() *tracing.Config {
	return p.p.TracingConfig("ORY Kratos")
}

func (p *Provider) IsInsecureDevMode() bool {
	return p.Source().Bool("dev")
}

func (p *Provider) SelfServiceFlowVerificationUI() *url.URL {
	return p.parseURIOrFail(ViperKeySelfServiceVerificationUI)
}

func (p *Provider) SelfServiceFlowVerificationRequestLifespan() time.Duration {
	return p.p.DurationF(ViperKeySelfServiceVerificationRequestLifespan, time.Hour)
}

func (p *Provider) SelfServiceFlowVerificationReturnTo(defaultReturnTo *url.URL) *url.URL {
	return p.p.RequestURIF(ViperKeySelfServiceVerificationBrowserDefaultReturnTo, defaultReturnTo)
}

func (p *Provider) SelfServiceFlowRecoveryReturnTo() *url.URL {
	return p.p.RequestURIF(ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, p.SelfServiceBrowserDefaultReturnTo())
}

func (p *Provider) SelfServiceFlowRecoveryRequestLifespan() time.Duration {
	return p.p.DurationF(ViperKeySelfServiceRecoveryRequestLifespan, time.Hour)
}

func (p *Provider) SelfServiceFlowSettingsPrivilegedSessionMaxAge() time.Duration {
	return p.p.DurationF(ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Hour)
}

func (p *Provider) SessionSameSiteMode() http.SameSite {
	switch p.p.StringF(ViperKeySessionSameSite, "Lax") {
	case "Lax":
		return http.SameSiteLaxMode
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	}
	return http.SameSiteDefaultMode
}

func (p *Provider) SelfServiceFlowLoginReturnTo(strategy string) *url.URL {
	return p.selfServiceReturnTo(ViperKeySelfServiceLoginAfter, strategy)
}

func (p *Provider) SelfServiceFlowRegistrationReturnTo(strategy string) *url.URL {
	return p.selfServiceReturnTo(ViperKeySelfServiceRegistrationAfter, strategy)
}

func (p *Provider) SelfServiceFlowSettingsReturnTo(strategy string, defaultReturnTo *url.URL) *url.URL {
	return p.p.RequestURIF(
		ViperKeySelfServiceSettingsAfter+"."+strategy+"."+DefaultBrowserReturnURL,
		p.p.RequestURIF(ViperKeySelfServiceSettingsAfter+"."+DefaultBrowserReturnURL,
			defaultReturnTo,
		),
	)
}

func (p *Provider) selfServiceReturnTo(key string, strategy string) *url.URL {
	return p.p.RequestURIF(
		key+"."+strategy+"."+DefaultBrowserReturnURL,
		p.p.RequestURIF(key+"."+DefaultBrowserReturnURL,
			p.SelfServiceBrowserDefaultReturnTo(),
		),
	)
}

func (p *Provider) ConfigVersion() string {
	return p.p.StringF(ViperKeyVersion, UnknownVersion)
}

func (p *Provider) PasswordPolicyConfig() *PasswordPolicyConfig {
	return &PasswordPolicyConfig{
		MaxBreaches:         uint(p.p.Int(ViperKeyPasswordMaxBreaches)),
		IgnoreNetworkErrors: p.p.BoolF(ViperKeyIgnoreNetworkErrors, true),
	}
}
