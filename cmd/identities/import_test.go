package identities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/ory/x/cmdx"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/x/pointerx"
)

func TestImportCmd(t *testing.T) {
	reg := setup(t, importCmd)

	t.Run("case=imports a new identity from file", func(t *testing.T) {
		i := models.CreateIdentity{
			SchemaID: pointerx.String(config.DefaultIdentityTraitsSchemaID),
			Traits:   map[string]interface{}{},
		}
		ij, err := json.Marshal(i)
		require.NoError(t, err)
		f, err := ioutil.TempFile("", "")
		require.NoError(t, err)
		_, err = f.Write(ij)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		stdOut := execNoErr(t, importCmd, f.Name())

		id, err := uuid.FromString(gjson.Get(stdOut, "id").String())
		require.NoError(t, err)
		_, err = reg.Persister().GetIdentity(context.Background(), id)
		assert.NoError(t, err)
	})

	t.Run("case=imports multiple identities from single file", func(t *testing.T) {
		i := []models.CreateIdentity{
			{
				SchemaID: pointerx.String(config.DefaultIdentityTraitsSchemaID),
				Traits:   map[string]interface{}{},
			},
			{
				SchemaID: pointerx.String(config.DefaultIdentityTraitsSchemaID),
				Traits:   map[string]interface{}{},
			},
		}
		ij, err := json.Marshal(i)
		require.NoError(t, err)
		f, err := ioutil.TempFile("", "")
		require.NoError(t, err)
		_, err = f.Write(ij)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		stdOut := execNoErr(t, importCmd, f.Name())

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
		i := []models.CreateIdentity{
			{
				SchemaID: pointerx.String(config.DefaultIdentityTraitsSchemaID),
				Traits:   map[string]interface{}{},
			},
			{
				SchemaID: pointerx.String(config.DefaultIdentityTraitsSchemaID),
				Traits:   map[string]interface{}{},
			},
		}
		ij, err := json.Marshal(i)
		require.NoError(t, err)

		stdOut, stdErr, err := exec(importCmd, bytes.NewBuffer(ij))
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
		i := models.CreateIdentity{
			SchemaID: pointerx.String(config.DefaultIdentityTraitsSchemaID),
			Traits:   map[string]interface{}{},
		}
		ij, err := json.Marshal(i)
		require.NoError(t, err)

		stdOut, stdErr, err := exec(importCmd, bytes.NewBuffer(ij))
		require.NoError(t, err, "%s %s", stdOut, stdErr)

		id, err := uuid.FromString(gjson.Get(stdOut, "id").String())
		require.NoError(t, err)
		_, err = reg.Persister().GetIdentity(context.Background(), id)
		assert.NoError(t, err)
	})

	t.Run("case=fails to import invalid identity", func(t *testing.T) {
		// validation is further tested with the validate command
		stdOut, stdErr, err := exec(importCmd, bytes.NewBufferString("{}"))
		assert.True(t, errors.Is(err, cmdx.ErrNoPrintButFail))
		assert.Contains(t, stdErr, "STD_IN[0]: not valid")
		assert.Len(t, stdOut, 0)
	})
}
