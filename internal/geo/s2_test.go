package geo

import (
	"testing"

	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGeoHash(t *testing.T) {
	tests := []struct {
		name      string
		lat       float64
		lon       float64
		wantError bool
	}{
		{
			name:      "Valid Coordinates",
			lat:       37.7749,
			lon:       -122.4194,
			wantError: false,
		},
		{
			name:      "Invalid Latitude (too high)",
			lat:       91.0,
			lon:       0.0,
			wantError: true,
		},
		{
			name:      "Invalid Latitude (too low)",
			lat:       -91.0,
			lon:       0.0,
			wantError: true,
		},
		{
			name:      "Invalid Longitude (too high)",
			lat:       0.0,
			lon:       181.0,
			wantError: true,
		},
		{
			name:      "Invalid Longitude (too low)",
			lat:       0.0,
			lon:       -181.0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := newGeoHash(tt.lat, tt.lon)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hash)
				assert.NotZero(t, hash.hash())
			}
		})
	}
}

func TestGeoHashTrimmed(t *testing.T) {
	hash, err := newGeoHash(37.7749, -122.4194)
	require.NoError(t, err)
	require.NotNil(t, hash)

	tests := []struct {
		name  string
		level int
	}{
		{"Level 0", 0},
		{"Level 10", 10},
		{"Level 15", 15},
		{"Level 30", 30},
		{"Invalid Level (negative)", -1},
		{"Invalid Level (too high)", 31},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trimmed := hash.trimmed(tt.level)
			assert.NotZero(t, trimmed)

			if tt.level >= 0 && tt.level <= 30 {
				// For valid levels, verify it creates a cell at the correct level
				cell := s2.CellID(trimmed)
				assert.Equal(t, tt.level, cell.Level())
			} else {
				// For invalid levels, it should return the original hash
				assert.Equal(t, hash.hash(), trimmed)
			}
		})
	}
}

func TestGeoHashRanges(t *testing.T) {
	hash, err := newGeoHash(37.7749, -122.4194)
	require.NoError(t, err)
	require.NotNil(t, hash)

	min := hash.min()
	max := hash.max()

	assert.NotZero(t, min)
	assert.NotZero(t, max)

	// For leaf cells, min and max might be equal
	// So we just verify they're valid
	if hash.hashID.IsLeaf() {
		assert.LessOrEqual(t, min, max)
	} else {
		assert.Less(t, min, max, "min should be less than max")
	}
}

func TestHashesFromRadiusCenter(t *testing.T) {
	tests := []struct {
		name      string
		center    Coordinates
		radius    float64
		coverer   *s2.RegionCoverer
		wantError bool
	}{
		{
			name:      "Valid center and radius",
			center:    Coordinates{Latitude: 37.7749, Longitude: -122.4194},
			radius:    1000, // 1km
			coverer:   nil,  // Use default
			wantError: false,
		},
		{
			name:      "Invalid center",
			center:    Coordinates{Latitude: 91.0, Longitude: 0.0},
			radius:    1000,
			coverer:   nil,
			wantError: true,
		},
		{
			name:      "Custom coverer",
			center:    Coordinates{Latitude: 37.7749, Longitude: -122.4194},
			radius:    5000, // 5km
			coverer:   &s2.RegionCoverer{MinLevel: 8, MaxLevel: 12, MaxCells: 10},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashes, err := newHashesFromRadiusCenter(tt.center, tt.radius, tt.coverer)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, hashes)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hashes)
				assert.NotEmpty(t, hashes)
			}
		})
	}
}

