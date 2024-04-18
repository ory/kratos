package sentry

import (
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

type Hook struct {
	hub    *sentry.Hub
	levels []logrus.Level
	tags   map[string]string
	extra  map[string]interface{}
	parser Parser
}

func New(levels []logrus.Level) Hook {
	h := Hook{
		levels: levels,
		hub:    sentry.CurrentHub(),
		parser: DefaultParser,
	}

	return h
}

func (hook Hook) Levels() []logrus.Level {
	return hook.levels
}

func (hook Hook) Fire(entry *logrus.Entry) error {
	event := sentry.NewEvent()

	for k, v := range hook.extra {
		event.Extra[k] = v
	}

	for k, v := range hook.tags {
		event.Tags[k] = v
	}

	hub := hook.hub

	if entry.Context != nil {
		h := sentry.GetHubFromContext(entry.Context)
		if h != nil {
			hub = h
		}
	}

	hook.parser(entry, event)

	hub.Client().CaptureEvent(
		event,
		&sentry.EventHint{Context: entry.Context},
		hub.Scope(),
	)

	return nil
}
