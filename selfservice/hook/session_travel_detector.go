package hook

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/node"
)

const earthRadiusKm = 6371.0

var (
	_ login.PostHookExecutor    = new(SessionImpossibleTravelDetector)
	_ recovery.PostHookExecutor = new(SessionImpossibleTravelDetector)
)

type SessionImpossibleTravelDetector struct {
}

func NewSessionImpossibleTravelDetector() *SessionImpossibleTravelDetector {
	return &SessionImpossibleTravelDetector{}
}

func (e *SessionImpossibleTravelDetector) ExecuteLoginPostHook(_ http.ResponseWriter, r *http.Request, _ node.UiNodeGroup, _ *login.Flow, s *session.Session) error {
	return e.checkLocationAndFlagSession(r, s)
}

func (e *SessionImpossibleTravelDetector) ExecutePostRecoveryHook(_ http.ResponseWriter, r *http.Request, _ *recovery.Flow, s *session.Session) error {
	return e.checkLocationAndFlagSession(r, s)
}

func (e *SessionImpossibleTravelDetector) checkLocationAndFlagSession(r *http.Request, s *session.Session) error {
	// since we assume Kratos is behind Cloudflare we can extract geolocation data
	// from CF-IPLatitude, CF-IPLongitude headers
	latStr := r.Header.Get("CF-IPLatitude")
	lonStr := r.Header.Get("CF-IPLongitude")

	// just for the sake of making everything safe
	if latStr == "" || lonStr == "" {
		return nil
	}

	// whenever those headers are invalid we don't have enough info to flag sessions
	// so we silently skip
	currentLat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return nil
	}
	currentLon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return nil
	}
	now := time.Now().UTC()
	if s.LastLocationLat != nil && s.LastLocationLon != nil && s.LastLocationAt != nil {
		distanceKm := e.calculateHaversineDistance(*s.LastLocationLat, *s.LastLocationLon, currentLat, currentLon)
		timeDiffMinutes := now.Sub(*s.LastLocationAt).Minutes()

		if timeDiffMinutes > 0 {
			// ignore small distances cause it can happen when connecting to different networks
			if distanceKm > 50.0 {
				speedKmPerMinute := distanceKm / timeDiffMinutes

				// 900 km/h is exactly 15 km/minute.
				// we flag everything that's faster than a commercial plane
				if speedKmPerMinute > 15.0 {
					s.ImpossibleTravel = true
				}
			}
		}
	}
	// update session with the latest values, executed every time
	// if session is new we just update the info so that consequential logins
	// can be checked
	s.LastLocationLon = &currentLon
	s.LastLocationLat = &currentLat
	s.LastLocationAt = &now
	return nil
}

// calculates distance between two coordinates in km
func (e *SessionImpossibleTravelDetector) calculateHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	lat1Rad := lat1 * (math.Pi / 180.0)
	lat2Rad := lat2 * (math.Pi / 180.0)

	// Haversine formula
	a := math.Pow(math.Sin(dLat/2), 2) + math.Pow(math.Sin(dLon/2), 2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}
