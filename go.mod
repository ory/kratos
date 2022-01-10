module github.com/ory/kratos

go 1.16

replace (
	github.com/bradleyjkemp/cupaloy/v2 => github.com/aeneasr/cupaloy/v2 v2.6.1-0.20210924214125-3dfdd01210a3
	github.com/gorilla/sessions => github.com/ory/sessions v1.2.2-0.20220110165800-b09c17334dc2
	github.com/jackc/pgconn => github.com/jackc/pgconn v1.10.1-0.20211002123621-290ee79d1e8d
	github.com/knadh/koanf => github.com/aeneasr/koanf v0.14.1-0.20211230115640-aa3902b3267a
	github.com/luna-duclos/instrumentedsql => github.com/ory/instrumentedsql v1.2.0
	github.com/luna-duclos/instrumentedsql/opentracing => github.com/ory/instrumentedsql/opentracing v0.0.0-20210903114257-c8963b546c5c
	github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v1.14.7-0.20210414154423-1157a4212dcb
	github.com/oleiade/reflections => github.com/oleiade/reflections v1.0.1
	// Use the internal httpclient which can be generated in this codebase but mark it as the
	// official SDK, allowing for the Ory CLI to consume Ory Kratos' CLI commands.
	github.com/ory/kratos-client-go => ./internal/httpclient
	go.mongodb.org/mongo-driver => go.mongodb.org/mongo-driver v1.4.6
	golang.org/x/sys => golang.org/x/sys v0.0.0-20211007075335-d3039528d8ac
	gopkg.in/DataDog/dd-trace-go.v1 => gopkg.in/DataDog/dd-trace-go.v1 v1.27.1-0.20201005154917-54b73b3e126a
)

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig/v3 v3.0.0
	github.com/arbovm/levenshtein v0.0.0-20160628152529-48b4e1c0c4d0
	github.com/avast/retry-go/v3 v3.1.1
	github.com/bradleyjkemp/cupaloy/v2 v2.6.0
	github.com/bwmarrin/discordgo v0.23.0
	github.com/bxcodec/faker/v3 v3.3.1
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/cortesi/modd v0.0.0-20210323234521-b35eddab86cc
	github.com/davecgh/go-spew v1.1.1
	github.com/davidrjonas/semver-cli v0.0.0-20190116233701-ee19a9a0dda6
	github.com/duo-labs/webauthn v0.0.0-20210727191636-9f1b88ef44cc
	github.com/fatih/color v1.13.0
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/strfmt v0.20.3
	github.com/go-playground/validator/v10 v10.4.1
	github.com/go-swagger/go-swagger v0.26.1
	github.com/gobuffalo/fizz v1.14.0
	github.com/gobuffalo/httptest v1.0.2
	github.com/gobuffalo/pop/v6 v6.0.1
	github.com/gofrs/uuid v4.1.0+incompatible
	github.com/golang-jwt/jwt/v4 v4.1.0
	github.com/golang/gddo v0.0.0-20190904175337-72a348e765d2
	github.com/golang/mock v1.6.0
	github.com/google/go-github/v27 v27.0.1
	github.com/google/go-github/v38 v38.1.0
	github.com/google/go-jsonnet v0.17.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/sessions v1.2.1
	github.com/gtank/cryptopasta v0.0.0-20170601214702-1f550f6f2f69
	github.com/hashicorp/consul/api v1.5.0
	github.com/hashicorp/go-retryablehttp v0.7.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/imdario/mergo v0.3.12
	github.com/inhies/go-bytesize v0.0.0-20210819104631-275770b98743
	github.com/jteeuwen/go-bindata v3.0.7+incompatible
	github.com/julienschmidt/httprouter v1.3.0
	github.com/knadh/koanf v1.3.3
	github.com/luna-duclos/instrumentedsql v1.1.3
	github.com/luna-duclos/instrumentedsql/opentracing v0.0.0-20201103091713-40d03108b6f4
	github.com/mattn/goveralls v0.0.7
	github.com/mikefarah/yq v1.15.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe
	github.com/ory/analytics-go/v4 v4.0.2
	github.com/ory/dockertest/v3 v3.8.1
	github.com/ory/go-acc v0.2.6
	github.com/ory/go-convenience v0.1.0
	github.com/ory/graceful v0.1.1
	github.com/ory/herodot v0.9.12
	github.com/ory/jsonschema/v3 v3.0.5-0.20211222152031-b530fb44a010
	github.com/ory/kratos-client-go v0.6.3-alpha.1
	github.com/ory/mail/v3 v3.0.0
	github.com/ory/nosurf v1.2.7
	github.com/ory/x v0.0.330
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.3.0
	github.com/rs/cors v1.8.0
	github.com/sirupsen/logrus v1.8.1
	github.com/slack-go/slack v0.7.4
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/sqs/goreturns v0.0.0-20181028201513-538ac6014518
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.9.4
	github.com/tidwall/sjson v1.2.2
	github.com/urfave/negroni v1.0.0
	github.com/zmb3/spotify/v2 v2.0.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/net v0.0.0-20211020060615-d418f374d309
	golang.org/x/oauth2 v0.0.0-20210810183815-faf39c7919d5
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/tools v0.1.7
)
