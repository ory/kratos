package x

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/gddo/httputil"
	"github.com/stretchr/testify/assert"
)

func TestIsBrowserOrAPIRequest(t *testing.T) {
	for k, tc := range []struct {
		ua string
		h  string
		e  bool
	}{
		{ua: "firefox-66", h: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8", e: true},
		{ua: "safari-chrome", h: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8", e: true},
		{ua: "ie8", h: "image/jpeg,application/x-ms-application,image/gif,application/xaml+xml,image/pjpeg,application/x-ms-xbap,application/x-shockwave-flash,application/msword,*/*", e: true},
		{ua: "ie8-any", h: "*/*", e: true},
		{ua: "edge", h: "text/html,application/xhtml+xml,image/jxr,*/*", e: true},
		{ua: "opera", h: "text/html,application/xml;q=0.9,application/xhtml+xml,image/png,image/webp,image/jpeg,image/gif,image/x-xbitmap,*/*;q=0.1", e: true},
		{ua: "json-api", h: "application/json", e: false},
		{ua: "no-accept", h: "", e: true},
	} {
		t.Run(fmt.Sprintf("case=%d/ua=%s", k, tc.ua), func(t *testing.T) {
			r := &http.Request{Header: map[string][]string{"Accept": {tc.h}}}
			t.Logf("isBrowser: %s", httputil.NegotiateContentType(r, offers, defaultOffer))

			t.Logf("isJSON: %s", httputil.NegotiateContentType(r,
				[]string{"application/json"},
				"text/html",
			))

			assert.Equal(t, tc.e, IsBrowserRequest(r))
			assert.Equal(t, !tc.e, IsJSONRequest(r))
		})
	}
}
