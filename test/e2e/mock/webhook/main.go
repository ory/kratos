// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

const AuthHeader = "X-Authorize-Request"

type (
	// copied from selfservice/hook/web_hook.go
	detailedMessage struct {
		ID      int             `json:"id"`
		Text    string          `json:"text"`
		Type    string          `json:"type"`
		Context json.RawMessage `json:"context,omitempty"`
	}

	errorMessage struct {
		InstancePtr string            `json:"instance_ptr"`
		Messages    []detailedMessage `json:"messages"`
	}

	rawHookResponse struct {
		Messages []errorMessage `json:"messages"`
	}

	logResponseWriter struct {
		Status int
		Size   int
		http.ResponseWriter
	}
)

// Header returns & satisfies the http.ResponseWriter interface
func (w *logResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Write satisfies the http.ResponseWriter interface and
// captures data written, in bytes
func (w *logResponseWriter) Write(data []byte) (int, error) {

	written, err := w.ResponseWriter.Write(data)
	w.Size += written

	return written, err
}

// WriteHeader satisfies the http.ResponseWriter interface and
// allows us to catch the status code
func (w *logResponseWriter) WriteHeader(statusCode int) {

	w.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func accessLog(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{"application": "webhooks", "method": r.Method, "path": r.URL.Path}).Info("incoming request")
		responseWriter := logResponseWriter{http.StatusOK, 0, w}
		next.ServeHTTP(&responseWriter, r)
		log.WithFields(log.Fields{"application": "webhooks", "status": responseWriter.Status, "size": responseWriter.Size, "path": r.URL.Path}).Info("response generated")
	}

	return fn
}

func headerAuth(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(AuthHeader) != "1" {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			next.ServeHTTP(w, r)
		}
	}

	return fn
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload := struct {
		IdentityId string          `json:"identity_id,omitempty"`
		Email      string          `json:"email,omitempty"`
		FlowId     string          `json:"flow_id"`
		FlowType   string          `json:"flow_type"`
		Context    json.RawMessage `json:"ctx"`
	}{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WithError(err).Warn("could not read request body")
		b := bytes.NewBufferString(fmt.Sprintf("error while reading request body: %s", err))
		_, _ = w.Write(b.Bytes())
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WithError(err).Warn("could not unmarshal request JSON")
		b := bytes.NewBufferString(fmt.Sprintf("error while parsing request JSON: %s", err))
		_, _ = w.Write(b.Bytes())
		return
	}

	log.WithField("payload", string(body)).Info("unmarshalled request")

	if !strings.Contains(payload.Email, "_blocked@ory.sh") || payload.FlowType == "api" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusConflict)

	detail := detailedMessage{
		ID:   1234,
		Type: "error",
		Text: "email could not be validated",
	}
	msg := errorMessage{InstancePtr: "#/traits/email", Messages: []detailedMessage{detail}}
	resp := rawHookResponse{Messages: []errorMessage{msg}}
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		buff := bytes.NewBufferString(err.Error())
		_, _ = w.Write(buff.Bytes())
		return
	}
}

func writeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.WithFields(log.Fields{"application": "webhooks", "path": r.URL.Path, "response": r.URL.Query().Get("response")}).Info("sending response")
	_, _ = w.Write([]byte(r.URL.Query().Get("response")))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthCheck)
	mux.HandleFunc("/webhook", accessLog(headerAuth(webhookHandler)))
	mux.HandleFunc("/webhook/write", accessLog(writeWebhookHandler))

	s := http.Server{
		Addr:    ":4459",
		Handler: mux,
	}

	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
