package identities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/internal/clihelpers"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/x/pointerx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"testing"
)

func TestImportCmd(t *testing.T) {
	reg := setup(t, importCmd)

	t.Run("case=imports a new identity from file", func(t *testing.T) {
		i := models.CreateIdentity{
			SchemaID: pointerx.String(configuration.DefaultIdentityTraitsSchemaID),
			Traits: map[string]interface{}{},
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

	t.Run("case=imports a new identity from stdIn", func(t *testing.T) {
		i := models.CreateIdentity{
			SchemaID: pointerx.String(configuration.DefaultIdentityTraitsSchemaID),
			Traits: map[string]interface{}{},
		}
		ij, err := json.Marshal(i)
		require.NoError(t, err)

		stdOut, stdErr, err := exec(importCmd, bytes.NewBuffer(ij))
		require.NoError(t, err, stdOut, stdErr)

		id, err := uuid.FromString(gjson.Get(stdOut, "id").String())
		require.NoError(t, err)
		_, err = reg.Persister().GetIdentity(context.Background(), id)
		assert.NoError(t, err)
	})

	t.Run("case=fails to import invalid identity", func(t *testing.T) {
		// validation is further tested with the validate command
		stdOut, stdErr, err := exec(importCmd, bytes.NewBufferString("{}"))
		assert.True(t, errors.Is(err, clihelpers.NoPrintButFailError))
		assert.Contains(t, stdErr, "STD_IN: not valid")
		assert.Len(t, stdOut, 0)
	})
}
