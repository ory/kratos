package oidc_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/ory/x/assertx"
	"github.com/ory/x/jsonx"
	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/ui/container"

	kratos "github.com/ory/kratos-client-go"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/kratos/corpx"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"

	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
)

func init() {
	corpx.RegisterFakes()
}

var (
	csrfField = testhelpers.NewFakeCSRFNode()
)

func newFakeProfile(email string) []kratos.UiNode {
	return []kratos.UiNode{
		*testhelpers.NewFakeCSRFNode(),
		{
			Type:  "input",
			Group: "profile",
			Attributes: kratos.UiNodeInputAttributesAsUiNodeAttributes(&kratos.UiNodeInputAttributes{
				Name:  "traits.email",
				Type:  "email",
				Value: &kratos.UiNodeInputAttributesValue{String: pointerx.String(email)},
			}),
		},
		{
			Type:  "input",
			Group: "profile",
			Attributes: kratos.UiNodeInputAttributesAsUiNodeAttributes(&kratos.UiNodeInputAttributes{
				Name: "traits.name",
				Type: "text",
			}),
		},
		*testhelpers.NewMethodSubmit("authenticator_password", "password"),
		*testhelpers.NewPasswordNode(),
	}
}

func TestSettingsStrategy(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var (
		conf, reg = internal.NewFastRegistryWithMocks(t)
		subject   string
		website   string
		scope     []string
	)

	remoteAdmin, remotePublic, _ := newHydra(t, &subject, &website, &scope)
	uiTS := newUI(t, reg)
	errTS := testhelpers.NewErrorTestServer(t, reg)
	publicTS, adminTS := testhelpers.NewKratosServers(t)

	viperSetProviderConfig(
		t,
		conf,
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "ory", "ory"),
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "google", "google"),
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "github", "github"),
	)
	testhelpers.InitKratosServers(t, reg, publicTS, adminTS)
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/settings.schema.json")
	conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/kratos")

	// Make test data for this test run unique
	testID := x.NewUUID().String()
	users := map[string]*identity.Identity{
		"password": {ID: x.NewUUID(), Traits: identity.Traits(`{"email":"john` + testID + `@doe.com"}`),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {Type: "password",
					Identifiers: []string{"john+" + testID + "@doe.com"},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$argon2id$iammocked...."}`)}},
		},
		"oryer": {ID: x.NewUUID(), Traits: identity.Traits(`{"email":"hackerman+` + testID + `@ory.sh"}`),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypeOIDC: {Type: identity.CredentialsTypeOIDC,
					Identifiers: []string{"ory:hackerman+" + testID},
					Config:      sqlxx.JSONRawMessage(`{"providers":[{"provider":"ory","subject":"hackerman+` + testID + `"}]}`)}},
		},
		"githuber": {ID: x.NewUUID(), Traits: identity.Traits(`{"email":"hackerman+github+` + testID + `@ory.sh"}`),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypeOIDC: {Type: identity.CredentialsTypeOIDC,
					Identifiers: []string{"ory:hackerman+github+" + testID, "github:hackerman+github+" + testID},
					Config:      sqlxx.JSONRawMessage(`{"providers":[{"provider":"ory","subject":"hackerman+github+` + testID + `"},{"provider":"github","subject":"hackerman+github+` + testID + `"}]}`)}},
			SchemaID: config.DefaultIdentityTraitsSchemaID,
		},
		"multiuser": {ID: x.NewUUID(), Traits: identity.Traits(`{"email":"hackerman+multiuser+` + testID + `@ory.sh"}`),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {Type: "password",
					Identifiers: []string{"hackerman+multiuser+" + testID + "@ory.sh"},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$argon2id$iammocked...."}`)},
				identity.CredentialsTypeOIDC: {Type: identity.CredentialsTypeOIDC,
					Identifiers: []string{"ory:hackerman+multiuser+" + testID, "google:hackerman+multiuser+" + testID},
					Config:      sqlxx.JSONRawMessage(`{"providers":[{"provider":"ory","subject":"hackerman+multiuser+` + testID + `"},{"provider":"google","subject":"hackerman+multiuser+` + testID + `"}]}`)}},
			SchemaID: config.DefaultIdentityTraitsSchemaID,
		},
	}
	agents := testhelpers.AddAndLoginIdentities(t, reg, publicTS, users)

	var newProfileFlow = func(t *testing.T, client *http.Client, redirectTo string, exp time.Duration) *settings.Flow {
		req, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(),
			x.ParseUUID(string(testhelpers.InitializeSettingsFlowViaBrowser(t, client, false, publicTS).Id)))
		require.NoError(t, err)
		assert.Empty(t, req.Active)

		if redirectTo != "" {
			req.RequestURL = redirectTo
		}
		req.ExpiresAt = time.Now().Add(exp)
		require.NoError(t, reg.SettingsFlowPersister().UpdateSettingsFlow(context.Background(), req))

		// sanity check
		got, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), req.ID)
		require.NoError(t, err)
		require.Len(t, got.UI.Nodes, len(req.UI.Nodes))

		return req
	}

	// does the same as new profile request but uses the SDK
	var nprSDK = func(t *testing.T, client *http.Client, redirectTo string, exp time.Duration) *kratos.SettingsFlow {
		return testhelpers.InitializeSettingsFlowViaBrowser(t, client, false, publicTS)
	}

	t.Run("case=should not be able to continue a flow with a malformed ID", func(t *testing.T) {
		body, res := testhelpers.HTTPPostForm(t, agents["password"], publicTS.URL+settings.RouteSubmitFlow+"?flow=i-am-not-a-uuid", nil)
		AssertSystemError(t, errTS, res, body, 400, "malformed")
	})

	t.Run("case=should not be able to continue a flow without the flow query parameter", func(t *testing.T) {
		body, res := testhelpers.HTTPPostForm(t, agents["password"], publicTS.URL+settings.RouteSubmitFlow, nil)
		AssertSystemError(t, errTS, res, body, 400, "query parameter is missing")
	})

	t.Run("case=should not be able to continue a flow with a non-existing ID", func(t *testing.T) {
		body, res := testhelpers.HTTPPostForm(t, agents["password"], publicTS.URL+settings.RouteSubmitFlow+"?flow="+x.NewUUID().String(), nil)
		AssertSystemError(t, errTS, res, body, 404, "not be found")
	})

	t.Run("case=should not be able to continue a flow that is expired", func(t *testing.T) {
		req := newProfileFlow(t, agents["password"], "", -time.Hour)
		body, res := testhelpers.HTTPPostForm(t, agents["password"], publicTS.URL+settings.RouteSubmitFlow+"?flow="+req.ID.String(), nil)

		require.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow=")
		assert.NotContains(t, res.Request.URL.String(), req.ID.String(), "should initialize a new flow")
		assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(), "expired")
	})

	t.Run("case=should not be able to fetch another user's data", func(t *testing.T) {
		req := newProfileFlow(t, agents["password"], "", time.Hour)

		_, _, err := testhelpers.NewSDKCustomClient(publicTS, agents["oryer"]).PublicApi.GetSelfServiceSettingsFlow(context.Background()).Id(req.ID.String()).Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "403")
	})

	t.Run("case=should fetch the settings request and expect data to be set appropriately", func(t *testing.T) {
		req := newProfileFlow(t, agents["password"], "", time.Hour)

		rs, _, err := testhelpers.NewSDKCustomClient(adminTS, agents["password"]).PublicApi.GetSelfServiceSettingsFlow(context.Background()).Id(req.ID.String()).Execute()
		require.NoError(t, err)

		// Check our sanity. Does the SDK relay the same info that we expect and got from the store?
		assert.Equal(t, publicTS.URL+"/self-service/settings/browser", req.RequestURL)
		assert.Empty(t, req.Active)
		assert.NotEmpty(t, req.IssuedAt)
		assert.EqualValues(t, users["password"].ID, req.Identity.ID)
		assert.EqualValues(t, users["password"].Traits, req.Identity.Traits)
		assert.EqualValues(t, users["password"].SchemaID, req.Identity.SchemaID)

		assert.EqualValues(t, req.ID.String(), rs.Id)
		assert.EqualValues(t, req.RequestURL, rs.RequestUrl)
		assert.EqualValues(t, req.Identity.ID.String(), rs.Identity.Id)
		assert.EqualValues(t, req.IssuedAt, rs.IssuedAt)

		require.NotNil(t, identity.CredentialsTypeOIDC.String(), rs.Ui)
		require.EqualValues(t, "POST", rs.Ui.Method)
		require.EqualValues(t, publicTS.URL+settings.RouteSubmitFlow+"?flow="+req.ID.String(),
			rs.Ui.Action)
	})

	expectedPasswordFields := json.RawMessage(`[
  {
    "attributes": {
      "disabled": false,
      "name": "csrf_token",
      "required": true,
      "type": "hidden",
      "value": "` + x.FakeCSRFToken + `"
    },
    "group": "default",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "traits.email",
      "type": "email",
      "value": "john` + testID + `@doe.com"
    },
    "group": "profile",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "traits.name",
      "type": "text"
    },
    "group": "profile",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "method",
      "type": "submit",
      "value": "profile"
    },
    "group": "profile",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "password",
      "required": true,
      "type": "password"
    },
    "group": "password",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070001,
        "text": "Password",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "method",
      "type": "submit",
      "value": "password"
    },
    "group": "password",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "link",
      "type": "submit",
      "value": "github"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "github"
        },
        "id": 1050002,
        "text": "Link github",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "link",
      "type": "submit",
      "value": "google"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "google"
        },
        "id": 1050002,
        "text": "Link google",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "link",
      "type": "submit",
      "value": "ory"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "ory"
        },
        "id": 1050002,
        "text": "Link ory",
        "type": "info"
      }
    },
    "type": "input"
  }
]`)
	expectedOryerFields := json.RawMessage(`[
  {
    "type": "input",
    "group": "default",
    "attributes": {
      "name": "csrf_token",
      "type": "hidden",
      "value": "` + x.FakeCSRFToken + `",
      "required": true,
      "disabled": false
    },
    "messages": [],
    "meta": {}
  },
  {
    "type": "input",
    "group": "profile",
    "attributes": {
      "name": "traits.email",
      "type": "email",
      "value": "hackerman+` + testID + `@ory.sh",
      "disabled": false
    },
    "messages": [],
    "meta": {}
  },
  {
    "type": "input",
    "group": "profile",
    "attributes": {
      "name": "traits.name",
      "type": "text",
      "disabled": false
    },
    "messages": [],
    "meta": {}
  },
  {
    "type": "input",
    "group": "profile",
    "attributes": {
      "name": "method",
      "type": "submit",
      "value": "profile",
      "disabled": false
    },
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    }
  },
  {
    "type": "input",
    "group": "password",
    "attributes": {
      "name": "password",
      "type": "password",
      "required": true,
      "disabled": false
    },
    "messages": [],
    "meta": {
      "label": {
        "id": 1070001,
        "text": "Password",
        "type": "info"
      }
    }
  },
  {
    "type": "input",
    "group": "password",
    "attributes": {
      "name": "method",
      "type": "submit",
      "value": "password",
      "disabled": false
    },
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    }
  },
  {
    "type": "input",
    "group": "oidc",
    "attributes": {
      "name": "link",
      "type": "submit",
      "value": "github",
      "disabled": false
    },
    "messages": [],
    "meta": {
      "label": {
        "id": 1050002,
        "text": "Link github",
        "type": "info",
        "context": {
          "provider": "github"
        }
      }
    }
  },
  {
    "type": "input",
    "group": "oidc",
    "attributes": {
      "name": "link",
      "type": "submit",
      "value": "google",
      "disabled": false
    },
    "messages": [],
    "meta": {
      "label": {
        "id": 1050002,
        "text": "Link google",
        "type": "info",
        "context": {
          "provider": "google"
        }
      }
    }
  }
]`)
	expectedGithuberFields := json.RawMessage(`[
  {
    "type": "input",
    "group": "default",
    "attributes": {
      "name": "csrf_token",
      "type": "hidden",
      "value": "` + x.FakeCSRFToken + `",
      "required": true,
      "disabled": false
    },
    "messages": [],
    "meta": {}
  },
  {
    "type": "input",
    "group": "profile",
    "attributes": {
      "name": "traits.email",
      "type": "email",
      "value": "hackerman+github+` + testID + `@ory.sh",
      "disabled": false
    },
    "messages": [],
    "meta": {}
  },
  {
    "type": "input",
    "group": "profile",
    "attributes": {
      "name": "traits.name",
      "type": "text",
      "disabled": false
    },
    "messages": [],
    "meta": {}
  },
  {
    "type": "input",
    "group": "profile",
    "attributes": {
      "name": "method",
      "type": "submit",
      "value": "profile",
      "disabled": false
    },
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    }
  },
  {
    "type": "input",
    "group": "password",
    "attributes": {
      "name": "password",
      "type": "password",
      "required": true,
      "disabled": false
    },
    "messages": [],
    "meta": {
      "label": {
        "id": 1070001,
        "text": "Password",
        "type": "info"
      }
    }
  },
  {
    "type": "input",
    "group": "password",
    "attributes": {
      "name": "method",
      "type": "submit",
      "value": "password",
      "disabled": false
    },
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    }
  },
  {
    "type": "input",
    "group": "oidc",
    "attributes": {
      "name": "unlink",
      "type": "submit",
      "value": "github",
      "disabled": false
    },
    "messages": [],
    "meta": {
      "label": {
        "id": 1050003,
        "text": "Unlink github",
        "type": "info",
        "context": {
          "provider": "github"
        }
      }
    }
  },
  {
    "type": "input",
    "group": "oidc",
    "attributes": {
      "name": "link",
      "type": "submit",
      "value": "google",
      "disabled": false
    },
    "messages": [],
    "meta": {
      "label": {
        "id": 1050002,
        "text": "Link google",
        "type": "info",
        "context": {
          "provider": "google"
        }
      }
    }
  },
  {
    "type": "input",
    "group": "oidc",
    "attributes": {
      "name": "unlink",
      "type": "submit",
      "value": "ory",
      "disabled": false
    },
    "messages": [],
    "meta": {
      "label": {
        "id": 1050003,
        "text": "Unlink ory",
        "type": "info",
        "context": {
          "provider": "ory"
        }
      }
    }
  }
]
`)

	t.Run("case=should adjust linkable providers based on linked credentials", func(t *testing.T) {
		for _, tc := range []struct {
			agent    string
			expected json.RawMessage
		}{
			{agent: "password", expected: json.RawMessage(jsonx.TestMarshalJSONString(t, expectedPasswordFields))},
			{agent: "oryer", expected: json.RawMessage(jsonx.TestMarshalJSONString(t, expectedOryerFields))},
			{agent: "githuber", expected: json.RawMessage(jsonx.TestMarshalJSONString(t, expectedGithuberFields))},
			{agent: "multiuser", expected: json.RawMessage(`[
  {
    "attributes": {
      "disabled": false,
      "name": "csrf_token",
      "required": true,
      "type": "hidden",
      "value": "` + x.FakeCSRFToken + `"
    },
    "group": "default",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "traits.email",
      "type": "email",
      "value": "hackerman+multiuser+` + testID + `@ory.sh"
    },
    "group": "profile",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "traits.name",
      "type": "text"
    },
    "group": "profile",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "method",
      "type": "submit",
      "value": "profile"
    },
    "group": "profile",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "password",
      "required": true,
      "type": "password"
    },
    "group": "password",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070001,
        "text": "Password",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "method",
      "type": "submit",
      "value": "password"
    },
    "group": "password",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "link",
      "type": "submit",
      "value": "github"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "github"
        },
        "id": 1050002,
        "text": "Link github",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "unlink",
      "type": "submit",
      "value": "google"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "google"
        },
        "id": 1050003,
        "text": "Unlink google",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "unlink",
      "type": "submit",
      "value": "ory"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "ory"
        },
        "id": 1050003,
        "text": "Unlink ory",
        "type": "info"
      }
    },
    "type": "input"
  }
]`)},
		} {
			t.Run("agent="+tc.agent, func(t *testing.T) {
				rs := nprSDK(t, agents[tc.agent], "", time.Hour)
				assertx.EqualAsJSON(t, tc.expected, rs.Ui.Nodes)
			})
		}
	})

	var action = func(req *kratos.SettingsFlow) string {
		return req.Ui.Action
	}

	var checkCredentials = func(t *testing.T, shouldExist bool, iid uuid.UUID, provider, subject string) {
		actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), iid)
		require.NoError(t, err)

		var cc oidc.CredentialsConfig
		creds, err := actual.ParseCredentials(identity.CredentialsTypeOIDC, &cc)
		require.NoError(t, err)

		if shouldExist {
			assert.Contains(t, creds.Identifiers, provider+":"+subject)
		} else {
			assert.NotContains(t, creds.Identifiers, provider+":"+subject)
		}

		var found bool
		for _, p := range cc.Providers {
			if p.Provider == provider && p.Subject == subject {
				found = true
				break
			}
		}

		require.EqualValues(t, shouldExist, found)
	}

	var reset = func(t *testing.T) func() {
		return func() {
			conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Minute*5)
			agents = testhelpers.AddAndLoginIdentities(t, reg, publicTS, users)
		}
	}

	t.Run("suite=unlink", func(t *testing.T) {
		var unlink = func(t *testing.T, agent, provider string) (body []byte, res *http.Response, req *kratos.SettingsFlow) {
			req = nprSDK(t, agents[agent], "", time.Hour)
			body, res = testhelpers.HTTPPostForm(t, agents[agent], action(req),
				&url.Values{"csrf_token": {x.FakeCSRFToken}, "unlink": {provider}})
			return
		}

		var unlinkInvalid = func(agent, provider string, expectedFields json.RawMessage) func(t *testing.T) {
			return func(t *testing.T) {
				body, res, req := unlink(t, agent, provider)
				assertx.EqualAsJSON(t, expectedFields, req.Ui.Nodes, "%s", body)

				t.Logf("%s", req.Id)
				t.Logf("%s", body)

				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)

				//assert.EqualValues(t, identity.CredentialsTypeOIDC.String(), gjson.GetBytes(body, "active").String())
				assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), publicTS.URL+settings.RouteSubmitFlow+"?flow=")

				// The original options to link google and github are still there
				assertx.EqualAsJSON(t, expectedFields,
					json.RawMessage(gjson.GetBytes(body, `ui.nodes`).Raw), "%s", body)

				assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(),
					"can not unlink non-existing OpenID Connect")
			}
		}

		t.Run("case=should not be able to unlink the last remaining connection",
			unlinkInvalid("oryer", "ory", expectedOryerFields))

		t.Run("case=should not be able to unlink an non-existing connection",
			unlinkInvalid("oryer", "i-do-not-exist", expectedOryerFields))

		t.Run("case=should not be able to unlink a connection not yet linked",
			unlinkInvalid("githuber", "google", expectedGithuberFields))

		t.Run("case=should unlink a connection", func(t *testing.T) {
			agent, provider := "githuber", "github"
			t.Cleanup(reset(t))

			body, res, req := unlink(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)
			require.Equal(t, "success", gjson.GetBytes(body, "state").String(), "%s", body)

			checkCredentials(t, false, users[agent].ID, provider, "hackerman+github+"+testID)
		})

		t.Run("case=should not be able to unlink a connection without a privileged session", func(t *testing.T) {
			agent, provider := "githuber", "github"

			var runUnauthed = func(t *testing.T) *kratos.SettingsFlow {
				conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Millisecond)
				time.Sleep(time.Millisecond)
				t.Cleanup(reset(t))
				_, res, req := unlink(t, agent, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login")

				rs, _, err := testhelpers.NewSDKCustomClient(adminTS, agents[agent]).PublicApi.GetSelfServiceSettingsFlow(context.Background()).Id(req.Id).Execute()
				require.NoError(t, err)
				require.EqualValues(t, settings.StateShowForm, rs.State)

				checkCredentials(t, true, users[agent].ID, provider, "hackerman+github+"+testID)

				return req
			}

			t.Run("subcase=should not update without re-auth", func(t *testing.T) {
				_ = runUnauthed(t)
			})

			t.Run("subcase=should update after re-auth", func(t *testing.T) {
				req := runUnauthed(t)

				// fake login by allowing longer sessions...
				conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Minute*5)

				body, res := testhelpers.HTTPPostForm(t, agents[agent], action(req),
					&url.Values{"csrf_token": {x.FakeCSRFToken}, "unlink": {provider}})
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)

				assert.Equal(t, "success", gjson.GetBytes(body, "state").String())

				checkCredentials(t, false, users[agent].ID, provider, "hackerman+github+"+testID)
			})
		})
	})

	t.Run("suite=link", func(t *testing.T) {
		var link = func(t *testing.T, agent, provider string) (body []byte, res *http.Response, req *kratos.SettingsFlow) {
			req = nprSDK(t, agents[agent], "", time.Hour)
			body, res = testhelpers.HTTPPostForm(t, agents[agent], action(req),
				&url.Values{"csrf_token": {x.FakeCSRFToken}, "link": {provider}})
			return
		}

		var linkInvalid = func(agent, provider string, expectedFields json.RawMessage) func(t *testing.T) {
			return func(t *testing.T) {
				body, res, req := link(t, agent, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)

				//assert.EqualValues(t, identity.CredentialsTypeOIDC.String(), gjson.GetBytes(body, "active").String())
				assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), publicTS.URL+settings.RouteSubmitFlow+"?flow=")

				// The original options to link google and github are still there
				assertx.EqualAsJSON(t, expectedFields, json.RawMessage(gjson.GetBytes(body, `ui.nodes`).Raw))

				assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(),
					"can not link unknown or already existing OpenID Connect connection")
			}
		}

		t.Run("case=should not be able to link an non-existing connection",
			linkInvalid("oryer", "i-do-not-exist", expectedOryerFields))

		t.Run("case=should not be able to link a connection which already exists",
			linkInvalid("githuber", "github", expectedGithuberFields))

		t.Run("case=should not be able to link a connection already linked by another identity", func(t *testing.T) {
			// While this theoretically allows for account enumeration - because we see an error indicator if an
			// oidc connection is being linked that exists already - it would require the attacker to already
			// have control over the social profile, in which case account enumeration is the least of our worries.
			// Instead of using the oidc profile for enumeration, the attacker would use it for account takeover.

			// This is the multiuser login id for google
			subject = "hackerman+multiuser+" + testID
			scope = []string{"openid"}

			agent, provider := "githuber", "google"
			body, res, _ := link(t, agent, provider)

			assert.Contains(t, res.Request.URL.String(), uiTS.URL)
			assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
		})

		t.Run("case=should not be able to link a connection which is missing the ID token", func(t *testing.T) {
			t.Cleanup(reset(t))

			subject = "hackerman+scope-missing+" + testID
			scope = []string{}

			agent, provider := "githuber", "google"
			body, res, _ := link(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			t.Logf("%s", body)
			assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(),
				"no id_token was returned")
		})

		t.Run("case=should not be able to link a connection which is missing the ID token", func(t *testing.T) {
			t.Cleanup(reset(t))

			subject = "hackerman+scope-missing+" + testID
			scope = []string{}

			agent, provider := "githuber", "google"
			body, res, _ := link(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(),
				"no id_token was returned")
		})

		t.Run("case=should link a connection", func(t *testing.T) {
			t.Cleanup(reset(t))

			subject = "hackerman+new-connection+" + testID
			scope = []string{"openid"}

			agent, provider := "githuber", "google"
			updatedFlow, res, originalFlow := link(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			updatedFlowSDK, _, err := testhelpers.NewSDKCustomClient(adminTS, agents[agent]).PublicApi.GetSelfServiceSettingsFlow(context.Background()).Id(originalFlow.Id).Execute()
			require.NoError(t, err)
			require.EqualValues(t, settings.StateSuccess, updatedFlowSDK.State)

			assertx.EqualAsJSON(t, json.RawMessage(`[
  {
    "attributes": {
      "disabled": false,
      "name": "csrf_token",
      "required": true,
      "type": "hidden",
      "value": "`+x.FakeCSRFToken+`"
    },
    "group": "default",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "traits.email",
      "type": "email",
      "value": "hackerman+github+`+testID+`@ory.sh"
    },
    "group": "profile",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "traits.name",
      "type": "text"
    },
    "group": "profile",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "method",
      "type": "submit",
      "value": "profile"
    },
    "group": "profile",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "password",
      "required": true,
      "type": "password"
    },
    "group": "password",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070001,
        "text": "Password",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "method",
      "type": "submit",
      "value": "password"
    },
    "group": "password",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "unlink",
      "type": "submit",
      "value": "github"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "github"
        },
        "id": 1050003,
        "text": "Unlink github",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "link",
      "type": "submit",
      "value": "google"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "google"
        },
        "id": 1050002,
        "text": "Link google",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "unlink",
      "type": "submit",
      "value": "ory"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "ory"
        },
        "id": 1050003,
        "text": "Unlink ory",
        "type": "info"
      }
    },
    "type": "input"
  }
]
`), originalFlow.Ui.Nodes)

			expected := json.RawMessage(`[
  {
    "attributes": {
      "disabled": false,
      "name": "csrf_token",
      "required": true,
      "type": "hidden",
      "value": "` + x.FakeCSRFToken + `"
    },
    "group": "default",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "traits.email",
      "type": "email",
      "value": "hackerman+github+` + testID + `@ory.sh"
    },
    "group": "profile",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "traits.name",
      "type": "text"
    },
    "group": "profile",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "method",
      "type": "submit",
      "value": "profile"
    },
    "group": "profile",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "password",
      "required": true,
      "type": "password"
    },
    "group": "password",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070001,
        "text": "Password",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "method",
      "type": "submit",
      "value": "password"
    },
    "group": "password",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "unlink",
      "type": "submit",
      "value": "github"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "github"
        },
        "id": 1050003,
        "text": "Unlink github",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "unlink",
      "type": "submit",
      "value": "google"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "google"
        },
        "id": 1050003,
        "text": "Unlink google",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "unlink",
      "type": "submit",
      "value": "ory"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "ory"
        },
        "id": 1050003,
        "text": "Unlink ory",
        "type": "info"
      }
    },
    "type": "input"
  }
]`)
			assertx.EqualAsJSON(t, expected, json.RawMessage(gjson.GetBytes(updatedFlow, "ui.nodes").Raw), res.Request.URL)
			assertx.EqualAsJSON(t, expected, updatedFlowSDK.Ui.Nodes)

			checkCredentials(t, true, users[agent].ID, provider, subject)
		})

		t.Run("case=should link a connection even if user does not have oidc credentials yet", func(t *testing.T) {
			t.Cleanup(reset(t))

			subject = "hackerman+new-connection-new-oidc+" + testID
			scope = []string{"openid"}

			agent, provider := "password", "google"
			_, res, req := link(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			rs, _, err := testhelpers.NewSDKCustomClient(adminTS, agents[agent]).PublicApi.GetSelfServiceSettingsFlow(context.Background()).Id(req.Id).Execute()
			require.NoError(t, err)
			require.EqualValues(t, settings.StateSuccess, rs.State)

			assertx.EqualAsJSON(t, json.RawMessage(`[
  {
    "attributes": {
      "disabled": false,
      "name": "csrf_token",
      "required": true,
      "type": "hidden",
      "value": "`+x.FakeCSRFToken+`"
    },
    "group": "default",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "traits.email",
      "type": "email",
      "value": "john`+testID+`@doe.com"
    },
    "group": "profile",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "traits.name",
      "type": "text"
    },
    "group": "profile",
    "messages": [],
    "meta": {},
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "method",
      "type": "submit",
      "value": "profile"
    },
    "group": "profile",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "password",
      "required": true,
      "type": "password"
    },
    "group": "password",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070001,
        "text": "Password",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "method",
      "type": "submit",
      "value": "password"
    },
    "group": "password",
    "messages": [],
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "link",
      "type": "submit",
      "value": "github"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "github"
        },
        "id": 1050002,
        "text": "Link github",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "unlink",
      "type": "submit",
      "value": "google"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "google"
        },
        "id": 1050003,
        "text": "Unlink google",
        "type": "info"
      }
    },
    "type": "input"
  },
  {
    "attributes": {
      "disabled": false,
      "name": "link",
      "type": "submit",
      "value": "ory"
    },
    "group": "oidc",
    "messages": [],
    "meta": {
      "label": {
        "context": {
          "provider": "ory"
        },
        "id": 1050002,
        "text": "Link ory",
        "type": "info"
      }
    },
    "type": "input"
  }
]`), rs.Ui.Nodes)

			checkCredentials(t, true, users[agent].ID, provider, subject)
		})

		t.Run("case=should not be able to link a connection without a privileged session", func(t *testing.T) {
			agent, provider := "githuber", "google"
			subject = "hackerman+new+google+" + testID

			var runUnauthed = func(t *testing.T) *kratos.SettingsFlow {
				conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Millisecond)
				time.Sleep(time.Millisecond)
				t.Cleanup(reset(t))
				_, res, req := link(t, agent, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login")

				rs, _, err := testhelpers.NewSDKCustomClient(adminTS, agents[agent]).PublicApi.GetSelfServiceSettingsFlow(context.Background()).Id(req.Id).Execute()
				require.NoError(t, err)
				require.EqualValues(t, settings.StateShowForm, rs.State)

				checkCredentials(t, false, users[agent].ID, provider, subject)

				return req
			}

			t.Run("subcase=should not update without re-auth", func(t *testing.T) {
				_ = runUnauthed(t)
			})

			t.Run("subcase=should update after re-auth", func(t *testing.T) {
				req := runUnauthed(t)

				// fake login by allowing longer sessions...
				conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Minute*5)

				body, res := testhelpers.HTTPPostForm(t, agents[agent], action(req),
					&url.Values{"csrf_token": {x.FakeCSRFToken}, "unlink": {provider}})
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)

				assert.Equal(t, "success", gjson.GetBytes(body, "state").String())

				checkCredentials(t, true, users[agent].ID, provider, subject)
			})
		})
	})
}

