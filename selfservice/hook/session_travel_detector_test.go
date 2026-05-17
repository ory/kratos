package hook

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionImpossibleTravelDetector(t *testing.T) {
	detector := NewSessionImpossibleTravelDetector()

	// Helper function to create float pointers
	floatPtr := func(f float64) *float64 { return &f }
	timePtr := func(t time.Time) *time.Time { return &t }

	now := time.Now().UTC()

	// Coordinates for testing
	nyLat, nyLon := 40.7128, -74.0060

	tests := []struct {
		name                 string
		reqHeaders           map[string]string
		existingSession      *session.Session
		expectedFlag         bool
		expectLocationUpdate bool
	}{
		{
			name: "missing headers does nothing",
			reqHeaders: map[string]string{
				"CF-IPLatitude": "", // Missing longitude
			},
			existingSession:      &session.Session{},
			expectedFlag:         false,
			expectLocationUpdate: false,
		},
		{
			name: "invalid headers does nothing",
			reqHeaders: map[string]string{
				"CF-IPLatitude":  "invalid",
				"CF-IPLongitude": "-74.0060",
			},
			existingSession:      &session.Session{},
			expectedFlag:         false,
			expectLocationUpdate: false,
		},
		{
			name: "first login stamps session but does not flag",
			reqHeaders: map[string]string{
				"CF-IPLatitude":  "40.7128",
				"CF-IPLongitude": "-74.0060",
			},
			existingSession:      &session.Session{}, // Empty session
			expectedFlag:         false,
			expectLocationUpdate: true,
		},
		{
			name: "subsequent login with normal travel (10 hours to London) does not flag",
			reqHeaders: map[string]string{
				"CF-IPLatitude":  "51.5074", // London
				"CF-IPLongitude": "-0.1278",
			},
			existingSession: &session.Session{
				LastLocationLat: floatPtr(nyLat),
				LastLocationLon: floatPtr(nyLon),
				LastLocationAt:  timePtr(now.Add(-10 * time.Hour)), // 10 hours ago
			},
			expectedFlag:         false,
			expectLocationUpdate: true,
		},
		{
			name: "subsequent login with impossible travel (1 hour to London) flags session",
			reqHeaders: map[string]string{
				"CF-IPLatitude":  "51.5074", // London
				"CF-IPLongitude": "-0.1278",
			},
			existingSession: &session.Session{
				LastLocationLat: floatPtr(nyLat),
				LastLocationLon: floatPtr(nyLon),
				LastLocationAt:  timePtr(now.Add(-1 * time.Hour)), // Only 1 hour ago!
			},
			expectedFlag:         true,
			expectLocationUpdate: true,
		},
		{
			name: "network jitter (instant travel but short distance) does not flag",
			reqHeaders: map[string]string{
				"CF-IPLatitude":  "40.7300", // NJ
				"CF-IPLongitude": "-74.0500",
			},
			existingSession: &session.Session{
				LastLocationLat: floatPtr(nyLat),
				LastLocationLon: floatPtr(nyLon),
				LastLocationAt:  timePtr(now.Add(-1 * time.Minute)), // 1 minute ago, technically > 900kmh but < 50km total distance
			},
			expectedFlag:         false,
			expectLocationUpdate: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/login", nil)
			require.NoError(t, err)
			for k, v := range tc.reqHeaders {
				req.Header.Set(k, v)
			}

			err = detector.ExecuteLoginPostHook(nil, req, node.DefaultGroup, nil, tc.existingSession)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedFlag, tc.existingSession.ImpossibleTravel, "ImpossibleTravel flag mismatch")

			if tc.expectLocationUpdate {
				assert.NotNil(t, tc.existingSession.LastLocationLat)
				assert.NotNil(t, tc.existingSession.LastLocationLon)
				assert.NotNil(t, tc.existingSession.LastLocationAt)

				// Verify the stamped coordinates match the headers we passed in
				expectedLat, _ := strconv.ParseFloat(tc.reqHeaders["CF-IPLatitude"], 64)
				assert.Equal(t, expectedLat, *tc.existingSession.LastLocationLat)
			} else {
				assert.Nil(t, tc.existingSession.LastLocationLat)
			}
		})
	}
}
