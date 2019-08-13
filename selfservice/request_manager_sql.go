package selfservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/ory/x/jsonx"

	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"

	"github.com/ory/herodot"

	"github.com/ory/hive/identity"
)

var _ RegistrationRequestManager = new(RequestManagerSQL)
var _ LoginRequestManager = new(RequestManagerSQL)

const requestSQLTableName = "self_service_request"

type RequestManagerSQL struct {
	db  *sqlx.DB
	rmf map[identity.CredentialsType]func() RequestMethodConfig
}

type requestSQL struct {
	ID             string          `db:"id"`
	IssuedAt       time.Time       `db:"issued_at"`
	ExpiresAt      time.Time       `db:"expires_at"`
	RequestURL     string          `db:"request_url"`
	RequestHeaders json.RawMessage `db:"request_headers"`
	Active         string          `db:"active"`
	Methods        json.RawMessage `db:"methods"`
	Kind           string          `db:"kind"`
}

func newRequestSQL(r *Request, kind string) (*requestSQL, error) {
	var requestHeaders bytes.Buffer
	var methods bytes.Buffer

	if err := json.NewEncoder(&requestHeaders).Encode(r.RequestHeaders); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := json.NewEncoder(&methods).Encode(r.Methods); err != nil {
		return nil, errors.WithStack(err)
	}

	return &requestSQL{
		ID:             r.ID,
		IssuedAt:       r.IssuedAt,
		ExpiresAt:      r.ExpiresAt,
		RequestURL:     r.RequestURL,
		RequestHeaders: requestHeaders.Bytes(),
		Active:         string(r.Active),
		Methods:        methods.Bytes(),
		Kind:           kind,
	}, nil
}

func NewRequestManagerSQL(db *sqlx.DB, factories map[identity.CredentialsType]func() RequestMethodConfig) *RequestManagerSQL {
	return &RequestManagerSQL{db: db, rmf: factories}
}

func (m *RequestManagerSQL) CreateLoginRequest(ctx context.Context, r *LoginRequest) error {
	return m.cr(ctx, r.Request, "login")
}

func (m *RequestManagerSQL) CreateRegistrationRequest(ctx context.Context, r *RegistrationRequest) error {
	return m.cr(ctx, r.Request, "registration")
}

func (m *RequestManagerSQL) GetLoginRequest(ctx context.Context, id string) (*LoginRequest, error) {
	r, err := m.gr(ctx, id, "login")
	if err != nil {
		return nil, err
	}

	return &LoginRequest{Request: r}, nil
}

func (m *RequestManagerSQL) GetRegistrationRequest(ctx context.Context, id string) (*RegistrationRequest, error) {
	r, err := m.gr(ctx, id, "registration")
	if err != nil {
		return nil, err
	}

	return &RegistrationRequest{Request: r}, nil
}

func (m *RequestManagerSQL) UpdateRegistrationRequest(ctx context.Context, id string, t identity.CredentialsType, c RequestMethodConfig) error {
	r, err := m.GetRegistrationRequest(ctx, id)
	if err != nil {
		return err
	}

	return m.ur(ctx, r.Request, t, c, "registration")
}

func (m *RequestManagerSQL) UpdateLoginRequest(ctx context.Context, id string, t identity.CredentialsType, c RequestMethodConfig) error {
	r, err := m.GetLoginRequest(ctx, id)
	if err != nil {
		return err
	}

	return m.ur(ctx, r.Request, t, c, "login")
}

func (m *RequestManagerSQL) ur(ctx context.Context, r *Request, t identity.CredentialsType, c RequestMethodConfig, kind string) error {
	me, ok := r.Methods[t]
	if !ok {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Expected %s request "%s" to have credentials type "%s", indicating an internal error.`, kind, r.ID, t))
	}

	me.Config = c
	r.Active = t
	r.Methods[t] = me

	update, err := newRequestSQL(r, kind)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE %s SET %s", requestSQLTableName, sqlxx.NamedUpdateArguments(update, "id", "issued_at", "expires_at", "request_url", "headers"))
	if _, err := m.db.NamedExecContext(ctx, m.db.Rebind(query), update); err != nil {
		return sqlcon.HandleError(err)
	}

	return nil
}
func (m *RequestManagerSQL) cr(ctx context.Context, r *Request, kind string) error {
	insert, err := newRequestSQL(r, kind)
	if err != nil {
		return err
	}

	columns, arguments := sqlxx.NamedInsertArguments(insert)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", requestSQLTableName, columns, arguments)
	if _, err := m.db.NamedExecContext(ctx, m.db.Rebind(query), insert); err != nil {
		return sqlcon.HandleError(err)
	}

	return nil
}

func (m *RequestManagerSQL) gr(ctx context.Context, id string, kind string) (*Request, error) {
	var r requestSQL

	columns, _ := sqlxx.NamedInsertArguments(r, "pk")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? AND kind=?", columns, requestSQLTableName)
	if err := sqlcon.HandleError(m.db.GetContext(ctx, &r, m.db.Rebind(query), id, kind)); err != nil {
		if errors.Cause(err) == sqlcon.ErrNoRows {
			return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("%s", err))
		}
		return nil, err
	}

	var header http.Header
	if err := jsonx.NewStrictDecoder(bytes.NewBuffer(r.RequestHeaders)).Decode(&header); err != nil {
		return nil, errors.WithStack(err)
	}

	var methodsRaw map[string]json.RawMessage
	if err := jsonx.NewStrictDecoder(bytes.NewBuffer(r.Methods)).Decode(&methodsRaw); err != nil {
		return nil, errors.WithStack(err)
	}

	methods := map[identity.CredentialsType]*DefaultRequestMethod{}
	for method, raw := range methodsRaw {
		ct := identity.CredentialsType(method)
		var config DefaultRequestMethod
		prototype, found := m.rmf[ct]
		if !found {
			panic(fmt.Sprintf("unknown credentials type: %s", method))
		}

		config.Config = prototype()
		if err := jsonx.NewStrictDecoder(bytes.NewBuffer(raw)).Decode(&config); err != nil {
			return nil, errors.WithStack(err)
		}
		methods[ct] = &config
	}

	return &Request{
		ID:             r.ID,
		IssuedAt:       r.IssuedAt.UTC(),
		ExpiresAt:      r.ExpiresAt.UTC(),
		RequestURL:     r.RequestURL,
		RequestHeaders: header,
		Active:         identity.CredentialsType(r.Active),
		Methods:        methods,
	}, nil
}
