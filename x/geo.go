// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"math"
	"strconv"
)

type GeoCoordinates struct {
	Latitude  float64
	Longitude float64
}

const earthRadiusKm = 6371.0

func NewGeoCoordinates(strLat string, strLong string) (GeoCoordinates, bool) {

	if strLat == "" {
		return GeoCoordinates{0, 0}, false

	}
	if strLong == "" {
		return GeoCoordinates{0, 0}, false

	}
	lat, err := strconv.ParseFloat(strLat, 64)
	if err != nil {
		return GeoCoordinates{0, 0}, false
	}
	long, err := strconv.ParseFloat(strLong, 64)
	if err != nil {
		return GeoCoordinates{0, 0}, false
	}
	return GeoCoordinates{
		Latitude:  lat,
		Longitude: long,
	}, true
}

func deg2rad(d float64) float64 {
	return d * math.Pi / 180.0
}
func calcHaversineKm(a, b GeoCoordinates) float64 {

	dlat := deg2rad(b.Latitude - a.Latitude)
	dlon := deg2rad(b.Longitude - a.Longitude)
	alat := deg2rad(a.Latitude)
	blat := deg2rad(b.Latitude)

	sinDlat := math.Sin(dlat / 2)
	sinDlon := math.Sin(dlon / 2)

	aa := sinDlat*sinDlat + math.Cos(alat)*math.Cos(blat)*sinDlon*sinDlon
	if aa > 1.0 {
		aa = 1.0
	}

	c := 2 * math.Atan2(math.Sqrt(aa), math.Sqrt(1-aa))
	return earthRadiusKm * c
}

// IsImpossibleTravel checks if travel between two locations is impossible
func IsImpossibleTravel(g1, g2 GeoCoordinates, timeDeltaHours float64, maxTravelSpeedKmH float64) bool {
	distance := calcHaversineKm(g1, g2)

	// Anticipate division by 0, and flag impossible due to different locations at the same time
	if timeDeltaHours == 0 {
		return distance > 0
	}

	travelSpeed := distance / timeDeltaHours
	return travelSpeed > maxTravelSpeedKmH
}