func TestPopulateSettingsMethod(t *testing.T) {
	nreg := func(t *testing.T, conf *oidc.ConfigurationCollection) *driver.RegistryDefault {
		c, reg := internal.NewFastRegistryWithMocks(t)

		c.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://stub/registration.schema.json")
		c.MustSet(config.ViperKeyPublicBaseURL, "https://www.ory.sh/")

		// Enabled per default:
		// 		conf.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": true})
		viperSetProviderConfig(t, c, conf.Providers...)
		return reg
	}

	ns := func(t *testing.T, reg *driver.RegistryDefault) *oidc.Strategy {
		ss, err := reg.SettingsStrategies(context.Background()).Strategy(identity.CredentialsTypeOIDC.String())
		require.NoError(t, err)
		return ss.(*oidc.Strategy)
	}

	nr := func() *settings.Flow {
		return &settings.Flow{Type: flow.TypeBrowser, ID: x.NewUUID(), UI: container.New("")}
	}

	populate := func(t *testing.T, reg *driver.RegistryDefault, i *identity.Identity, req *settings.Flow) *container.Container {
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
		require.NoError(t, ns(t, reg).PopulateSettingsMethod(new(http.Request), i, req))
		require.NotNil(t, req.UI)
		require.NotNil(t, req.UI.Nodes)
		assert.Equal(t, "POST", req.UI.Method)
		return req.UI
	}

	defaultConfig := []oidc.Configuration{
		{Provider: "generic", ID: "facebook"},
		{Provider: "generic", ID: "google"},
		{Provider: "generic", ID: "github"},
	}

	t.Run("case=should not populate non-browser flow", func(t *testing.T) {
		reg := nreg(t, &oidc.ConfigurationCollection{Providers: []oidc.Configuration{{Provider: "generic", ID: "github"}}})
		i := &identity.Identity{Traits: []byte(`{"subject":"foo@bar.com"}`)}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
		req := &settings.Flow{Type: flow.TypeAPI, ID: x.NewUUID(), UI: container.New("")}
		require.NoError(t, ns(t, reg).PopulateSettingsMethod(new(http.Request), i, req))
		require.Empty(t, req.UI.Nodes)
	})

	for k, tc := range []struct {
		c      []oidc.Configuration
		i      *identity.Credentials
		e      node.Nodes
		withpw bool
	}{
		{
			c: []oidc.Configuration{},
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
			},
		},
		{
			c: []oidc.Configuration{
				{Provider: "generic", ID: "github"},
			},
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				oidc.NewLinkNode("github"),
			},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				oidc.NewLinkNode("facebook"),
				oidc.NewLinkNode("google"),
				oidc.NewLinkNode("github"),
			},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				oidc.NewLinkNode("facebook"),
				oidc.NewLinkNode("google"),
				oidc.NewLinkNode("github"),
			},
			i: &identity.Credentials{Type: identity.CredentialsTypeOIDC, Identifiers: []string{}, Config: []byte(`{}`)},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				oidc.NewLinkNode("facebook"),
				oidc.NewLinkNode("github"),
			},
			i: &identity.Credentials{Type: identity.CredentialsTypeOIDC, Identifiers: []string{
				"google:1234",
			}, Config: []byte(`{"providers":[{"provider":"google","subject":"1234"}]}`)},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				oidc.NewLinkNode("facebook"),
				oidc.NewLinkNode("github"),
				oidc.NewUnlinkNode("google"),
			},
			withpw: true,
			i: &identity.Credentials{Type: identity.CredentialsTypeOIDC, Identifiers: []string{
				"google:1234",
			},
				Config: []byte(`{"providers":[{"provider":"google","subject":"1234"}]}`)},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				oidc.NewLinkNode("github"),
				oidc.NewUnlinkNode("google"),
				oidc.NewUnlinkNode("facebook"),
			},
			i: &identity.Credentials{Type: identity.CredentialsTypeOIDC, Identifiers: []string{
				"google:1234",
				"facebook:1234",
			},
				Config: []byte(`{"providers":[{"provider":"google","subject":"1234"},{"provider":"facebook","subject":"1234"}]}`)},
		},
	} {
		t.Run("iteration="+strconv.Itoa(k), func(t *testing.T) {
			reg := nreg(t, &oidc.ConfigurationCollection{Providers: tc.c})
			i := &identity.Identity{
				Traits:      []byte(`{"subject":"foo@bar.com"}`),
				Credentials: make(map[identity.CredentialsType]identity.Credentials, 2),
			}
			if tc.i != nil {
				i.Credentials[identity.CredentialsTypeOIDC] = *tc.i
			}
			if tc.withpw {
				i.Credentials[identity.CredentialsTypePassword] = identity.Credentials{
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{"foo@bar.com"},
					Config:      []byte(`{"hashed_password":"$argon2id$..."}`),
				}
			}
			actual := populate(t, reg, i, nr())
			assert.EqualValues(t, tc.e, actual.Nodes)
		})
	}
}
