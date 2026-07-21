// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier_test

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/courier/template"
	templates "github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/pkg"
	"github.com/ory/x/configx"
	"github.com/ory/x/sqlxx"
)

// smtpRecorder captures the raw SMTP commands and DATA lines a client sends.
type smtpRecorder struct {
	mu   sync.Mutex
	cmds []string
	data []string
}

func (r *smtpRecorder) snapshot() (cmds, data []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.cmds...), append([]string(nil), r.data...)
}

// newRecordingSMTPServer starts a minimal SMTP server that accepts everything
// and records the raw protocol. It returns a courier SMTP connection URI.
func newRecordingSMTPServer(t *testing.T) (string, *smtpRecorder) {
	t.Helper()
	rec := &smtpRecorder{}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = l.Close() })

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer func() { _ = conn.Close() }()
				_, _ = fmt.Fprintf(conn, "220 fake ESMTP\r\n")
				sc := bufio.NewScanner(conn)
				inData := false
				for sc.Scan() {
					line := sc.Text()
					if inData {
						if line == "." {
							inData = false
							_, _ = fmt.Fprintf(conn, "250 ok\r\n")
							continue
						}
						rec.mu.Lock()
						rec.data = append(rec.data, line)
						rec.mu.Unlock()
						continue
					}
					rec.mu.Lock()
					rec.cmds = append(rec.cmds, line)
					rec.mu.Unlock()
					switch verb := strings.ToUpper(line); {
					case strings.HasPrefix(verb, "EHLO"), strings.HasPrefix(verb, "HELO"):
						_, _ = fmt.Fprintf(conn, "250-fake\r\n250 8BITMIME\r\n")
					case strings.HasPrefix(verb, "DATA"):
						inData = true
						_, _ = fmt.Fprintf(conn, "354 go ahead\r\n")
					case strings.HasPrefix(verb, "QUIT"):
						_, _ = fmt.Fprintf(conn, "221 bye\r\n")
						return
					default:
						_, _ = fmt.Fprintf(conn, "250 ok\r\n")
					}
				}
			}(conn)
		}
	}()

	return fmt.Sprintf("smtp://%s/?disable_starttls=true", l.Addr().String()), rec
}

func rcptLines(cmds []string) []string {
	var out []string
	for _, c := range cmds {
		if strings.HasPrefix(strings.ToUpper(c), "RCPT TO:") {
			out = append(out, c)
		}
	}
	return out
}

func headerLines(data []string, prefix string) []string {
	var out []string
	for _, d := range data {
		if d == "" {
			break // End of headers.
		}
		if strings.HasPrefix(d, prefix) {
			out = append(out, d)
		}
	}
	return out
}

func newWireCourier(t *testing.T, smtpURL string, extra map[string]any) (courier.Courier, *config.Config) {
	t.Helper()
	values := map[string]any{
		config.ViperKeyCourierSMTPURL:              smtpURL,
		config.ViperKeyCourierSMTPFrom:             "from@ory.sh",
		config.ViperKeyCourierMessageRetries:       5,
		config.ViperKeyClientSMTPNoPrivateIPRanges: false,
	}
	for k, v := range extra {
		values[k] = v
	}
	conf, reg := pkg.NewRegistryDefaultWithDSN(t, "", configx.WithValues(values))
	c, err := reg.Courier(t.Context())
	require.NoError(t, err)
	c.FailOnDispatchError()
	return c, conf
}

func TestDispatchSMTPWireProtocol(t *testing.T) {
	t.Run("case=rfc recipient keeps Cc and Bcc header recipients in the envelope", func(t *testing.T) {
		smtpURL, rec := newRecordingSMTPServer(t)
		c, _ := newWireCourier(t, smtpURL, map[string]any{
			config.ViperKeyCourierSMTPHeaders + ".Bcc": "archive@ory.sh",
		})

		_, err := c.QueueEmail(t.Context(), templates.NewTestStub(&templates.TestStubModel{
			To: "user@example.org", Subject: "s", Body: "b",
		}))
		require.NoError(t, err)
		require.NoError(t, c.DispatchQueue(t.Context()))

		require.EventuallyWithT(t, func(t *assert.CollectT) {
			cmds, _ := rec.snapshot()
			rcpts := rcptLines(cmds)
			assert.Contains(t, rcpts, "RCPT TO:<user@example.org>")
			assert.Contains(t, rcpts, "RCPT TO:<archive@ory.sh>")
		}, 5*time.Second, 50*time.Millisecond)
	})

	t.Run("case=non-rfc recipient produces exactly one RCPT and one To header", func(t *testing.T) {
		smtpURL, rec := newRecordingSMTPServer(t)
		c, _ := newWireCourier(t, smtpURL, nil)

		_, err := c.QueueEmail(t.Context(), templates.NewTestStub(&templates.TestStubModel{
			To: "foo.@docomo.ne.jp", Subject: "s", Body: "b",
		}))
		require.NoError(t, err)
		require.NoError(t, c.DispatchQueue(t.Context()))

		require.EventuallyWithT(t, func(t *assert.CollectT) {
			cmds, data := rec.snapshot()
			assert.Equal(t, []string{"RCPT TO:<foo.@docomo.ne.jp>"}, rcptLines(cmds))
			assert.Equal(t, []string{"To: foo.@docomo.ne.jp"}, headerLines(data, "To:"))
		}, 5*time.Second, 50*time.Millisecond)
	})

	t.Run("case=poisoned recipient is abandoned at dispatch, nothing hits the wire", func(t *testing.T) {
		smtpURL, rec := newRecordingSMTPServer(t)
		_, reg := pkg.NewRegistryDefaultWithDSN(t, "", configx.WithValues(map[string]any{
			config.ViperKeyCourierSMTPURL:              smtpURL,
			config.ViperKeyCourierSMTPFrom:             "from@ory.sh",
			config.ViperKeyCourierMessageRetries:       5,
			config.ViperKeyClientSMTPNoPrivateIPRanges: false,
		}))
		c, err := reg.Courier(t.Context())
		require.NoError(t, err)

		// Insert a row with a smuggled header directly, bypassing QueueEmail's
		// validation, to simulate a poisoned message from another path.
		require.NoError(t, reg.CourierPersister().AddMessage(t.Context(), &courier.Message{
			Type:         courier.MessageTypeEmail,
			Recipient:    "victim@example.org\r\nBcc: attacker@evil.com",
			Body:         "b",
			Subject:      "s",
			TemplateType: template.TypeTestStub,
			Channel:      sqlxx.NullString("email"),
		}))

		// Dispatch the message directly. DispatchMessage (unlike DispatchQueue)
		// does not reset a failed message back to "queued", so the abandonment
		// is observable.
		queued, err := reg.CourierPersister().LatestQueuedMessage(t.Context())
		require.NoError(t, err)
		require.Error(t, c.DispatchMessage(t.Context(), *queued))

		var got courier.Message
		require.NoError(t, reg.Persister().GetConnection(t.Context()).Where("id = ?", queued.ID).First(&got))
		assert.Equal(t, courier.MessageStatusAbandoned, got.Status)

		cmds, data := rec.snapshot()
		assert.Empty(t, rcptLines(cmds), "poisoned recipient must not produce a RCPT")
		assert.Empty(t, headerLines(data, "Bcc:"), "poisoned recipient must not smuggle a Bcc header")
		for _, line := range data {
			assert.NotContains(t, line, "attacker@evil.com", "attacker address must never reach the wire")
		}
	})
}
