package identity_test

import (
	"testing"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
)

func TestManager(t *testing.T) {
	// _, reg := internal.NewRegistryDefault(t)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/extension/schema.json")
	viper.Set(configuration.ViperKeyURLsSelfPublic, "https://www.ory.sh/")
	viper.Set(configuration.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")

	t.Fatal("missing tests")
	//
	// i := NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
	// require.NoError(t, reg.IdentityPool().CreateIdentity(context.Background(), i))
	//
	// t.Run("method=TrackVerifiableAddresses", func(t *testing.T) {
	// 	addresses := []Address{
	// 		*MustNewEmailAddress("foo@ory.sh", i.ID, time.Minute),
	// 		*MustNewEmailAddress("bar@ory.sh", i.ID, time.Minute),
	// 	}
	// 	require.NoError(t, reg.VerificationManager().TrackVerifiableAddresses(context.Background(), addresses))
	//
	// 	actual, err := reg.VerificationPersister().FindAddressByValue(context.Background(), ViaEmail, "foo@ory.sh")
	// 	require.NoError(t, err)
	// 	assert.EqualValues(t, "foo@ory.sh", actual.Value)
	//
	// 	actual, err = reg.VerificationPersister().FindAddressByValue(context.Background(), ViaEmail, "bar@ory.sh")
	// 	require.NoError(t, err)
	// 	assert.EqualValues(t, "bar@ory.sh", actual.Value)
	//
	// 	messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
	// 	require.NoError(t, err)
	// 	require.Len(t, messages, 2)
	//
	// 	assert.EqualValues(t, "foo@ory.sh", messages[0].Recipient)
	// 	assert.Contains(t, messages[0].Subject, "Please verify")
	// 	assert.EqualValues(t, "bar@ory.sh", messages[1].Recipient)
	// 	assert.Contains(t, messages[1].Subject, "Please verify")
	// })
	//
	// t.Run("method=SendCode", func(t *testing.T) {
	// 	require.NoError(t, reg.VerificationManager().SendCode(context.Background(), ViaEmail, "not-tracked@ory.sh"))
	//
	// 	messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
	// 	require.NoError(t, err)
	//
	// 	require.Len(t, messages, 3)
	// 	assert.EqualValues(t, "not-tracked@ory.sh", messages[2].Recipient)
	// 	assert.Contains(t, messages[2].Subject, "tried to verify")
	// })
	//
	//
	//
	//
	// t.Run("case=fail to update an identity because credentials changed but update was called", func(t *testing.T) {
	// 	initial := oidcIdentity("", x.NewUUID().String())
	// 	require.NoError(t, p.CreateIdentity(context.Background(), initial))
	// 	createdIDs = append(createdIDs, initial.ID)
	//
	// 	assert.Equal(t, configuration.DefaultIdentityTraitsSchemaID, initial.TraitsSchemaID)
	// 	assert.Equal(t, defaultSchema.SchemaURL(exampleServerURL).String(), initial.TraitsSchemaURL)
	//
	// 	toUpdate := initial.CopyWithoutCredentials()
	// 	toUpdate.SetCredentials(CredentialsTypePassword, Credentials{
	// 		Type:        CredentialsTypePassword,
	// 		Identifiers: []string{"ignore-me"},
	// 		Config:      json.RawMessage(`{"oh":"nono"}`),
	// 	})
	// 	toUpdate.Traits = Traits(`{"update":"me"}`)
	// 	toUpdate.TraitsSchemaID = altSchema.ID
	//
	// 	err := p.UpdateIdentity(context.Background(), toUpdate)
	// 	require.Error(t, err)
	// 	assert.Contains(t, fmt.Sprintf("%+v", err), "A field was modified that updates one or more credentials-related settings.")
	//
	// 	actual, err := p.GetIdentityConfidential(context.Background(), toUpdate.ID)
	// 	require.NoError(t, err)
	// 	assert.Equal(t, configuration.DefaultIdentityTraitsSchemaID, actual.TraitsSchemaID)
	// 	assert.Equal(t, defaultSchema.SchemaURL(exampleServerURL).String(), actual.TraitsSchemaURL)
	// 	assert.Empty(t, actual.Credentials[CredentialsTypePassword])
	// 	assert.NotEmpty(t, actual.Credentials[CredentialsTypeOIDC])
	// })

}
