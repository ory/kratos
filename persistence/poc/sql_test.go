package poc

import (
	"log"
	"testing"

	"github.com/gobuffalo/pop"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/persistence/poc/models"
	"github.com/ory/kratos/selfservice/form"
)

func TestRun(t *testing.T) {
	tx, err := pop.Connect("development")
	if err != nil {
		log.Panic(err)
	}

	var reqs models.Requests
	require.NoError(t, tx.Eager().All(&reqs))

	for _, r := range reqs {
		t.Logf("req: %s\n", r.Methods)
		t.Logf("req: %s\n", r.PPMethods)
	}

	// peter := models.Request{
	// 	Methods: []models.Method{{}},
	// }
	peter := models.Request{
		Methods: map[string]models.Method{
			"bar": {
				Config: &models.MethodConfig{
					MethodConfigurator: &form.HTMLForm{
						Action: "foo",
						Method: "bar",
						Fields: map[string]form.Field{
							"foo": {Name: "baz"},
						},
						Errors: []form.Error{{Message: "some error"}},
					},
				},
			},
		},
	}
	_, err = tx.Eager().ValidateAndSave(&peter)
	require.NoError(t, err)
	t.Logf("peter: %+v", peter.Methods)

	// c, err := pop.NewConnection(&pop.ConnectionDetails{URL :"postgres://postgres:secret@localhost:3445/postgres?sslmode=disable"})
	// pop.Connect()
}
