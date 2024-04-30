module github.com/ory/kratos

go 1.21

replace (
	github.com/go-sql-driver/mysql => github.com/go-sql-driver/mysql v1.7.2-0.20231005084435-37980127edfb

	// https://github.com/gobuffalo/pop/pull/833
	github.com/gobuffalo/pop/v6 => github.com/alnr/pop/v6 v6.1.2-0.20240220141536-653aad67c0c2

	github.com/gorilla/sessions => github.com/ory/sessions v1.2.2-0.20220110165800-b09c17334dc2
	github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v1.14.16

	// Use the internal httpclient which can be generated in this codebase but mark it as the
	// official SDK, allowing for the Ory CLI to consume Ory Kratos' CLI commands.
	github.com/ory/client-go => ./internal/client-go
)

require (
	code.dny.dev/ssrf v0.2.0 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/arbovm/levenshtein v0.0.0-20160628152529-48b4e1c0c4d0
	github.com/avast/retry-go/v3 v3.1.1
	github.com/bradleyjkemp/cupaloy/v2 v2.8.0
	github.com/bwmarrin/discordgo v0.23.0
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cortesi/modd v0.0.0-20210323234521-b35eddab86cc
	github.com/davecgh/go-spew v1.1.1
	github.com/davidrjonas/semver-cli v0.0.0-20190116233701-ee19a9a0dda6
	github.com/dgraph-io/ristretto v0.1.1
	github.com/fatih/color v1.13.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-crypt/crypt v0.2.9
	github.com/go-faker/faker/v4 v4.2.0
	github.com/go-openapi/strfmt v0.21.7
	github.com/go-playground/validator/v10 v10.4.1
	github.com/go-swagger/go-swagger v0.30.5
	github.com/go-webauthn/webauthn v0.8.4
	github.com/gobuffalo/fizz v1.14.4
	github.com/gobuffalo/httptest v1.5.2
	github.com/gobuffalo/pop/v6 v6.1.2-0.20230318123913-c85387acc9a0
	github.com/gofrs/uuid v4.3.1+incompatible
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/golang/gddo v0.0.0-20190904175337-72a348e765d2
	github.com/golang/mock v1.6.0
	github.com/google/go-github/v27 v27.0.1
	github.com/google/go-github/v38 v38.1.0
	github.com/google/go-jsonnet v0.20.0
	github.com/gorilla/sessions v1.2.1
	github.com/gtank/cryptopasta v0.0.0-20170601214702-1f550f6f2f69
	github.com/hashicorp/consul/api v1.20.0
	github.com/hashicorp/go-retryablehttp v0.7.2
	github.com/hashicorp/golang-lru v0.5.4
	github.com/imdario/mergo v0.3.13
	github.com/inhies/go-bytesize v0.0.0-20220417184213-4913239db9cf
	github.com/jarcoal/httpmock v1.0.5
	github.com/jmoiron/sqlx v1.3.5
	github.com/jteeuwen/go-bindata v3.0.7+incompatible
	github.com/julienschmidt/httprouter v1.3.0
	github.com/knadh/koanf/parsers/json v0.1.0
	github.com/laher/mergefs v0.1.2-0.20230223191438-d16611b2f4e7
	github.com/lestrrat-go/jwx v1.2.29 // indirect
	github.com/luna-duclos/instrumentedsql v1.1.3
	github.com/mailhog/MailHog v1.0.1
	github.com/mattn/goveralls v0.0.7
	github.com/mikefarah/yq/v4 v4.19.1
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe
	github.com/ory/analytics-go/v5 v5.0.1
	github.com/ory/client-go v0.2.0-alpha.60
	github.com/ory/dockertest/v3 v3.9.1
	github.com/ory/go-acc v0.2.9-0.20230103102148-6b1c9a70dbbe
	github.com/ory/graceful v0.1.4-0.20230301144740-e222150c51d0
	github.com/ory/herodot v0.10.3-0.20230626083119-d7e5192f0d88
	github.com/ory/hydra-client-go/v2 v2.2.0-rc.3.0.20240202131107-1c7b57df3bb0
	github.com/ory/jsonschema/v3 v3.0.8
	github.com/ory/mail/v3 v3.0.0
	github.com/ory/nosurf v1.2.7
	github.com/ory/x v0.0.623
	github.com/peterhellberg/link v1.2.0
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.4.0
	github.com/rakutentech/jwk-go v1.1.3
	github.com/rs/cors v1.8.2
	github.com/samber/lo v1.37.0
	github.com/sirupsen/logrus v1.9.0
	github.com/slack-go/slack v0.7.4
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/sqs/goreturns v0.0.0-20181028201513-538ac6014518
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/gjson v1.14.3
	github.com/tidwall/sjson v1.2.5
	github.com/urfave/negroni v1.0.0
	github.com/zmb3/spotify/v2 v2.4.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.47.0
	go.opentelemetry.io/otel v1.22.0
	go.opentelemetry.io/otel/sdk v1.21.0
	go.opentelemetry.io/otel/trace v1.22.0
	golang.org/x/crypto v0.21.0
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa
	golang.org/x/net v0.23.0
	golang.org/x/oauth2 v0.16.0
	golang.org/x/sync v0.5.0
	golang.org/x/text v0.14.0
	golang.org/x/tools/cmd/cover v0.1.0-deprecated
	google.golang.org/grpc v1.59.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/semver/v3 v3.2.0 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/a8m/envsubst v1.3.0 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15 // indirect
	github.com/armon/go-metrics v0.4.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/avast/retry-go/v4 v4.3.0 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/cockroach-go/v2 v2.3.5
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/cortesi/moddwatch v0.0.0-20210222043437-a6aaad86a36e // indirect
	github.com/cortesi/termlog v0.0.0-20210222042314-a1eec763abec // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/docker/cli v20.10.21+incompatible // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v20.10.24+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/elliotchance/orderedmap v1.4.0 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/fgprof v0.9.3 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/go-crypt/x v0.2.1 // indirect
	github.com/go-jose/go-jose/v3 v3.0.3 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.21.4 // indirect
	github.com/go-openapi/errors v0.20.4 // indirect
	github.com/go-openapi/inflect v0.19.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/loads v0.21.2 // indirect
	github.com/go-openapi/runtime v0.26.0 // indirect
	github.com/go-openapi/spec v0.20.9 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/go-openapi/validate v0.22.1 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/go-webauthn/x v0.1.4 // indirect
	github.com/gobuffalo/envy v1.10.2 // indirect
	github.com/gobuffalo/flect v1.0.0 // indirect
	github.com/gobuffalo/github_flavored_markdown v1.1.3 // indirect
	github.com/gobuffalo/helpers v0.6.7 // indirect
	github.com/gobuffalo/nulls v0.4.2 // indirect
	github.com/gobuffalo/plush/v4 v4.1.18 // indirect
	github.com/gobuffalo/tags/v3 v3.1.4 // indirect
	github.com/gobuffalo/validate/v3 v3.3.3 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/goccy/go-yaml v1.9.6 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.1.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/go-tpm v0.9.0 // indirect
	github.com/google/pprof v0.0.0-20221010195024-131d412537ea // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/gorilla/pat v1.0.1 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.2.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/ian-kent/envconf v0.0.0-20141026121121-c19809918c02 // indirect
	github.com/ian-kent/go-log v0.0.0-20160113211217-5731446c36ab // indirect
	github.com/ian-kent/goose v0.0.0-20141221090059-c3541ea826ad // indirect
	github.com/ian-kent/linkio v0.0.0-20170807205755-97566b872887 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jackc/pgx/v4 v4.18.2 // indirect
	github.com/jandelgado/gcov2lcov v1.0.5 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/jinzhu/copier v0.3.5 // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/parsers/toml v0.1.0 // indirect
	github.com/knadh/koanf/parsers/yaml v0.1.0 // indirect
	github.com/knadh/koanf/providers/posflag v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.0.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/lib/pq v1.10.7 // indirect
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
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/microcosm-cc/bluemonday v1.0.21 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/term v0.0.0-20220808134915-39b0c02b01ae // indirect
	github.com/nyaruka/phonenumbers v1.1.6 // indirect
	github.com/ogier/pflag v0.0.1 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/opencontainers/runc v1.1.12 // indirect
	github.com/openzipkin/zipkin-go v0.4.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/pkg/profile v1.7.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rjeczalik/notify v0.0.0-20181126183243-629144ba06a1 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/seatgeek/logrus-gelf-formatter v0.0.0-20210414080842-5b05eb8ff761 // indirect
	github.com/segmentio/backo-go v1.0.1 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/sourcegraph/annotate v0.0.0-20160123013949-f4cad6c6324d // indirect
	github.com/sourcegraph/syntaxhighlight v0.0.0-20170531221838-bd320f5d308e // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.16.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/t-k/fluent-logger-golang v1.0.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/timtadh/data-structures v0.5.3 // indirect
	github.com/timtadh/lexmachine v0.2.2 // indirect
	github.com/tinylib/msgp v1.1.8 // indirect
	github.com/toqueteos/webbrowser v1.2.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	go.mongodb.org/mongo-driver v1.11.3 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.47.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.21.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.21.1 // indirect
	go.opentelemetry.io/contrib/samplers/jaegerremote v0.15.1 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.21.0 // indirect; / indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.21.0 // indirect; / indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.21.0 // indirect; / indirect
	go.opentelemetry.io/otel/metric v1.22.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/term v0.18.0 // indirect
	golang.org/x/tools v0.15.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mvdan.cc/sh/v3 v3.3.0-0.dev.0.20210224101809-fb5052e7a010 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

require (
	github.com/coreos/go-oidc/v3 v3.9.0
	github.com/dghubble/oauth1 v0.7.2
	github.com/lestrrat-go/jwx/v2 v2.0.19
)

require (
	github.com/jackc/puddle/v2 v2.1.2 // indirect
	github.com/lestrrat-go/httprc v1.0.4 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
)
