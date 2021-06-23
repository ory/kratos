package main

import (
	"testing"

	"github.com/ory/kratos/examples/go/pkg"
)

func TestRegistration(t *testing.T) {
	publicURL, _ := pkg.NewKratosServer(t)
	client = pkg.NewSDKForSelfHosted(publicURL)
	performRegistration()
}
