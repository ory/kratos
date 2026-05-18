// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3/httploader"
	"github.com/ory/kratos/driver/config"
	. "github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/x/configx"
	"github.com/ory/x/httpx"
)

func TestSchemaValidatorDisallowsInternalNetworkRequests(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(testhelpers.IdentitySchemasConfig(map[string]string{
			"localhost":  "https://localhost/schema/whatever",
			"privateRef": "file://stub/localhost-ref.schema.json",
			"inlineRef":  "base64://" + base64.StdEncoding.EncodeToString([]byte(`{"traits": {}}`)),
		})),
		configx.WithValue(config.ViperKeyClientHTTPNoPrivateIPRanges, true),
	)

	v := NewValidator(reg)

	for id, expectedErr := range map[string]string{
		"localhost":  "no such host",
		"privateRef": "no route to host",
		"inlineRef":  "",
	} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()

			i := &Identity{
				SchemaID: id,
				Traits:   Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			}
			ctx := context.WithValue(t.Context(), httploader.ContextKey, reg.HTTPClient(t.Context()))
			err := v.Validate(ctx, i)
			if expectedErr == "" {
				assert.NoError(t, err)
				return
			}

			var hErr *herodot.DefaultError
			require.ErrorAs(t, err, &hErr)
			assert.Contains(t, hErr.Debug(), expectedErr)
		})
	}
}

func TestSchemaValidator(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	router := http.NewServeMux()
	router.HandleFunc("GET /schema/{name}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/person.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
	"traits": {
	  "type": "object",
	  "properties": {
        "` + r.PathValue("name") + `": {
          "type": "string",
          "description": "The person's first name."
        },
        "lastName": {
          "type": "string",
          "description": "The person's last name."
        },
        "age": {
          "description": "Age in years which must be equal to or greater than zero.",
          "type": "integer",
          "minimum": 1
        }
	  },
	  "additionalProperties": false
	}
  },
  "additionalProperties": false
}`))
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	_, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(testhelpers.IdentitySchemasConfig(map[string]string{
			"default":         ts.URL + "/schema/firstName",
			"whatever":        ts.URL + "/schema/whatever",
			"unreachable-url": ts.URL + "/404-not-found",
		})),
	)
	v := NewValidator(reg)

	for k, tc := range []struct {
		i   *Identity
		err string
	}{
		{
			i: &Identity{
				Traits: Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			},
		},
		{
			i: &Identity{
				Traits: Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": -1 }`),
			},
			err: "I[#/traits/age] S[#/properties/traits/properties/age/minimum] must be >= 1 but found -1",
		},
		{
			i: &Identity{
				Traits: Traits(`{ "whatever": "first-name", "lastName": "last-name", "age": 1 }`),
			},
			err: `I[#/traits] S[#/properties/traits/additionalProperties] additionalProperties "whatever" not allowed`,
		},
		{
			i: &Identity{
				SchemaID: "whatever",
				Traits:   Traits(`{ "whatever": "first-name", "lastName": "last-name", "age": 1 }`),
			},
		},
		{
			i: &Identity{
				SchemaID: "whatever",
				Traits:   Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			},
			err: `I[#/traits] S[#/properties/traits/additionalProperties] additionalProperties "firstName" not allowed`,
		},
		{
			i: &Identity{
				SchemaID: "unreachable-url",
				Traits:   Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			},
			err: "Invalid configuration",
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			ctx := context.WithValue(t.Context(), httploader.ContextKey, httpx.NewResilientClient())
			err := v.Validate(ctx, tc.i)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.err)
			}
		})
	}
}

