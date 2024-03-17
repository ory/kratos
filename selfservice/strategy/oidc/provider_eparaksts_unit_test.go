package oidc

import (
	"testing"
)

type User struct {
	SerialNumber      string
	ParseSerialNumber (ProviderEParaksts)
}

func TestSerialNumberModification(t *testing.T) {
	expectedSerialNumber := "1234678900"
	user := &User{SerialNumber: "PVOLV-12346-78900"}

	g := &ProviderEParaksts{}

	actualSerialNumber := g.ParseSerialNumber(user.SerialNumber)

	if actualSerialNumber != expectedSerialNumber {
		t.Errorf("Expected SerialNumber to be "+expectedSerialNumber+", but got ", user.SerialNumber)
	}
}

func ParseSerialNumber(s string, provider ProviderEParaksts) {
	panic("unimplemented")
}
