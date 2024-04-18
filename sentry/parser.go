// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sentry

import (
	"encoding/json"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

var levelMap = map[logrus.Level]sentry.Level{
	logrus.TraceLevel: sentry.LevelDebug,
	logrus.DebugLevel: sentry.LevelDebug,
	logrus.InfoLevel:  sentry.LevelInfo,
	logrus.WarnLevel:  sentry.LevelWarning,
	logrus.ErrorLevel: sentry.LevelError,
	logrus.FatalLevel: sentry.LevelFatal,
	logrus.PanicLevel: sentry.LevelFatal,
}

// ErrorDescription title and subtitle of the error in Sentry
type ErrorDescription struct {
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

type Parser func(entry *logrus.Entry, event *sentry.Event)

func DefaultParser(entry *logrus.Entry, event *sentry.Event) {
	event.Level = levelMap[entry.Level]
	event.Message = entry.Message

	for k, v := range entry.Data {
		event.Extra[k] = v
	}

	errJson, err := json.Marshal(entry.Data[logrus.ErrorKey])
	if err != nil {
		return
	}

	ed := ErrorDescription{}

	err = json.Unmarshal(errJson, &ed)
	if err != nil {
		return
	}

	exception := sentry.Exception{
		Type:  ed.Reason,
		Value: ed.Message,
	}

	event.Exception = []sentry.Exception{exception}
}
