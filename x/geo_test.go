// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCoordinates(t *testing.T) {
	testData := []struct {
		title             string
		strLat            string
		strLng            string
		expectedGeoCoords GeoCoordinates
		expectedOk        bool
	}{
		{
			title:             "Empty Latitude and empty Longitude",
			strLat:            "",
			strLng:            "",
			expectedGeoCoords: GeoCoordinates{0, 0},
			expectedOk:        false,
		},
		{
			title:             "Empty Latitude and non empty Longitude",
			strLat:            "",
			strLng:            "12.0100",
			expectedGeoCoords: GeoCoordinates{0, 0},
			expectedOk:        false,
		},
		{
			title:             "Bad Latitude and Bad Longitude",
			strLat:            "non_float",
			strLng:            "also_non_float",
			expectedGeoCoords: GeoCoordinates{0, 0},
			expectedOk:        false,
		},
		{
			title:             "Bad Latitude and ok Longitude",
			strLat:            "non_float",
			strLng:            "12.123456",
			expectedGeoCoords: GeoCoordinates{0, 0},
			expectedOk:        false,
		},
		{
			title:             "Ok Latitude and bad Longitude",
			strLat:            "-15.33",
			strLng:            "non_float",
			expectedGeoCoords: GeoCoordinates{0, 0},
			expectedOk:        false,
		},
		{
			title:             "Positive both",
			strLat:            "15.0214533",
			strLng:            "19.1021021",
			expectedGeoCoords: GeoCoordinates{15.0214533, 19.1021021},
			expectedOk:        true,
		},
		{
			title:             "Negative both",
			strLat:            "-13.0214533",
			strLng:            "-11.0214533",
			expectedGeoCoords: GeoCoordinates{-13.0214533, -11.0214533},
			expectedOk:        true,
		},
		{
			title:             "Positive and Negative",
			strLat:            "-13.0214533",
			strLng:            "11.0214533",
			expectedGeoCoords: GeoCoordinates{-13.0214533, 11.0214533},
			expectedOk:        true,
		},
	}
	for _, testCase := range testData {
		t.Run(testCase.title, func(t *testing.T) {
			actualGeoCoordinates, actualOk := NewGeoCoordinates(testCase.strLat, testCase.strLng)
			assert.Equal(t, testCase.expectedGeoCoords, actualGeoCoordinates)
			assert.Equal(t, testCase.expectedOk, actualOk)
		})
	}
}

func TestCalcHaversine(t *testing.T) {

	const tolerance = 0.001
	const earthArc = math.Pi * earthRadiusKm
	origin := GeoCoordinates{0, 0}

	london := GeoCoordinates{51.5074, -0.1278}
	londonAntipode := GeoCoordinates{-51.5074, 179.8722}

	newYork := GeoCoordinates{40.7128, -74.0060}

	northPole := GeoCoordinates{90.0, 0.0}
	southPole := GeoCoordinates{-90.0, 0.0}

	equatorP1 := GeoCoordinates{0.0, 0.0}
	equatorP2 := GeoCoordinates{0.0, 90.0}

	testData := []struct {
		title    string
		g1       GeoCoordinates
		g2       GeoCoordinates
		expected float64
	}{
		{

			title:    "Zero distance (origin)",
			g1:       origin,
			g2:       origin,
			expected: 0.0,
		},
		{
			title:    "Zero distance (non-origin)",
			g1:       london,
			g2:       london,
			expected: 0.0,
		},
		{
			title:    "London to New York",
			g1:       london,
			g2:       newYork,
			expected: 5570.222180, // manually calculated
		},
		{
			title:    "North Pole to South Pole",
			g1:       northPole,
			g2:       southPole,
			expected: earthArc,
		},
		{
			title:    "Antipodal points (London vs. its antipode)",
			g1:       london,
			g2:       londonAntipode,
			expected: earthArc,
		},
		{
			title:    "Points on the equator (1/4 circumference)",
			g1:       equatorP1,
			g2:       equatorP2,
			expected: (2 * earthArc) / 4,
		},
		{
			title:    "Points on the same meridian (10 degrees apart)",
			g1:       GeoCoordinates{10.0, 0.0},
			g2:       GeoCoordinates{20.0, 0.0},
			expected: 1111.950, // manually calculated
		},
	}
	for _, testCase := range testData {
		t.Run(testCase.title, func(t *testing.T) {
			actual := calcHaversineKm(testCase.g1, testCase.g2)
			assert.InDelta(t, testCase.expected, actual, tolerance)
		})
	}
}

func TestIsImpossibleTravel(t *testing.T) {
	london := GeoCoordinates{Latitude: 51.5074, Longitude: -0.1278}
	newYork := GeoCoordinates{Latitude: 40.7128, Longitude: -74.0060}

	testData := []struct {
		title             string
		g1                GeoCoordinates
		g2                GeoCoordinates
		timeDeltaHours    float64
		maxTravelSpeedKmH float64
		expected          bool
	}{
		{
			title:             "Same place, same time",
			g1:                london,
			g2:                london,
			timeDeltaHours:    0.0,
			maxTravelSpeedKmH: 1000,
			expected:          false,
		},
		{
			title:             "Different place, same time",
			g1:                london,
			g2:                newYork,
			timeDeltaHours:    0.0,
			maxTravelSpeedKmH: 1000,
			expected:          true,
		},
		{
			title:             "normal flights",
			g1:                london,
			g2:                newYork,
			timeDeltaHours:    10.0,
			maxTravelSpeedKmH: 900,
			expected:          false,
		},
		{
			title:             "Impossible travel for normal flights (2785 km/h)",
			g1:                london,
			g2:                newYork,
			timeDeltaHours:    2.0,
			maxTravelSpeedKmH: 900,

			expected: true,
		},
		{
			title:             "Possible travel for super-sonic flights (2785 km/h)",
			g1:                london,
			g2:                newYork,
			timeDeltaHours:    2.0,
			maxTravelSpeedKmH: 3000,
			expected:          false,
		},
	}

	for _, testCase := range testData {
		t.Run(testCase.title, func(t *testing.T) {
			actual := IsImpossibleTravel(
				testCase.g1,
				testCase.g2,
				testCase.timeDeltaHours,
				testCase.maxTravelSpeedKmH,
			)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
