package internal

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/bxcodec/faker"
)

func RegisterFakes() {
	if err := faker.AddProvider("birthdate", func(v reflect.Value) (interface{}, error) {
		return time.Now().Add(time.Duration(rand.Int())), nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("time_types", func(v reflect.Value) (interface{}, error) {
		es := make([]time.Time, rand.Intn(5))
		for k := range es {
			es[k] = time.Now().Add(time.Duration(rand.Int()))
		}
		return es, nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("time_type", func(v reflect.Value) (interface{}, error) {
		return time.Now().Add(time.Duration(rand.Int())), nil
	}); err != nil {
		panic(err)
	}

	if err := faker.AddProvider("identity_timezone", func(v reflect.Value) (interface{}, error) {
		var s struct {
			TZ string `faker:"timezone"`
		}
		if err := faker.FakeData(&s); err != nil {
			return nil, err
		}

		l, err := time.LoadLocation(s.TZ)
		if err != nil {
			return nil, err
		}
		return l, nil
	}); err != nil {
		panic(err)
	}
}
