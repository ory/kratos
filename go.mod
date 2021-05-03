module github.com/ory/kratos

go 1.16

replace gopkg.in/DataDog/dd-trace-go.v1 => gopkg.in/DataDog/dd-trace-go.v1 v1.27.1-0.20201005154917-54b73b3e126a

// Use the internal httpclient which can be generated in this codebase but mark it as the
// official SDK, allowing for the Ory CLI to consume Ory Kratos' CLI commands.
replace github.com/ory/kratos-client-go => ./internal/httpclient

// Use the internal name for tablename generation
replace github.com/ory/kratos/corp => ./corp

replace github.com/ory/cli => github.com/ory/cli v0.0.45-0.20210312151016-84cf0bd14396

replace go.mongodb.org/mongo-driver => go.mongodb.org/mongo-driver v1.4.6

replace github.com/gobuffalo/pop/v5 => github.com/gobuffalo/pop/v5 v5.3.2-0.20210412125924-f5ad9021d6ac

replace github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v1.14.7

require (
	github.com/DataDog/datadog-go v4.5.1+incompatible // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/Masterminds/sprig/v3 v3.0.0
	github.com/arbovm/levenshtein v0.0.0-20160628152529-48b4e1c0c4d0
	github.com/bwmarrin/discordgo v0.23.0
	github.com/bxcodec/faker/v3 v3.3.1
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/containerd/containerd v1.4.4 // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/davidrjonas/semver-cli v0.0.0-20190116233701-ee19a9a0dda6
	github.com/fatih/color v1.9.0
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/strfmt v0.20.0
	github.com/go-playground/validator/v10 v10.4.1
	github.com/go-swagger/go-swagger v0.26.1
	github.com/gobuffalo/fizz v1.13.1-0.20201104174146-3416f0e6618f
	github.com/gobuffalo/httptest v1.0.2
	github.com/gobuffalo/pop/v5 v5.3.3
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang/gddo v0.0.0-20190904175337-72a348e765d2
	github.com/golang/mock v1.5.0
	github.com/google/go-github/v27 v27.0.1
	github.com/google/go-jsonnet v0.16.0
	github.com/google/uuid v1.2.0
	github.com/gorilla/context v1.1.1
	github.com/gorilla/sessions v1.1.3
	github.com/hashicorp/consul/api v1.5.0
	github.com/hashicorp/go-retryablehttp v0.6.8
	github.com/hashicorp/golang-lru v0.5.4
	github.com/imdario/mergo v0.3.7
	github.com/inhies/go-bytesize v0.0.0-20201103132853-d0aed0d254f8
	github.com/jteeuwen/go-bindata v3.0.7+incompatible
	github.com/julienschmidt/httprouter v1.3.0
	github.com/knadh/koanf v0.14.1-0.20201201075439-e0853799f9ec
	github.com/luna-duclos/instrumentedsql v1.1.3
	github.com/luna-duclos/instrumentedsql/opentracing v0.0.0-20201103091713-40d03108b6f4
	github.com/mattn/goveralls v0.0.7
	github.com/mikefarah/yq v1.15.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7 // indirect
	github.com/ory/analytics-go/v4 v4.0.0
	github.com/ory/cli v0.0.49
	github.com/ory/dockertest/v3 v3.6.5
	github.com/ory/go-acc v0.2.6
	github.com/ory/go-convenience v0.1.0
	github.com/ory/graceful v0.1.1
	github.com/ory/herodot v0.9.5
	github.com/ory/jsonschema/v3 v3.0.3
	github.com/ory/kratos-client-go v0.5.4-alpha.1.0.20210308170950-06c2c1c071a8
	github.com/ory/kratos/corp v0.0.0-00010101000000-000000000000
	github.com/ory/mail/v3 v3.0.0
	github.com/ory/nosurf v1.2.4
	github.com/ory/x v0.0.231
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.5.0 // indirect
	github.com/prometheus/client_golang v1.4.0
	github.com/prometheus/common v0.9.1
	github.com/rs/cors v1.6.0
	github.com/sirupsen/logrus v1.8.1
	github.com/slack-go/slack v0.7.4
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/sqs/goreturns v0.0.0-20181028201513-538ac6014518
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.7.1
	github.com/tidwall/sjson v1.1.5
	github.com/uber/jaeger-lib v2.4.0+incompatible // indirect
	github.com/urfave/negroni v1.0.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/tools v0.1.0
)
