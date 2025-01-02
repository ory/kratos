module github.com/ory/kratos

go 1.23

toolchain go1.23.2

replace (
	github.com/coreos/go-oidc/v3 => github.com/ory/go-oidc/v3 v3.0.0-20241127113405-e5362711266b

	github.com/go-swagger/go-swagger => github.com/aeneasr/go-swagger v0.19.1-0.20241013070044-bccef3a12e26 // See https://github.com/go-swagger/go-swagger/issues/3131
	// github.com/go-swagger/go-swagger => ../../go-swagger/go-swagger
	// https://github.com/gobuffalo/pop/pull/833
	github.com/gobuffalo/pop/v6 => github.com/ory/pop/v6 v6.2.1-0.20241121111754-e5dfc0f3344b

	github.com/gorilla/sessions => github.com/ory/sessions v1.2.2-0.20220110165800-b09c17334dc2
	github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v1.14.22

	// Use the internal httpclient which can be generated in this codebase but mark it as the
	// official SDK, allowing for the Ory CLI to consume Ory Kratos' CLI commands.
	github.com/ory/client-go => ./internal/client-go
)

require (
	dario.cat/mergo v1.0.0
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/arbovm/levenshtein v0.0.0-20160628152529-48b4e1c0c4d0
	github.com/avast/retry-go/v3 v3.1.1
	github.com/bradleyjkemp/cupaloy/v2 v2.8.0
	github.com/bwmarrin/discordgo v0.28.1
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/coreos/go-oidc/v3 v3.11.0
	github.com/cortesi/modd v0.8.1
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/dghubble/oauth1 v0.7.3
	github.com/dgraph-io/ristretto v1.0.0
	github.com/fatih/color v1.17.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-crypt/crypt v0.2.25
	github.com/go-faker/faker/v4 v4.4.2
	github.com/go-openapi/strfmt v0.23.0
	github.com/go-playground/validator/v10 v10.22.0
	github.com/go-swagger/go-swagger v0.31.0
	github.com/go-webauthn/webauthn v0.11.2
	github.com/gobuffalo/httptest v1.5.2
	github.com/gobuffalo/pop/v6 v6.1.2-0.20230318123913-c85387acc9a0
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/golang-jwt/jwt/v4 v4.5.1
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/golang/gddo v0.0.0-20190904175337-72a348e765d2
	github.com/golang/mock v1.6.0
	github.com/google/go-github/v38 v38.1.0
	github.com/google/go-jsonnet v0.20.0
	github.com/gorilla/sessions v1.3.0
	github.com/gtank/cryptopasta v0.0.0-20170601214702-1f550f6f2f69
	github.com/hashicorp/go-retryablehttp v0.7.7
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/inhies/go-bytesize v0.0.0-20220417184213-4913239db9cf
	github.com/jarcoal/httpmock v1.3.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/knadh/koanf/parsers/json v0.1.0
	github.com/laher/mergefs v0.1.2-0.20230223191438-d16611b2f4e7 // indirect
	github.com/lestrrat-go/jwx/v2 v2.1.1
	github.com/luna-duclos/instrumentedsql v1.1.3
	github.com/mailhog/MailHog v1.0.1
	github.com/mattn/goveralls v0.0.12
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe
	github.com/ory/analytics-go/v5 v5.0.1
	github.com/ory/client-go v1.14.3
	github.com/ory/dockertest/v3 v3.11.0
	github.com/ory/go-acc v0.2.9-0.20230103102148-6b1c9a70dbbe
	github.com/ory/graceful v0.1.4-0.20230301144740-e222150c51d0
	github.com/ory/herodot v0.10.3-0.20230626083119-d7e5192f0d88
	github.com/ory/hydra-client-go/v2 v2.2.1
	github.com/ory/jsonschema/v3 v3.0.8
	github.com/ory/mail/v3 v3.0.0
	github.com/ory/nosurf v1.2.7
	github.com/ory/x v0.0.675
	github.com/peterhellberg/link v1.2.0
	github.com/phayes/freeport v0.0.0-20220201140144-74d24b5ae9f5
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.4.0
	github.com/rakutentech/jwk-go v1.1.3
	github.com/rs/cors v1.11.0
	github.com/samber/lo v1.46.0
	github.com/sirupsen/logrus v1.9.3
	github.com/slack-go/slack v0.13.1
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/gjson v1.17.3
	github.com/tidwall/sjson v1.2.5
	github.com/urfave/negroni v1.0.0
	github.com/zmb3/spotify/v2 v2.4.2
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.57.0
	go.opentelemetry.io/otel v1.32.0
	go.opentelemetry.io/otel/sdk v1.32.0
	go.opentelemetry.io/otel/trace v1.32.0
	golang.org/x/crypto v0.31.0
	golang.org/x/exp v0.0.0-20241217172543-b2144cdd0a67 // indirect
	golang.org/x/net v0.33.0
	golang.org/x/oauth2 v0.24.0
	golang.org/x/sync v0.10.0
	golang.org/x/text v0.21.0
	google.golang.org/grpc v1.67.1
)

