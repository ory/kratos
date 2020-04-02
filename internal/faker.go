package internal

import (
	"math/rand"
	"net/http"
	"reflect"
	"time"

	"github.com/bxcodec/faker"

	"github.com/ory/x/randx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/x"
)

func RegisterFakes() {
	if err := faker.AddProvider("birthdate", func(v reflect.Value) (interface{}, error) {
		return time.Now().Add(time.Duration(rand.Int())).Round(time.Second).UTC(), nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("time_types", func(v reflect.Value) (interface{}, error) {
		es := make([]time.Time, rand.Intn(5))
		for k := range es {
			es[k] = time.Now().Add(time.Duration(rand.Int())).Round(time.Second).UTC()
		}
		return es, nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("http_header", func(v reflect.Value) (interface{}, error) {
		headers := http.Header{}
		for i := 0; i <= rand.Intn(5); i++ {
			values := make([]string, rand.Intn(4)+1)
			for k := range values {
				values[k] = randx.MustString(8, randx.AlphaNum)
			}
			headers[randx.MustString(8, randx.AlphaNum)] = values
		}

		return headers, nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("time_type", func(v reflect.Value) (interface{}, error) {
		return time.Now().Add(time.Duration(rand.Int())).Round(time.Second).UTC(), nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("login_request_methods", func(v reflect.Value) (interface{}, error) {
		var methods = make(map[identity.CredentialsType]*login.RequestMethod)
		for _, ct := range []identity.CredentialsType{identity.CredentialsTypePassword, identity.CredentialsTypeOIDC} {
			var f form.HTMLForm
			if err := faker.FakeData(&f); err != nil {
				return nil, err
			}
			methods[ct] = &login.RequestMethod{
				Method: ct,
				Config: &login.RequestMethodConfig{RequestMethodConfigurator: &f},
			}

		}
		return methods, nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("registration_request_methods", func(v reflect.Value) (interface{}, error) {
		var methods = make(map[identity.CredentialsType]*registration.RequestMethod)
		for _, ct := range []identity.CredentialsType{identity.CredentialsTypePassword, identity.CredentialsTypeOIDC} {
			var f form.HTMLForm
			if err := faker.FakeData(&f); err != nil {
				return nil, err
			}
			methods[ct] = &registration.RequestMethod{
				Method: ct,
				Config: &registration.RequestMethodConfig{RequestMethodConfigurator: &f},
			}
		}
		return methods, nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("settings_request_methods", func(v reflect.Value) (interface{}, error) {
		var methods = make(map[string]*settings.RequestMethod)
		for _, ct := range []string{settings.StrategyTraitsID, string(identity.CredentialsTypePassword), string(identity.CredentialsTypeOIDC)} {
			var f form.HTMLForm
			if err := faker.FakeData(&f); err != nil {
				return nil, err
			}
			methods[ct] = &settings.RequestMethod{
				Method: ct,
				Config: &settings.RequestMethodConfig{RequestMethodConfigurator: &f},
			}
		}
		return methods, nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("uuid", func(v reflect.Value) (interface{}, error) {
		return x.NewUUID(), nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("identity", func(v reflect.Value) (interface{}, error) {
		var i identity.Identity
		return &i, faker.FakeData(&i)
	}); err != nil {
		panic(err)
	}
}
