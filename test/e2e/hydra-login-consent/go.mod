module github.com/ory/kratos/test/e2e/hydra-login-consent

go 1.16

replace github.com/oleiade/reflections => github.com/oleiade/reflections v1.0.1

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8

require (
	github.com/julienschmidt/httprouter v1.3.0
	github.com/ory/hydra-client-go v1.7.4
	github.com/ory/x v0.0.116
)
