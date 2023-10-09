// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/httpx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
)

type testCSRFTokenGenerator struct{}

func (t *testCSRFTokenGenerator) GenerateCSRFToken(_ *http.Request) string {
	return "csrf_token_value"
}

// testFlow is a minimalistic flow implementation to satisfy interface and is used only in tests.
type testFlow struct {
	// ID represents the flow's unique ID.
	//
	// required: true
	ID uuid.UUID `json:"id" faker:"-" db:"id" rw:"r"`

	// Type represents the flow's type which can be either "api" or "browser", depending on the flow interaction.
	//
	// required: true
	Type Type `json:"type" db:"type" faker:"flow_type"`

	// RequestURL is the initial URL that was requested from Ory Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	//
	// required: true
	RequestURL string `json:"request_url" db:"request_url"`

	// UI contains data which must be shown in the user interface.
	//
	// required: true
	UI *container.Container `json:"ui" db:"ui"`

	// Flow State
	//
	// The state represents the state of the verification flow.
	//
	// - choose_method: ask the user to choose a method (e.g. recover account via email)
	// - sent_email: the email has been sent to the user
	// - passed_challenge: the request was successful and the recovery challenge was passed.
	//
	// required: true
	State State `json:"state" db:"state"`
}

func (t *testFlow) GetID() uuid.UUID {
	return t.ID
}

func (t *testFlow) GetType() Type {
	return t.Type
}

func (t *testFlow) GetRequestURL() string {
	return t.RequestURL
}

func (t *testFlow) AppendTo(url *url.URL) *url.URL {
	return AppendFlowTo(url, t.ID)
}

func (t *testFlow) GetUI() *container.Container {
	return t.UI
}

func (t *testFlow) GetState() State {
	return t.State
}

func (t *testFlow) GetFlowName() FlowName {
	return FlowName("test")
}

func (t *testFlow) SetState(state State) {
	t.State = state
}

func newTestFlow(r *http.Request, flowType Type) Flow {
	id := x.NewUUID()
	requestURL := x.RequestURL(r).String()
	ui := &container.Container{
		Method: "POST",
		Action: "/test",
	}

	ui.Nodes.Append(node.NewInputField("traits.username", nil, node.PasswordGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute))
	ui.Nodes.Append(node.NewInputField("traits.password", nil, node.PasswordGroup, node.InputAttributeTypePassword, node.WithRequiredInputAttribute))

	return &testFlow{
		ID:         id,
		UI:         ui,
		RequestURL: requestURL,
		Type:       flowType,
	}
}

func prepareTraits(username, password string) identity.Traits {
	payload := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{username, password}

	data, _ := json.Marshal(payload)
	return data
}

func TestHandleHookError(t *testing.T) {
	r := &http.Request{URL: &url.URL{RawQuery: ""}}
	logger := logrusx.New("kratos", "test", logrusx.ForceLevel(logrus.FatalLevel))
	l := &x.SimpleLoggerWithClient{L: logger, C: httpx.NewResilientClient(), T: otelx.NewNoop(logger, &otelx.Config{ServiceName: "kratos"})}
	csrf := testCSRFTokenGenerator{}
	f := newTestFlow(r, TypeBrowser)
	tr := prepareTraits("foo", "bar")

	t.Run("case=fill_in_traits", func(t *testing.T) {
		ve := schema.NewValidationListError([]*schema.ValidationError{schema.NewHookValidationError("traits.username", "invalid username", text.Messages{})})

		err := HandleHookError(nil, r, f, tr, node.PasswordGroup, ve, l, &csrf)
		assert.ErrorIs(t, err, ve)
		if assert.NotEmpty(t, f.GetUI()) {
			ui := f.GetUI()
			assert.Len(t, ui.Nodes, 3)
			assert.ElementsMatch(t, ui.Nodes,
				node.Nodes{
					&node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "traits.username", Type: node.InputAttributeTypeText, FieldValue: "foo", Required: true}, Meta: &node.Meta{}},
					&node.Node{Type: node.Input, Group: node.PasswordGroup, Attributes: &node.InputAttributes{Name: "traits.password", Type: node.InputAttributeTypePassword, FieldValue: "bar", Required: true}, Meta: &node.Meta{}},
					&node.Node{Type: node.Input, Group: node.DefaultGroup, Attributes: &node.InputAttributes{Name: "csrf_token", Type: node.InputAttributeTypeHidden, FieldValue: "csrf_token_value", Required: true}},
				})
		}
	})

	t.Run("case=unmarshal_fail", func(t *testing.T) {
		ve := schema.NewValidationListError([]*schema.ValidationError{schema.NewHookValidationError("traits.username", "invalid username", text.Messages{})})

		err := HandleHookError(nil, r, f, []byte("garbage"), node.PasswordGroup, ve, l, &csrf)
		var jsonErr *json.SyntaxError
		assert.ErrorAs(t, err, &jsonErr)
	})
}
