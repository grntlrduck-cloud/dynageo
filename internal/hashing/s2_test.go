package hashing

import (
	"testing"

	"github.com/golang/geo/s2"
	"github.com/grntlrduck-cloud/dynageo/geo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewS2GeoHash(t *testing.T) {
	tests := []struct {
		name      string
		coords    geo.Coordinates
		level     int
		wantError bool
	}{
		{
			name:      "Valid Coordinates",
			coords:    geo.Coordinates{Latitude: 37.7749, Longitude: -122.4194},
			level:     15,
			wantError: false,
		},
		{
			name:      "Invalid Latitude (too high)",
			coords:    geo.Coordinates{Latitude: 91.0, Longitude: 0.0},
			level:     15,
			wantError: true,
		},
		{
			name:      "Invalid Latitude (too low)",
			coords:    geo.Coordinates{Latitude: -91.0, Longitude: 0.0},
			level:     15,
			wantError: true,
		},
		{
			name:      "Invalid Longitude (too high)",
			coords:    geo.Coordinates{Latitude: 0.0, Longitude: 181.0},
			level:     15,
			wantError: true,
		},
		{
			name:      "Invalid Longitude (too low)",
			coords:    geo.Coordinates{Latitude: 0.0, Longitude: -181.0},
			level:     15,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := NewS2GeoHash(tt.coords, tt.level)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hash)
				assert.NotZero(t, hash.Hash())
			}
		})
	}
}

func TestGeoHashTrimmed(t *testing.T) {
	coords := geo.Coordinates{Latitude: 37.7749, Longitude: -122.4194}

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
			hash, err := NewS2GeoHash(coords, tt.level)

			if tt.level >= 0 && tt.level <= 30 {
				trimmed := hash.Trimmed()
				assert.NotZero(t, trimmed)

				assert.Equal(t, tt.level, hash.Level())
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestGeoHashRanges(t *testing.T) {
	coords := geo.Coordinates{Latitude: 37.7749, Longitude: -122.4194}
	hash, err := NewS2GeoHash(coords, 15)
	require.NoError(t, err)
	require.NotNil(t, hash)

	minH := hash.Min()
	maxH := hash.Max()

	assert.NotZero(t, minH)
	assert.NotZero(t, maxH)

	// For leaf cells, min and max might be equal
	// So we just verify they're valid
	cellID := s2.CellID(hash.Hash())
	if cellID.IsLeaf() {
		assert.LessOrEqual(t, minH, maxH)
	} else {
		assert.Less(t, minH, maxH, "min should be less than max")
	}
}

func TestHashesFromRadiusCenter(t *testing.T) {
	tests := []struct {
		name      string
		center    geo.Coordinates
		radius    float64
		level     int
		coverer   *s2.RegionCoverer
		wantError bool
	}{
		{
			name:      "Valid center and radius",
			center:    geo.Coordinates{Latitude: 37.7749, Longitude: -122.4194},
			radius:    1000, // 1km
			level:     15,
			coverer:   nil, // Use default
			wantError: false,
		},
		{
			name:      "Invalid center",
			center:    geo.Coordinates{Latitude: 91.0, Longitude: 0.0},
			radius:    1000,
			level:     15,
			coverer:   nil,
			wantError: true,
		},
		{
			name:      "Custom coverer",
			center:    geo.Coordinates{Latitude: 37.7749, Longitude: -122.4194},
			radius:    5000, // 5km
			level:     12,
			coverer:   &s2.RegionCoverer{MinLevel: 8, MaxLevel: 12, MaxCells: 10},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashes, err := NewHashesFromRadiusCenter(tt.center, tt.radius, tt.level, tt.coverer)
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
		ne        geo.Coordinates
		sw        geo.Coordinates
		level     int
		coverer   *s2.RegionCoverer
		wantError bool
	}{
		{
			name:      "Valid bounding box",
			ne:        geo.Coordinates{Latitude: 38.0, Longitude: -122.0},
			sw:        geo.Coordinates{Latitude: 37.0, Longitude: -123.0},
			level:     15,
			coverer:   nil, // Use default
			wantError: false,
		},
		{
			name:      "Invalid NE coordinate",
			ne:        geo.Coordinates{Latitude: 91.0, Longitude: 0.0},
			sw:        geo.Coordinates{Latitude: 37.0, Longitude: -123.0},
			level:     15,
			coverer:   nil,
			wantError: true,
		},
		{
			name:      "Invalid SW coordinate",
			ne:        geo.Coordinates{Latitude: 38.0, Longitude: -122.0},
			sw:        geo.Coordinates{Latitude: 37.0, Longitude: -181.0},
			level:     15,
			coverer:   nil,
			wantError: true,
		},
		{
			name:      "Custom coverer",
			ne:        geo.Coordinates{Latitude: 38.0, Longitude: -122.0},
			sw:        geo.Coordinates{Latitude: 37.0, Longitude: -123.0},
			level:     12,
			coverer:   &s2.RegionCoverer{MinLevel: 8, MaxLevel: 12, MaxCells: 10},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashes, err := NewHashesFromBbox(tt.ne, tt.sw, tt.level, tt.coverer)
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
		path      []geo.Coordinates
		level     int
		coverer   *s2.RegionCoverer
		wantError bool
	}{
		{
			name: "Valid path",
			path: []geo.Coordinates{
				{Latitude: 37.7749, Longitude: -122.4194},
				{Latitude: 37.7755, Longitude: -122.4200},
				{Latitude: 37.7760, Longitude: -122.4210},
			},
			level:     15,
			coverer:   nil, // Use default
			wantError: false,
		},
		{
			name:      "Path too short",
			path:      []geo.Coordinates{{Latitude: 37.7749, Longitude: -122.4194}},
			level:     15,
			coverer:   nil,
			wantError: true,
		},
		{
			name: "Invalid coordinates in path",
			path: []geo.Coordinates{
				{Latitude: 37.7749, Longitude: -122.4194},
				{Latitude: 91.0, Longitude: 0.0},
			},
			level:     15,
			coverer:   nil,
			wantError: true,
		},
		{
			name: "Custom coverer",
			path: []geo.Coordinates{
				{Latitude: 37.7749, Longitude: -122.4194},
				{Latitude: 37.7755, Longitude: -122.4200},
				{Latitude: 37.7760, Longitude: -122.4210},
			},
			level:     15,
			coverer:   &s2.RegionCoverer{MinLevel: 10, MaxLevel: 16, MaxCells: 150},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashes, err := NewHashesFromRoute(tt.path, tt.level, tt.coverer)
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
	coords := geo.Coordinates{Latitude: 37.7749, Longitude: -122.4194}
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

	level := 10
	hashes := newGeoHashesFromCells(cells, level)
	assert.Equal(t, len(cells), len(hashes))

	for i, cell := range cells {
		assert.Equal(t, uint64(cell), hashes[i].Hash())
		assert.Equal(t, level, hashes[i].level)
	}
}
