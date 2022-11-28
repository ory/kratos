// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/x"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/urlx"
	"github.com/ory/x/uuidx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultPageToken = url.QueryEscape(new(courier.Message).DefaultPageToken())

func TestHandler(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	// Start kratos server
	publicTS, adminTS := testhelpers.NewKratosServerWithCSRF(t, reg)

	mockServerURL := urlx.ParseOrPanic(publicTS.URL)
	conf.MustSet(ctx, config.ViperKeyAdminBaseURL, adminTS.URL)
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, mockServerURL.String())

	var get = func(t *testing.T, base *httptest.Server, href string, expectCode int) gjson.Result {
		t.Helper()
		res, err := base.Client().Get(base.URL + href)
		require.NoError(t, err)
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		assert.EqualValues(t, expectCode, res.StatusCode, "%s", body)
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
		require.True(t, parsed.IsArray(), "%s", parsed.Raw)
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

		tss := [...]string{"public", "admin"}

		t.Run("paging", func(t *testing.T) {
			t.Run("case=should return half of the messages", func(t *testing.T) {
				qs := fmt.Sprintf("?page_token=%s&page_size=%d", defaultPageToken, msgCount/2)

				for _, name := range tss {
					t.Run("endpoint="+name, func(t *testing.T) {
						parsed := getList(t, name, qs)
						assert.Len(t, parsed.Array(), msgCount/2)
					})
				}
			})
			t.Run("case=should return no message", func(t *testing.T) {
				qs := `?page_token=id=1232&page_size=250`

				for _, name := range tss {
					t.Run("endpoint="+name, func(t *testing.T) {
						parsed := getList(t, name, qs)
						assert.Len(t, parsed.Array(), 0)
					})
				}
			})
		})
		t.Run("filtering", func(t *testing.T) {
			t.Run("case=should return all queued messages", func(t *testing.T) {
				qs := fmt.Sprintf(`?page_token=%s&page_size=250&status=queued`, defaultPageToken)

				for _, name := range tss {
					t.Run("endpoint="+name, func(t *testing.T) {
						parsed := getList(t, name, qs)
						assert.Len(t, parsed.Array(), msgCount-procCount)

						for _, item := range parsed.Array() {
							assert.Equal(t, "queued", item.Get("status").String())
						}
					})
				}
			})
			t.Run("case=should return all processing messages", func(t *testing.T) {
				qs := fmt.Sprintf(`?page_token=%s&page_size=250&status=processing`, defaultPageToken)

				for _, name := range tss {
					t.Run("endpoint="+name, func(t *testing.T) {
						parsed := getList(t, name, qs)
						assert.Len(t, parsed.Array(), procCount)

						for _, item := range parsed.Array() {
							assert.Equal(t, "processing", item.Get("status").String())
						}
					})
				}
			})
			t.Run("case=should return all messages with recipient equals to noreply@ory.sh", func(t *testing.T) {
				qs := fmt.Sprintf(`?page_token=%s&page_size=250&recipient=noreply@ory.sh`, defaultPageToken)

				for _, name := range tss {
					t.Run("endpoint="+name, func(t *testing.T) {
						parsed := getList(t, name, qs)
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
			for _, name := range tss {
				t.Run("endpoint="+name, func(t *testing.T) {
					parsed := getList(t, name, "")
					require.Len(t, parsed.Array(), msgCount, "%s", parsed.Raw)

					for _, item := range parsed.Array() {
						assert.Equal(t, "<redacted-unless-dev-mode>", item.Get("body").String())
					}
				})
			}
		})
		t.Run("case=body should not be redacted if kratos is in dev mode", func(t *testing.T) {
			conf.MustSet(ctx, "dev", true)
			for _, name := range tss {
				t.Run("endpoint="+name, func(t *testing.T) {
					parsed := getList(t, name, "")
					require.Len(t, parsed.Array(), msgCount, "%s", parsed.Raw)

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
}

func createMessage(id uuid.UUID, i int) courier.Message {

	status := func(i int) courier.MessageStatus {
		if i%2 == 0 {
			return courier.MessageStatusAbandoned
		}
		return courier.MessageStatusSent
	}

	templateType := func(i int) courier.TemplateType {
		if i%2 == 0 {
			return courier.TypeRecoveryCodeInvalid
		}
		return courier.TypeRecoveryCodeValid
	}

	return courier.Message{
		ID:           id,
		Status:       status(i),
		Type:         courier.MessageTypeEmail,
		Recipient:    fmt.Sprintf("test%d@test.com", i),
		Body:         fmt.Sprintf("test body %d", i),
		Subject:      fmt.Sprintf("test subject %d", i),
		TemplateType: templateType(i),
		SendCount:    9,
	}
}

func getNextToken(links []string) string {

	nextLink := links[1]

	re := regexp.MustCompile("<.*page_token=(?P<uuid>.*)>.*")

	g := re.FindStringSubmatch(nextLink)

	return g[1]
}

func TestPaginationWithOrder(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	// Start kratos server
	_, adminTS := testhelpers.NewKratosServerWithCSRF(t, reg)

	conf.MustSet(ctx, config.ViperKeyAdminBaseURL, adminTS.URL)

	var getList = func(t *testing.T, pageToken string, pageSize int) (gjson.Result, []string) {
		t.Helper()
		v := url.Values{}
		if pageToken != "" {
			v.Add("page_token", pageToken)
		}
		if pageSize > 0 {
			v.Add("page_size", fmt.Sprintf("%d", pageSize))
		}

		href := adminTS.URL + courier.AdminRouteListMessages + "?" + v.Encode()

		resp, err := adminTS.Client().Get(href)
		require.NoError(t, err)
		body := ioutilx.MustReadAll(resp.Body)
		links := resp.Header.Values("link")
		return gjson.ParseBytes(body), links
	}

	expectedUUIDs := ""

	for i := 11; i <= 20; i++ {

		uuid := uuidx.NewV4()

		expectedUUIDs = expectedUUIDs + uuid.String()

		msg := createMessage(uuid, i)
		msg.NID = reg.Persister().NetworkID(ctx)
		// time.Sleep(1 * time.Second)
		require.NoError(t, reg.Persister().GetConnection(ctx).Create(&msg))
	}

	r, links := getList(t, "", 5)
	require.Len(t, r.Array(), 5)
	require.Len(t, links, 2)

	actualUUIDs := ""

	for _, m := range r.Array() {
		actualUUIDs = m.Get("id").String() + actualUUIDs
	}

	nextToken := getNextToken(links)

	r, _ = getList(t, nextToken, 5)

	for _, m := range r.Array() {
		actualUUIDs = m.Get("id").String() + actualUUIDs
	}

	assert.Equal(t, expectedUUIDs, actualUUIDs)
}
