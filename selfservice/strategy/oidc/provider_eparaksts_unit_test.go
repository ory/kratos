package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	SerialNumber      string
	ParseSerialNumber (ProviderEParaksts)
}

func TestParseSerialNumber(t *testing.T) {
	expected := "1234678900"
	serialNumber := "PVOLV-12346-78900"

	g := &ProviderEParaksts{}

	actual := g.ParseSerialNumber(serialNumber)
	assert.Equal(t, expected, actual)
}
