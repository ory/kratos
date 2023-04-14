// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/x"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/snapshotx"
	"github.com/ory/x/urlx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultPageToken = new(courier.Message).DefaultPageToken().Encode()

func TestHandler(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	// Start kratos server
	publicTS, adminTS := testhelpers.NewKratosServerWithCSRF(t, reg)

	tss := []struct {
		name string
		s    *httptest.Server
	}{
		{
			name: "public",
			s:    publicTS,
		},
		{
			name: "admin",
			s:    adminTS,
		},
	}

	mockServerURL := urlx.ParseOrPanic(publicTS.URL)
	conf.MustSet(ctx, config.ViperKeyAdminBaseURL, adminTS.URL)
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, mockServerURL.String())

	var get = func(t *testing.T, base *httptest.Server, href string, expectCode int) gjson.Result {
		t.Helper()
		res, err := base.Client().Get(base.URL + href)
		require.NoError(t, err)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		assert.EqualValuesf(t, expectCode, res.StatusCode, "%s", body)
		return gjson.ParseBytes(body)
	}

	var getList = func(t *testing.T, tsName string, qs string) gjson.Result {
		t.Helper()
		href := courier.AdminRouteListMessages + qs
		ts := adminTS

		if tsName == "public" {
			href = x.AdminPrefix + href
			ts = publicTS
		}

		parsed := get(t, ts, href, http.StatusOK)
		require.Truef(t, parsed.IsArray(), "%s", parsed.Raw)
		return parsed
	}

	t.Run("case=should return an empty list of messages", func(t *testing.T) {
		for _, name := range []string{"public", "admin"} {
			t.Run("endpoint="+name, func(t *testing.T) {
				parsed := getList(t, name, "")
				assert.Len(t, parsed.Array(), 0)
			})
		}
	})

	t.Run("case=list messages", func(t *testing.T) {
		// Arrange test data
		const msgCount = 10    // total message count
		const procCount = 5    // how many messages' status should be equal to `processing`
		const rcptOryCount = 2 // how many messages' recipient should be equal to `noreply@ory.sh`
		messages := make([]courier.Message, msgCount)

		for i := range messages {
			require.NoError(t, faker.FakeData(&messages[i]))
			messages[i].Type = courier.MessageTypeEmail
			messages[i].Body = "body content"
			if i < rcptOryCount {
				messages[i].Recipient = "noreply@ory.sh"
			}
			require.NoError(t, reg.CourierPersister().AddMessage(context.Background(), &messages[i]))
		}
		for i := 0; i < procCount; i++ {
			require.NoError(t, reg.CourierPersister().SetMessageStatus(context.Background(), messages[i].ID, courier.MessageStatusProcessing))
		}

		t.Run("paging", func(t *testing.T) {
			t.Run("case=should return half of the messages", func(t *testing.T) {
				qs := fmt.Sprintf("?page_token=%s&page_size=%d", defaultPageToken, msgCount/2)

				for _, tc := range tss {
					t.Run("endpoint="+tc.name, func(t *testing.T) {
						parsed := getList(t, tc.name, qs)
						assert.Len(t, parsed.Array(), msgCount/2)
					})
				}
			})
			t.Run("case=should return no message", func(t *testing.T) {
				token := keysetpagination.MapPageToken{
					"id":         "1232",
					"created_at": time.Now().Add(time.Duration(-10) * time.Hour).Format("2006-01-02 15:04:05.99999-07:00"),
				}
				qs := fmt.Sprintf(`?page_token=%s&page_size=%s`, token.Encode(), "250")

				for _, tc := range tss {
					t.Run("endpoint="+tc.name, func(t *testing.T) {
						parsed := getList(t, tc.name, qs)
						assert.Len(t, parsed.Array(), 0)
					})
				}
			})
		})
		t.Run("filtering", func(t *testing.T) {
			t.Run("case=should return all queued messages", func(t *testing.T) {
				qs := fmt.Sprintf(`?page_token=%s&page_size=250&status=queued`, defaultPageToken)

				for _, tc := range tss {
					t.Run("endpoint="+tc.name, func(t *testing.T) {
						parsed := getList(t, tc.name, qs)
						assert.Len(t, parsed.Array(), msgCount-procCount)

						for _, item := range parsed.Array() {
							assert.Equal(t, "queued", item.Get("status").String())
						}
					})
				}
			})
			t.Run("case=should return all processing messages", func(t *testing.T) {
				qs := fmt.Sprintf(`?page_token=%s&page_size=250&status=processing`, defaultPageToken)

				for _, tc := range tss {
					t.Run("endpoint="+tc.name, func(t *testing.T) {
						parsed := getList(t, tc.name, qs)
						assert.Len(t, parsed.Array(), procCount)

						for _, item := range parsed.Array() {
							assert.Equal(t, "processing", item.Get("status").String())
						}
					})
				}
			})
			t.Run("case=should return all messages with recipient equals to noreply@ory.sh", func(t *testing.T) {
				qs := fmt.Sprintf(`?page_token=%s&page_size=250&recipient=noreply@ory.sh`, defaultPageToken)

				for _, tc := range tss {
					t.Run("endpoint="+tc.name, func(t *testing.T) {
						parsed := getList(t, tc.name, qs)
						assert.Len(t, parsed.Array(), rcptOryCount)

						for _, item := range parsed.Array() {
							assert.Equal(t, "noreply@ory.sh", item.Get("recipient").String())
						}
					})
				}
			})
		})
		t.Run("case=body should be redacted if kratos is not in dev mode", func(t *testing.T) {
			conf.MustSet(ctx, "dev", false)
			for _, tc := range tss {
				t.Run("endpoint="+tc.name, func(t *testing.T) {
					parsed := getList(t, tc.name, "")
					require.Lenf(t, parsed.Array(), msgCount, "%s", parsed.Raw)

					for _, item := range parsed.Array() {
						assert.Equal(t, "<redacted-unless-dev-mode>", item.Get("body").String())
					}
				})
			}
		})
		t.Run("case=body should not be redacted if kratos is in dev mode", func(t *testing.T) {
			conf.MustSet(ctx, "dev", true)
			for _, tc := range tss {
				t.Run("endpoint="+tc.name, func(t *testing.T) {
					parsed := getList(t, tc.name, "")
					require.Lenf(t, parsed.Array(), msgCount, "%s", parsed.Raw)

					for _, item := range parsed.Array() {
						assert.Equal(t, "body content", item.Get("body").String())
					}
				})
			}
		})
		t.Run("case=should return with http status BadRequest when given status is invalid", func(t *testing.T) {
			qs := fmt.Sprintf(`?page_token=%s&page_size=250&status=invalid_status`, defaultPageToken)

			res, err := adminTS.Client().Get(adminTS.URL + courier.AdminRouteListMessages + qs)

			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode, "status code should be equal to StatusBadRequest")
		})

	})
	t.Run("handler=getCourierMessage", func(t *testing.T) {

		message := courier.Message{}
		require.NoError(t, faker.FakeData(&message))
		message.Type = courier.MessageTypeEmail
		message.Body = "body content"
		require.NoError(t, reg.CourierPersister().AddMessage(context.Background(), &message))
		require.NoError(t, reg.CourierPersister().RecordDispatch(ctx, message.ID, courier.CourierMessageDispatchStatusSuccess, errors.New("some error")))

		getCourierMessag := func(s *httptest.Server, id string) gjson.Result {

			r, err := s.Client().Get(s.URL + "/admin/courier/messages/" + id)
			require.NoError(t, err)
			return gjson.ParseBytes(ioutilx.MustReadAll(r.Body))
		}

		t.Run("case=should return a message by id", func(t *testing.T) {
			conf.MustSet(ctx, "dev", true)

			for _, tc := range tss {
				t.Run("endpoint="+tc.name, func(t *testing.T) {
					body := getCourierMessag(tc.s, message.ID.String())
					assert.Equal(t, message.ID.String(), body.Get("id").String())
					assert.Equal(t, message.Recipient, body.Get("recipient").String())
					assert.Equal(t, message.Body, body.Get("body").String())

					// assert Eager works
					assert.NotEmpty(t, body.Get("dispatches").Array())
				})
			}
		})
		t.Run("case=does not contain body if not in production", func(t *testing.T) {
			conf.MustSet(ctx, "dev", false)

			for _, tc := range tss {
				t.Run("endpoint="+tc.name, func(t *testing.T) {
					body := getCourierMessag(tc.s, message.ID.String())
					assert.Equal(t, message.ID.String(), body.Get("id").String())
					assert.Equal(t, "<redacted-unless-dev-mode>", body.Get("body").String())
				})
			}
		})
		t.Run("case=returns an error if parameter is malformed", func(t *testing.T) {
			for _, tc := range tss {
				t.Run("endpoint="+tc.name, func(t *testing.T) {
					body := getCourierMessag(tc.s, "not-a-uuid")

					snapshotx.SnapshotTJSONString(t, body.String())
				})
			}
		})
		t.Run("case=returns an error if no message is found", func(t *testing.T) {
			for _, tc := range tss {
				t.Run("endpoint="+tc.name, func(t *testing.T) {
					body := getCourierMessag(tc.s, uuid.Nil.String())
					snapshotx.SnapshotTJSONString(t, body.String())
				})
			}
		})
	})
}
