package internal

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/bxcodec/faker"
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

	if err := faker.AddProvider("time_type", func(v reflect.Value) (interface{}, error) {
		return time.Now().Add(time.Duration(rand.Int())).Round(time.Second).UTC(), nil
	}); err != nil {
		panic(err)
	}
}