require github.com/wI2L/jsondiff v0.6.0

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/cortesi/moddwatch v0.1.0 // indirect
	github.com/cortesi/termlog v0.0.0-20210222042314-a1eec763abec // indirect
	github.com/dgraph-io/ristretto/v2 v2.0.0 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/rjeczalik/notify v0.9.3 // indirect
	golang.org/x/term v0.27.0 // indirect
	golang.org/x/time v0.8.0 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6 // indirect
	mvdan.cc/sh/v3 v3.6.0 // indirect
)

require (
	code.dny.dev/ssrf v0.2.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/avast/retry-go/v4 v4.3.0 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/cockroach-go/v2 v2.3.5
	github.com/containerd/continuity v0.4.3 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/cli v26.1.4+incompatible // indirect
	github.com/docker/docker v27.1.1+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/fgprof v0.9.3 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-crypt/x v0.2.18 // indirect
	github.com/go-jose/go-jose/v3 v3.0.3 // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.23.0 // indirect
	github.com/go-openapi/errors v0.22.0 // indirect
	github.com/go-openapi/inflect v0.21.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/loads v0.22.0 // indirect
	github.com/go-openapi/runtime v0.28.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-openapi/validate v0.24.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/go-webauthn/x v0.1.14 // indirect
	github.com/gobuffalo/envy v1.10.2 // indirect
	github.com/gobuffalo/fizz v1.14.4 // indirect
	github.com/gobuffalo/flect v1.0.2 // indirect
	github.com/gobuffalo/github_flavored_markdown v1.1.4 // indirect
	github.com/gobuffalo/helpers v0.6.7 // indirect
	github.com/gobuffalo/nulls v0.4.2 // indirect
	github.com/gobuffalo/plush/v4 v4.1.21 // indirect
	github.com/gobuffalo/tags/v3 v3.1.4 // indirect
	github.com/gobuffalo/validate/v3 v3.3.3 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/goccy/go-yaml v1.11.3 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/go-tpm v0.9.1 // indirect
	github.com/google/pprof v0.0.0-20221010195024-131d412537ea // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/context v1.1.2 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/gorilla/handlers v1.5.2 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/pat v1.0.2 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.23.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/ian-kent/envconf v0.0.0-20141026121121-c19809918c02 // indirect
	github.com/ian-kent/go-log v0.0.0-20160113211217-5731446c36ab // indirect
	github.com/ian-kent/goose v0.0.0-20141221090059-c3541ea826ad // indirect
	github.com/ian-kent/linkio v0.0.0-20170807205755-97566b872887 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jandelgado/gcov2lcov v1.0.5 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/parsers/toml v0.1.0 // indirect
	github.com/knadh/koanf/parsers/yaml v0.1.0 // indirect
	github.com/knadh/koanf/providers/posflag v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.0.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.6 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/jwx v1.2.29
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailhog/MailHog-Server v1.0.1 // indirect
	github.com/mailhog/MailHog-UI v1.0.1 // indirect
	github.com/mailhog/data v1.0.1 // indirect
	github.com/mailhog/http v1.0.1 // indirect
	github.com/mailhog/mhsendmail v0.2.0 // indirect
	github.com/mailhog/smtp v1.0.1 // indirect
	github.com/mailhog/storage v1.0.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/microcosm-cc/bluemonday v1.0.26 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/nyaruka/phonenumbers v1.4.1
	github.com/ogier/pflag v0.0.1 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/opencontainers/runc v1.1.14 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/philhofer/fwd v1.1.3-0.20240612014219-fbbf4953d986 // indirect
	github.com/pkg/profile v1.7.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/seatgeek/logrus-gelf-formatter v0.0.0-20210414080842-5b05eb8ff761 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/segmentio/backo-go v1.0.1 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/sourcegraph/annotate v0.0.0-20160123013949-f4cad6c6324d // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/sourcegraph/syntaxhighlight v0.0.0-20170531221838-bd320f5d308e // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/viper v1.18.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/t-k/fluent-logger-golang v1.0.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tinylib/msgp v1.2.0 // indirect
	github.com/toqueteos/webbrowser v1.2.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	go.mongodb.org/mongo-driver v1.14.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.57.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.32.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.32.0 // indirect
	go.opentelemetry.io/contrib/samplers/jaegerremote v0.26.0 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.32.0 // indirect; / indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.32.0 // indirect; / indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.32.0 // indirect; / indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/tools v0.28.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241104194629-dd2ea8efbc28 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241104194629-dd2ea8efbc28 // indirect
	google.golang.org/protobuf v1.35.1
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
