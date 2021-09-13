module github.com/ory/kratos

go 1.16

replace (
	github.com/gobuffalo/pop/v5 => github.com/gobuffalo/pop/v5 v5.3.4-0.20210608105745-bb07a373cc0e
	github.com/luna-duclos/instrumentedsql => github.com/ory/instrumentedsql v1.2.0
	github.com/luna-duclos/instrumentedsql/opentracing => github.com/ory/instrumentedsql/opentracing v0.0.0-20210903114257-c8963b546c5c
	github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v1.14.7-0.20210414154423-1157a4212dcb
	github.com/oleiade/reflections => github.com/oleiade/reflections v1.0.1
	// Use the internal httpclient which can be generated in this codebase but mark it as the
	// official SDK, allowing for the Ory CLI to consume Ory Kratos' CLI commands.
	github.com/ory/kratos-client-go => ./internal/httpclient
	github.com/ory/x => github.com/ory/x v0.0.272
	go.mongodb.org/mongo-driver => go.mongodb.org/mongo-driver v1.4.6
	gopkg.in/DataDog/dd-trace-go.v1 => gopkg.in/DataDog/dd-trace-go.v1 v1.27.1-0.20201005154917-54b73b3e126a
)

require (
	github.com/DataDog/datadog-go v4.8.2+incompatible // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/alecthomas/units v0.0.0-20210912230133-d1bdfacee922 // indirect
	github.com/arbovm/levenshtein v0.0.0-20160628152529-48b4e1c0c4d0
	github.com/armon/go-metrics v0.3.9 // indirect
	github.com/avast/retry-go/v3 v3.1.1
	github.com/bwmarrin/discordgo v0.23.2
	github.com/bxcodec/faker/v3 v3.6.0
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/containerd/containerd v1.5.2 // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/davidrjonas/semver-cli v0.0.0-20200305203455-0b3cebbaa360
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/docker/cli v20.10.8+incompatible // indirect
	github.com/docker/docker v20.10.8+incompatible // indirect
	github.com/elastic/go-sysinfo v1.7.0 // indirect
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/fatih/color v1.12.0
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/analysis v0.20.1 // indirect
	github.com/go-openapi/errors v0.20.1 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/runtime v0.19.31 // indirect
	github.com/go-openapi/strfmt v0.20.2
	github.com/go-playground/validator/v10 v10.9.0
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-swagger/go-swagger v0.27.0
	github.com/gobuffalo/envy v1.9.0 // indirect
	github.com/gobuffalo/fizz v1.13.1-0.20201104174146-3416f0e6618f
	github.com/gobuffalo/flect v0.2.3 // indirect
	github.com/gobuffalo/helpers v0.6.2 // indirect
	github.com/gobuffalo/here v0.6.2 // indirect
	github.com/gobuffalo/httptest v1.0.2
	github.com/gobuffalo/nulls v0.4.0 // indirect
	github.com/gobuffalo/plush/v4 v4.1.6 // indirect
	github.com/gobuffalo/pop/v5 v5.3.4
	github.com/gobuffalo/validate/v3 v3.3.0 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/golang/gddo v0.0.0-20210115222349-20d68f94ee1f
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/mock v1.5.0
	github.com/google/go-github/v27 v27.0.6
	github.com/google/go-jsonnet v0.17.0
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/sessions v1.2.1
	github.com/hashicorp/consul/api v1.10.1
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v0.16.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12
	github.com/inhies/go-bytesize v0.0.0-20210819104631-275770b98743
	github.com/instana/go-sensor v1.31.0 // indirect
	github.com/jackc/pgx/v4 v4.13.0 // indirect
	github.com/jandelgado/gcov2lcov v1.0.5 // indirect
	github.com/jmoiron/sqlx v1.3.4 // indirect
	github.com/jteeuwen/go-bindata v3.0.7+incompatible
	github.com/julienschmidt/httprouter v1.3.0
	github.com/knadh/koanf v1.2.3
	github.com/lib/pq v1.10.3 // indirect
	github.com/looplab/fsm v0.3.0 // indirect
	github.com/luna-duclos/instrumentedsql v1.1.3
	github.com/luna-duclos/instrumentedsql/opentracing v0.0.0-20201103091713-40d03108b6f4
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/goveralls v0.0.9
	github.com/microcosm-cc/bluemonday v1.0.15 // indirect
	github.com/mikefarah/yq v1.15.1-0.20191031234738-3c701fe98e3d
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/montanaflynn/stats v0.6.6
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/runc v1.0.2 // indirect
	github.com/openzipkin/zipkin-go v0.2.5 // indirect
	github.com/ory/analytics-go/v4 v4.0.2
	github.com/ory/dockertest/v3 v3.7.0
	github.com/ory/go-acc v0.2.6
	github.com/ory/go-convenience v0.1.0
	github.com/ory/graceful v0.1.1
	github.com/ory/herodot v0.9.7
	github.com/ory/jsonschema/v3 v3.0.3
	github.com/ory/kratos-client-go v0.7.6-alpha.1
	github.com/ory/mail/v3 v3.0.0
	github.com/ory/nosurf v1.2.5
	github.com/ory/x v0.0.280
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.6.0 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus/common v0.30.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rs/cors v1.8.0
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/slack-go/slack v0.9.4
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/sqs/goreturns v0.0.0-20181028201513-538ac6014518
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.9.0
	github.com/tidwall/sjson v1.2.1
	github.com/tinylib/msgp v1.1.6 // indirect
	github.com/uber/jaeger-client-go v2.29.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/urfave/negroni v1.0.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	go.elastic.co/apm/module/apmot v1.13.1 // indirect
	go.mongodb.org/mongo-driver v1.7.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.23.0 // indirect
	go.opentelemetry.io/otel/internal/metric v0.23.0 // indirect
	go.opentelemetry.io/otel/oteltest v1.0.0-RC2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/mod v0.5.0 // indirect
	golang.org/x/net v0.0.0-20210908191846-a5e095526f91 // indirect
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/tools v0.1.5
	google.golang.org/genproto v0.0.0-20210909211513-a8c4777a87af // indirect
	gopkg.in/DataDog/dd-trace-go.v1 v1.33.0 // indirect
	gopkg.in/imdario/mergo.v0 v0.3.9 // indirect
	gopkg.in/ini.v1 v1.63.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	howett.net/plist v0.0.0-20201203080718-1454fab16a06 // indirect
)
