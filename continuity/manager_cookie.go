package continuity

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var _ Manager = new(ManagerCookie)
var ErrNotResumable = *herodot.ErrBadRequest.WithError("no resumable session found").WithReasonf("The browser does not contain the neccesary cookie to resume the session. This is a security violation and was thus blocked. Please clear your browser's cookies and cache and try again!")

const CookieName = "ory_kratos_continuity"

type (
	managerCookieDependencies interface {
		PersistenceProvider
		x.CookieProvider
		session.ManagementProvider
	}
	ManagerCookie struct {
		d managerCookieDependencies
	}
)

func NewManagerCookie(d managerCookieDependencies) *ManagerCookie {
	return &ManagerCookie{d: d}
}

func (m *ManagerCookie) Pause(ctx context.Context, w http.ResponseWriter, r *http.Request, name string, opts ...ManagerOption) error {
	if len(name) == 0 {
		return errors.Errorf("continuity container name must be set")
	}

	o, err := newManagerOptions(opts)
	if err != nil {
		return err
	}
	c := NewContainer(name, *o)

	if err := x.SessionPersistValues(w, r, m.d.ContinuityCookieManager(ctx), CookieName, map[string]interface{}{
		name: c.ID.String(),
	}); err != nil {
		return err
	}

	if err := m.d.ContinuityPersister().SaveContinuitySession(ctx, c); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (m *ManagerCookie) Continue(ctx context.Context, w http.ResponseWriter, r *http.Request, name string, opts ...ManagerOption) (*Container, error) {
	container, err := m.container(ctx, w, r, name)
	if err != nil {
		return nil, err
	}

	o, err := newManagerOptions(opts)
	if err != nil {
		return nil, err
	}

	if err := container.Valid(o.iid); err != nil {
		return nil, err
	}

	if o.payloadRaw != nil && container.Payload != nil {
		if err := json.NewDecoder(bytes.NewBuffer(container.Payload)).Decode(o.payloadRaw); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if err := x.SessionUnsetKey(w, r, m.d.ContinuityCookieManager(ctx), CookieName, name); err != nil {
		return nil, err
	}

	if err := m.d.ContinuityPersister().DeleteContinuitySession(ctx, container.ID); err != nil && !errors.Is(err, sqlcon.ErrNoRows) {
		return nil, err
	}

	return container, nil
}

func (m *ManagerCookie) sid(ctx context.Context, w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, error) {
	var sid uuid.UUID
	if s, err := x.SessionGetString(r, m.d.ContinuityCookieManager(ctx), CookieName, name); err != nil {
		_ = x.SessionUnsetKey(w, r, m.d.ContinuityCookieManager(ctx), CookieName, name)
		return sid, errors.WithStack(ErrNotResumable.WithDebugf("%+v", err))
	} else if sid = x.ParseUUID(s); sid == uuid.Nil {
		_ = x.SessionUnsetKey(w, r, m.d.ContinuityCookieManager(ctx), CookieName, name)
		return sid, errors.WithStack(ErrNotResumable.WithDebug("session id is not a valid uuid"))
	}

	return sid, nil
}

func (m *ManagerCookie) container(ctx context.Context, w http.ResponseWriter, r *http.Request, name string) (*Container, error) {
	sid, err := m.sid(ctx, w, r, name)
	if err != nil {
		return nil, err
	}

	container, err := m.d.ContinuityPersister().GetContinuitySession(ctx, sid)
	// If an error happens, we need to clean up the cookie.
	if err != nil {
		_ = x.SessionUnsetKey(w, r, m.d.ContinuityCookieManager(ctx), CookieName, name)
	}

	if errors.Is(err, sqlcon.ErrNoRows) {
		return nil, errors.WithStack(ErrNotResumable.WithDebugf("Resumable ID from cookie could not be found in the datastore: %+v", err))
	} else if err != nil {
		return nil, err
	}

	return container, err
}

func (m ManagerCookie) Abort(ctx context.Context, w http.ResponseWriter, r *http.Request, name string) error {
	sid, err := m.sid(ctx, w, r, name)
	if errors.Is(err, &ErrNotResumable) {
		// We do not care about an error here
		return nil
	} else if err != nil {
		return err
	}

	if err := x.SessionUnsetKey(w, r, m.d.ContinuityCookieManager(ctx), CookieName, name); err != nil {
		return err
	}

	if err := m.d.ContinuityPersister().DeleteContinuitySession(ctx, sid); err != nil && !errors.Is(err, sqlcon.ErrNoRows) {
		return errors.WithStack(err)
	}

	return nil
}