func TestHashesFromBbox(t *testing.T) {
	tests := []struct {
		name      string
		ne        Coordinates
		sw        Coordinates
		coverer   *s2.RegionCoverer
		wantError bool
	}{
		{
			name:      "Valid bounding box",
			ne:        Coordinates{Latitude: 38.0, Longitude: -122.0},
			sw:        Coordinates{Latitude: 37.0, Longitude: -123.0},
			coverer:   nil, // Use default
			wantError: false,
		},
		{
			name:      "Invalid NE coordinate",
			ne:        Coordinates{Latitude: 91.0, Longitude: 0.0},
			sw:        Coordinates{Latitude: 37.0, Longitude: -123.0},
			coverer:   nil,
			wantError: true,
		},
		{
			name:      "Invalid SW coordinate",
			ne:        Coordinates{Latitude: 38.0, Longitude: -122.0},
			sw:        Coordinates{Latitude: 37.0, Longitude: -181.0},
			coverer:   nil,
			wantError: true,
		},
		{
			name:      "Custom coverer",
			ne:        Coordinates{Latitude: 38.0, Longitude: -122.0},
			sw:        Coordinates{Latitude: 37.0, Longitude: -123.0},
			coverer:   &s2.RegionCoverer{MinLevel: 8, MaxLevel: 12, MaxCells: 10},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashes, err := newHashesFromBbox(tt.ne, tt.sw, tt.coverer)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, hashes)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hashes)
				assert.NotEmpty(t, hashes)
			}
		})
	}
}

func TestHashesFromRoute(t *testing.T) {
	tests := []struct {
		name      string
		path      []Coordinates
		coverer   *s2.RegionCoverer
		wantError bool
	}{
		{
			name: "Valid path",
			path: []Coordinates{
				{Latitude: 37.7749, Longitude: -122.4194},
				{Latitude: 37.7755, Longitude: -122.4200},
				{Latitude: 37.7760, Longitude: -122.4210},
			},
			coverer:   nil, // Use default
			wantError: false,
		},
		{
			name:      "Path too short",
			path:      []Coordinates{{Latitude: 37.7749, Longitude: -122.4194}},
			coverer:   nil,
			wantError: true,
		},
		{
			name: "Invalid coordinates in path",
			path: []Coordinates{
				{Latitude: 37.7749, Longitude: -122.4194},
				{Latitude: 91.0, Longitude: 0.0},
			},
			coverer:   nil,
			wantError: true,
		},
		{
			name: "Custom coverer",
			path: []Coordinates{
				{Latitude: 37.7749, Longitude: -122.4194},
				{Latitude: 37.7755, Longitude: -122.4200},
				{Latitude: 37.7760, Longitude: -122.4210},
			},
			coverer:   &s2.RegionCoverer{MinLevel: 10, MaxLevel: 16, MaxCells: 150},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashes, err := newHashesFromRoute(tt.path, tt.coverer)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, hashes)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hashes)
				assert.NotEmpty(t, hashes)
			}
		})
	}
}

func TestIsValidLatLon(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		lon      float64
		expected bool
	}{
		{"Valid coordinates", 37.7749, -122.4194, true},
		{"Max valid values", maxLatitude, maxLongitude, true},
		{"Min valid values", minLatitude, minLongitude, true},
		{"Invalid latitude (too high)", 90.1, 0, false},
		{"Invalid latitude (too low)", -90.1, 0, false},
		{"Invalid longitude (too high)", 0, 180.1, false},
		{"Invalid longitude (too low)", 0, -180.1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidLatLon(tt.lat, tt.lon)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPointFromCords(t *testing.T) {
	coords := Coordinates{Latitude: 37.7749, Longitude: -122.4194}
	point := pointFromCords(coords)

	// Verify the point was created correctly by converting back to lat/lng
	latLng := s2.LatLngFromPoint(point)
	assert.InDelta(t, coords.Latitude, latLng.Lat.Degrees(), 0.0001)
	assert.InDelta(t, coords.Longitude, latLng.Lng.Degrees(), 0.0001)
}

func TestNewGeoHashesFromCells(t *testing.T) {
	cells := []s2.CellID{
		s2.CellIDFromFacePosLevel(0, 0, 10),
		s2.CellIDFromFacePosLevel(1, 0, 10),
		s2.CellIDFromFacePosLevel(2, 0, 10),
	}

	hashes := newGeoHashesFromCells(cells)
	assert.Equal(t, len(cells), len(hashes))

	for i, cell := range cells {
		assert.Equal(t, uint64(cell), hashes[i].hash())
	}
}
