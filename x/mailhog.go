// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/pat"
	"github.com/ian-kent/go-log/levels"
	"github.com/ian-kent/go-log/log"
	"github.com/mailhog/MailHog-Server/api"
	mailhogconf "github.com/mailhog/MailHog-Server/config"
	"github.com/mailhog/MailHog-Server/monkey"
	"github.com/mailhog/MailHog-Server/smtp"
	"github.com/mailhog/data"
	"github.com/mailhog/storage"
	"github.com/stretchr/testify/require"
)

// StartMailhog starts a MailHog server for testing purposes.
// It returns the SMTP connection URL and the API URL.
// If withChaosMonkey is true, the SMTP server will randomly drop connections and simulate network issues.
func StartMailhog(t testing.TB, withChaosMonkey bool) (smtpAddr, apiAddr string) {
	t.Helper()

	// hacky but should silence most MailHog logs during tests
	log.Logger().SetLevel(levels.FATAL)
	stdlog.Default().SetOutput(io.Discard)

	apiconf := &mailhogconf.Config{
		Storage:     storage.CreateInMemory(),
		MessageChan: make(chan *data.Message),
	}
	if withChaosMonkey {
		jim := &monkey.Jim{
			DisconnectChance: 0.005,
			AcceptChance:     0.99,
			LinkSpeedAffect:  0.05,
			LinkSpeedMin:     1250,
			LinkSpeedMax:     12500,
			RejectAuthChance: 0.05,

			// important: set to 0 to avoid flakes in tests, because those errors are not retryable
			RejectSenderChance:    0,
			RejectRecipientChance: 0,
		}
		jim.Configure(t.Logf)
		apiconf.Monkey = jim
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	go serveSMTP(t.Context(), ln, apiconf)

	r := pat.New()
	api.CreateAPI(apiconf, r)
	s := httptest.NewServer(r)
	t.Cleanup(s.Close)

	return fmt.Sprintf("smtp://%s?disable_starttls=true", ln.Addr().String()), s.URL
}

func serveSMTP(ctx context.Context, ln net.Listener, cfg *mailhogconf.Config) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				fmt.Printf("[SMTP] Error accepting connection: %s\n", err)
				continue
			}

			if cfg.Monkey != nil {
				ok := cfg.Monkey.Accept(conn)
				if !ok {
					_ = conn.Close()
					continue
				}
			}

			go smtp.Accept(
				conn.(*net.TCPConn).RemoteAddr().String(),
				io.ReadWriteCloser(conn),
				cfg.Storage,
				cfg.MessageChan,
				cfg.Hostname,
				cfg.Monkey,
			)
		}
	}
}
