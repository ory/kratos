module github.com/ory/kratos

go 1.13

replace github.com/santhosh-tekuri/jsonschema/v2 => github.com/ory/jsonschema/v2 v2.1.1-0.20191123130340-1c20114d2c04

replace github.com/ory/herodot => ../herodot

replace github.com/ory/x => ../x

require (
	github.com/bxcodec/faker v2.0.1+incompatible
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/coreos/go-oidc v2.0.0+incompatible
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/errors v0.18.0
	github.com/go-openapi/inflect v0.19.0 // indirect
	github.com/go-openapi/runtime v0.18.0
	github.com/go-openapi/strfmt v0.18.0
	github.com/go-openapi/swag v0.18.0
	github.com/go-openapi/validate v0.18.0
	github.com/go-playground/locales v0.12.1 // indirect
	github.com/go-playground/universal-translator v0.16.0 // indirect
	github.com/go-swagger/go-swagger v0.19.0
	github.com/go-swagger/scan-repo-boundary v0.0.0-20180623220736-973b3573c013 // indirect
	github.com/gobuffalo/httptest v1.0.2
	github.com/gobuffalo/packr v1.22.0
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/gobuffalo/pop v4.12.2+incompatible
	github.com/gobuffalo/tags v2.1.7+incompatible // indirect
	github.com/gobuffalo/uuid v2.0.5+incompatible
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang/gddo v0.0.0-20190904175337-72a348e765d2
	github.com/golang/mock v1.3.1
	github.com/google/go-github/v27 v27.0.1
	github.com/google/uuid v1.1.1
	github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers v1.4.1 // indirect
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.1.3
	github.com/imdario/mergo v0.3.7
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/jteeuwen/go-bindata v3.0.7+incompatible
	github.com/julienschmidt/httprouter v1.2.0
	github.com/justinas/nosurf v0.0.0-20190118163749-6453469bdcc9
	github.com/leodido/go-urn v1.1.0 // indirect
	github.com/luna-duclos/instrumentedsql v1.1.1 // indirect
	github.com/mattn/go-sqlite3 v1.13.0 // indirect
	github.com/mattn/goveralls v0.0.4
	github.com/mitchellh/go-homedir v1.1.0
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/ory/go-acc v0.0.0-20181118080137-ddc355013f90
	github.com/ory/gojsonschema v1.2.0
	github.com/ory/graceful v0.1.1
	github.com/ory/herodot v0.6.2
	github.com/ory/viper v1.5.6
	github.com/ory/x v0.0.83
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.8.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/rogpeppe/go-internal v1.5.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v2 v2.1.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.5.0
	github.com/sqs/goreturns v0.0.0-20181028201513-538ac6014518
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/gjson v1.3.2
	github.com/tidwall/sjson v1.0.4
	github.com/toqueteos/webbrowser v1.1.0 // indirect
	github.com/urfave/negroni v1.0.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	golang.org/x/crypto v0.0.0-20191122220453-ac88ee75c92c
	golang.org/x/net v0.0.0-20191125084936-ffdde1057850 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20191120155948-bd437916bb0e // indirect
	golang.org/x/tools v0.0.0-20191105231337-689d0f08e67a
	golang.org/x/xerrors v0.0.0-20191011141410-1b5146add898 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.28.0
)

replace github.com/gobuffalo/pop => ../../gobuffalo/pop
