package login

import (
	"github.com/ory/kratos/identity"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithIdentityHint(t *testing.T) {
	expected := new(identity.Identity)
	opts := NewFormHydratorOptions([]FormHydratorModifier{WithIdentityHint(expected)})
	assert.Equal(t, expected, opts.IdentityHint)
}

func TestWithAccountEnumerationBucket(t *testing.T) {
	opts := NewFormHydratorOptions([]FormHydratorModifier{})
	for _, c := range identity.AllCredentialTypes {
		assert.Falsef(t, opts.BucketShowsCredential(c), "expected false for %s", c)
	}

	opts = NewFormHydratorOptions([]FormHydratorModifier{WithAccountEnumerationBucket("hello@ory.sh")})
	found := 0
	var foundType identity.CredentialsType
	for _, c := range identity.AllCredentialTypes {
		c := c
		if opts.BucketShowsCredential(c) {
			foundType = c
			found++
		}
	}

	assert.Equal(t, 1, found, "expected exactly one to be true")

	opts = NewFormHydratorOptions([]FormHydratorModifier{WithAccountEnumerationBucket("hello@ory.sh")})
	assert.Truef(t, opts.BucketShowsCredential(foundType), "expected true for %s because bucket should be stable", foundType)
}
