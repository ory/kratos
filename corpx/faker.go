package corpx

import (
	"math/rand"
	"net/http"
	"reflect"
	"time"

	"github.com/ory/kratos/ui/container"

	"github.com/ory/kratos/ui/node"

	"github.com/bxcodec/faker/v3"

	"github.com/ory/x/randx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"

	"github.com/ory/kratos/x"
)

var setup bool

func RegisterFakes() {
	if setup {
		return
	}
	setup = true

	_ = faker.SetRandomMapAndSliceSize(4)

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

	if err := faker.AddProvider("http_method", func(v reflect.Value) (interface{}, error) {
		methods := []string{"POST", "PUT", "GET", "PATCH"}
		return methods[rand.Intn(len(methods))], nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("identity_credentials_type", func(v reflect.Value) (interface{}, error) {
		methods := []identity.CredentialsType{identity.CredentialsTypePassword, identity.CredentialsTypePassword}
		return string(methods[rand.Intn(len(methods))]), nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("string", func(v reflect.Value) (interface{}, error) {
		return randx.MustString(25, randx.AlphaNum), nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("time_type", func(v reflect.Value) (interface{}, error) {
		return time.Now().Add(time.Duration(rand.Int())).Round(time.Second).UTC(), nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("settings_flow_methods", func(v reflect.Value) (interface{}, error) {
		var methods = make(map[string]*settings.FlowMethod)
		for _, ct := range []string{settings.StrategyProfile, string(identity.CredentialsTypePassword), string(identity.CredentialsTypeOIDC)} {
			var f container.Container
			if err := faker.FakeData(&f); err != nil {
				return nil, err
			}
			methods[ct] = &settings.FlowMethod{
				Method: ct,
				Config: &settings.FlowMethodConfig{FlowMethodConfigurator: &f},
			}
		}
		return methods, nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("recovery_flow_methods", func(v reflect.Value) (interface{}, error) {
		var methods = make(map[string]*recovery.FlowMethod)
		for _, ct := range []string{recovery.StrategyRecoveryLinkName} {
			var f container.Container
			if err := faker.FakeData(&f); err != nil {
				return nil, err
			}
			methods[ct] = &recovery.FlowMethod{
				Method: ct,
				Config: &recovery.FlowMethodConfig{FlowMethodConfigurator: &f},
			}
		}
		return methods, nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("verification_flow_methods", func(v reflect.Value) (interface{}, error) {
		var methods = make(map[string]*verification.FlowMethod)
		for _, ct := range []string{verification.StrategyVerificationLinkName} {
			var f container.Container
			if err := faker.FakeData(&f); err != nil {
				return nil, err
			}
			methods[ct] = &verification.FlowMethod{
				Method: ct,
				Config: &verification.FlowMethodConfig{FlowMethodConfigurator: &f},
			}
		}
		return methods, nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("ui_node_attributes", func(v reflect.Value) (interface{}, error) {
		var a node.Attributes
		switch rand.Intn(4) {
		case 0:
			a = new(node.InputAttributes)
		case 1:
			a = new(node.ImageAttributes)
		case 2:
			a = new(node.AnchorAttributes)
		case 3:
			a = new(node.TextAttributes)
		}

		if err := faker.FakeData(a); err != nil {
			return nil, err
		}
		return a, nil
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

	if err := faker.AddProvider("flow_type", func(v reflect.Value) (interface{}, error) {
		if rand.Intn(2) == 0 {
			return flow.TypeAPI, nil
		}
		return flow.TypeBrowser, nil
	}); err != nil {
		panic(err)
	}
}
