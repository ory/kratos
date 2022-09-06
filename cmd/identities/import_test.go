package identities_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/identities"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	kratos "github.com/ory/kratos-client-go"
	"github.com/ory/kratos/driver/config"
)

func TestImportCmd(t *testing.T) {
	c := identities.NewImportIdentitiesCmd(new(cobra.Command))
	reg := setup(t, c)

	t.Run("case=imports a new identity from file", func(t *testing.T) {
		i := kratos.AdminCreateIdentityBody{
			SchemaId: config.DefaultIdentityTraitsSchemaID,
			Traits:   map[string]interface{}{},
		}
		ij, err := json.Marshal(i)
		require.NoError(t, err)
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		_, err = f.Write(ij)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		stdOut := execNoErr(t, c, f.Name())

		id, err := uuid.FromString(gjson.Get(stdOut, "id").String())
		require.NoError(t, err)
		_, err = reg.Persister().GetIdentity(context.Background(), id)
		assert.NoError(t, err)
	})

	t.Run("case=imports multiple identities from single file", func(t *testing.T) {
		i := []kratos.AdminCreateIdentityBody{
			{
				SchemaId: config.DefaultIdentityTraitsSchemaID,
				Traits:   map[string]interface{}{},
			},
			{
				SchemaId: config.DefaultIdentityTraitsSchemaID,
				Traits:   map[string]interface{}{},
			},
		}
		ij, err := json.Marshal(i)
		require.NoError(t, err)
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		_, err = f.Write(ij)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		stdOut := execNoErr(t, c, f.Name())

		id, err := uuid.FromString(gjson.Get(stdOut, "0.id").String())
		require.NoError(t, err)
		_, err = reg.Persister().GetIdentity(context.Background(), id)
		assert.NoError(t, err)

		id, err = uuid.FromString(gjson.Get(stdOut, "1.id").String())
		require.NoError(t, err)
		_, err = reg.Persister().GetIdentity(context.Background(), id)
		assert.NoError(t, err)
	})

	t.Run("case=imports a new identity from STD_IN", func(t *testing.T) {
		i := []kratos.AdminCreateIdentityBody{
			{
				SchemaId: config.DefaultIdentityTraitsSchemaID,
				Traits:   map[string]interface{}{},
			},
			{
				SchemaId: config.DefaultIdentityTraitsSchemaID,
				Traits:   map[string]interface{}{},
			},
		}
		ij, err := json.Marshal(i)
		require.NoError(t, err)

		stdOut, stdErr, err := exec(c, bytes.NewBuffer(ij))
		require.NoError(t, err, "%s %s", stdOut, stdErr)

		id, err := uuid.FromString(gjson.Get(stdOut, "0.id").String())
		require.NoError(t, err)
		_, err = reg.Persister().GetIdentity(context.Background(), id)
		assert.NoError(t, err)

		id, err = uuid.FromString(gjson.Get(stdOut, "1.id").String())
		require.NoError(t, err)
		_, err = reg.Persister().GetIdentity(context.Background(), id)
		assert.NoError(t, err)
	})

	t.Run("case=imports multiple identities from STD_IN", func(t *testing.T) {
		i := kratos.AdminCreateIdentityBody{
			SchemaId: config.DefaultIdentityTraitsSchemaID,
			Traits:   map[string]interface{}{},
		}
		ij, err := json.Marshal(i)
		require.NoError(t, err)

		stdOut, stdErr, err := exec(c, bytes.NewBuffer(ij))
		require.NoError(t, err, "%s %s", stdOut, stdErr)

		id, err := uuid.FromString(gjson.Get(stdOut, "id").String())
		require.NoError(t, err)
		_, err = reg.Persister().GetIdentity(context.Background(), id)
		assert.NoError(t, err)
	})
}