// TestSchemaValidatorNormalizesIdentifierTraits verifies that trait values
// used as credential, recovery, or verification identifiers are normalized
// to the same canonical form Kratos stores in the side tables. This keeps
// webhook payloads consistent with the value Kratos uses internally.
//
// Regression test for https://github.com/ory-corp/cloud/issues/11739.
func TestSchemaValidatorNormalizesIdentifierTraits(t *testing.T) {
	t.Parallel()

	router := http.NewServeMux()
	router.HandleFunc("GET /schema/code-sms", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/code-sms.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "phone": {
          "type": "string",
          "format": "tel",
          "ory.sh/kratos": {
            "credentials": {
              "code": {
                "identifier": true,
                "via": "sms"
              }
            }
          }
        }
      }
    }
  }
}`))
	})
	router.HandleFunc("GET /schema/recovery-sms", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/recovery-sms.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "my.phone.number": {
          "type": "string",
          "format": "tel",
          "ory.sh/kratos": {
            "recovery": {
              "via": "sms"
            }
          }
        }
      }
    }
  }
}`))
	})
	router.HandleFunc("GET /schema/verification-sms", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/verification-sms.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "phone": {
          "type": "string",
          "format": "tel",
          "ory.sh/kratos": {
            "verification": {
              "via": "sms"
            }
          }
        }
      }
    }
  }
}`))
	})
	router.HandleFunc("GET /schema/verification-sms-array", func(w http.ResponseWriter, _ *http.Request) {
		// Array of phones, each element a verification+sms identifier.
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/verification-sms-array.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "phones": {
          "type": "array",
          "items": {
            "type": "string",
            "format": "tel",
            "ory.sh/kratos": {
              "verification": {
                "via": "sms"
              }
            }
          }
        }
      }
    }
  }
}`))
	})
	router.HandleFunc("GET /schema/code-sms-shared-ref", func(w http.ResponseWriter, _ *http.Request) {
		// Two trait properties share the same $defs/PhoneIdentifier
		// definition. With recursion-stack semantics both must
		// normalize; with a one-shot visited-set only the first would.
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/code-sms-shared-ref.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$defs": {
    "PhoneIdentifier": {
      "type": "string",
      "format": "tel",
      "ory.sh/kratos": {
        "credentials": {
          "code": {
            "identifier": true,
            "via": "sms"
          }
        }
      }
    }
  },
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "primary_phone": { "$ref": "#/$defs/PhoneIdentifier" },
        "secondary_phone": { "$ref": "#/$defs/PhoneIdentifier" }
      }
    }
  }
}`))
	})
	router.HandleFunc("GET /schema/code-email", func(w http.ResponseWriter, _ *http.Request) {
		// Email-channel identifier: trait must NOT be normalized
		// (case-preservation is asserted by ory/kratos#3187).
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/code-email.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "email": {
          "type": "string",
          "format": "email",
          "ory.sh/kratos": {
            "credentials": {
              "code": {
                "identifier": true,
                "via": "email"
              }
            }
          }
        }
      }
    }
  }
}`))
	})
	router.HandleFunc("GET /schema/code-sms-recursive", func(w http.ResponseWriter, _ *http.Request) {
		// Self-referential schema: `Person` has a `spouse` property
		// pointing back to `Person`. Without the recursion-stack guard
		// in walkPhoneTraits, the walk would loop forever; with it the
		// walk visits `phone` and `spouse.phone`, then stops.
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/code-sms-recursive.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$defs": {
    "Person": {
      "type": "object",
      "properties": {
        "phone": {
          "type": "string",
          "format": "tel",
          "ory.sh/kratos": {
            "credentials": {
              "code": {
                "identifier": true,
                "via": "sms"
              }
            }
          }
        },
        "spouse": { "$ref": "#/$defs/Person" }
      }
    }
  },
  "type": "object",
  "properties": {
    "traits": { "$ref": "#/$defs/Person" }
  }
}`))
	})
	router.HandleFunc("GET /schema/code-sms-anyof", func(w http.ResponseWriter, _ *http.Request) {
		// Sibling combinator coverage: the production code lumps
		// anyOf/oneOf/allOf into the same slices.Concat, but only
		// allOf is otherwise exercised. Drop the wrong key from the
		// concat and this test catches it.
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/code-sms-anyof.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "phone": {
          "anyOf": [
            {
              "type": "string",
              "format": "tel",
              "ory.sh/kratos": {
                "credentials": {
                  "code": {
                    "identifier": true,
                    "via": "sms"
                  }
                }
              }
            }
          ]
        }
      }
    }
  }
}`))
	})
	router.HandleFunc("GET /schema/code-sms-allof", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/code-sms-allof.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "phone": {
          "allOf": [
            {
              "type": "string",
              "format": "tel",
              "ory.sh/kratos": {
                "credentials": {
                  "code": {
                    "identifier": true,
                    "via": "sms"
                  }
                }
              }
            }
          ]
        }
      }
    }
  }
}`))
	})

	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	_, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(testhelpers.IdentitySchemasConfig(map[string]string{
			"code-sms":               ts.URL + "/schema/code-sms",
			"recovery-sms":           ts.URL + "/schema/recovery-sms",
			"verification-sms":       ts.URL + "/schema/verification-sms",
			"verification-sms-array": ts.URL + "/schema/verification-sms-array",
			"code-sms-allof":         ts.URL + "/schema/code-sms-allof",
			"code-sms-anyof":         ts.URL + "/schema/code-sms-anyof",
			"code-sms-recursive":     ts.URL + "/schema/code-sms-recursive",
			"code-sms-shared-ref":    ts.URL + "/schema/code-sms-shared-ref",
			"code-email":             ts.URL + "/schema/code-email",
		})),
	)
	v := NewValidator(reg)

	for _, tc := range []struct {
		name       string
		schemaID   string
		input      string
		wantTraits string
		wantErr    string
	}{
		{
			name:       "credential identifier code+sms normalizes phone in traits",
			schemaID:   "code-sms",
			input:      `{"phone":"+49 176 671 11 638"}`,
			wantTraits: `{"phone":"+4917667111638"}`,
		},
		{
			name:       "credential identifier code+sms normalizes already E.164 phone idempotently",
			schemaID:   "code-sms",
			input:      `{"phone":"+4917667111638"}`,
			wantTraits: `{"phone":"+4917667111638"}`,
		},
		{
			name:       "recovery via sms normalizes phone in traits",
			schemaID:   "recovery-sms",
			input:      `{"my.phone.number":"+49 176 671 11 638"}`,
			wantTraits: `{"my.phone.number":"+4917667111638"}`,
		},
		{
			name:       "verification via sms normalizes phone in traits",
			schemaID:   "verification-sms",
			input:      `{"phone":"+49 176 671 11 638"}`,
			wantTraits: `{"phone":"+4917667111638"}`,
		},
		{
			name:       "phone identifier nested under allOf is normalized",
			schemaID:   "code-sms-allof",
			input:      `{"phone":"+49 176 671 11 638"}`,
			wantTraits: `{"phone":"+4917667111638"}`,
		},
		{
			name:       "phone identifier nested under anyOf is normalized",
			schemaID:   "code-sms-anyof",
			input:      `{"phone":"+49 176 671 11 638"}`,
			wantTraits: `{"phone":"+4917667111638"}`,
		},
		{
			// Cycle guard: a self-referential schema (Person.spouse
			// → Person) must not send walkPhoneTraits into infinite
			// recursion. The top-level phone still normalizes; the
			// schema walk hits the spouse → Person cycle and
			// short-circuits without exploding the stack. A hang or
			// stack overflow here means the recursion guard regressed.
			name:       "recursive schema terminates and normalizes the top-level phone",
			schemaID:   "code-sms-recursive",
			input:      `{"phone":"+49 176 671 11 638"}`,
			wantTraits: `{"phone":"+4917667111638"}`,
		},
		{
			// Array elements at `phones[*]` are configured as
			// verification+sms identifiers and must be normalized.
			name:       "verification via sms normalizes phones inside an array",
			schemaID:   "verification-sms-array",
			input:      `{"phones":["+49 176 671 11 638","+44 20 7946 0958"]}`,
			wantTraits: `{"phones":["+4917667111638","+442079460958"]}`,
		},
		{
			// Two trait properties share the same $ref definition.
			// The walk must visit both, not skip the second because
			// the shared schema node was seen during the first walk.
			name:       "shared $ref between two trait paths normalizes both",
			schemaID:   "code-sms-shared-ref",
			input:      `{"primary_phone":"+49 176 671 11 638","secondary_phone":"+44 20 7946 0958"}`,
			wantTraits: `{"primary_phone":"+4917667111638","secondary_phone":"+442079460958"}`,
		},
		{
			// Regression guard for ory/kratos#3187: email-channel
			// identifiers must keep the user-supplied case in traits.
			name:       "email-channel identifier preserves trait case verbatim",
			schemaID:   "code-email",
			input:      `{"email":"UpperCased@Ory.SH"}`,
			wantTraits: `{"email":"UpperCased@Ory.SH"}`,
		},
		{
			// When the trait value is not a parseable phone number,
			// validation should fail and traits should be untouched.
			name:       "unparseable phone leaves traits untouched and surfaces validation error",
			schemaID:   "code-sms",
			input:      `{"phone":"not-a-phone-number"}`,
			wantTraits: `{"phone":"not-a-phone-number"}`,
			wantErr:    "is not valid",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.WithValue(t.Context(), httploader.ContextKey, httpx.NewResilientClient())
			i := &Identity{
				SchemaID: tc.schemaID,
				Traits:   Traits(tc.input),
			}
			err := v.Validate(ctx, i)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.JSONEq(t, tc.wantTraits, string(i.Traits),
				"trait JSON must match the normalized identifier form so webhook payloads are consistent with what Kratos stores")
		})
	}
}
