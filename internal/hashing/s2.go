package hashing

import (
	"fmt"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/grntlrduck-cloud/dynageo/geo"
)

// in the case of a radius search we want to return more results than in the radius intentiaionally
// so that if a user zooms there are still enough PoI centered
// http://s2geometry.io/resources/s2cell_statistics.html
// BBox and Radius search default
var defaultAreaCoverer = s2.RegionCoverer{
	MinLevel: 9, // more coarse
	MaxLevel: 13,
	MaxCells: 15,
	LevelMod: 1,
}

// default for Route search
var defaultPolylineCoverer = s2.RegionCoverer{
	MinLevel: 9,
	MaxLevel: 15,  // fine grainer
	MaxCells: 100, // needs to cover longer area
	LevelMod: 1,
}

// The GeoHash wraps and hides the actual geohashing complexity
type S2GeoHash struct {
	level  int
	hashID s2.CellID
}

func (s *S2GeoHash) Hash() uint64 {
	return uint64(s.hashID)
}

func (s *S2GeoHash) Trimmed() uint64 {
	if s.level < 0 || s.level > 30 {
		return uint64(s.hashID)
	}
	parent := s2.CellIDFromFacePosLevel(s.hashID.Face(), s.hashID.Pos(), s.level)
	return uint64(parent)
}

func (s *S2GeoHash) Min() uint64 {
	return uint64(s.hashID.RangeMin())
}

func (s *S2GeoHash) Max() uint64 {
	return uint64(s.hashID.RangeMax())
}

func (s *S2GeoHash) Level() int {
	return s.level
}

func NewS2GeoHash(coordinates geo.Coordinates, level int) (*S2GeoHash, error) {
	if coordinates.Latitude < minLatitude || coordinates.Latitude > maxLatitude ||
		coordinates.Longitude < minLongitude ||
		coordinates.Longitude > maxLongitude {
		return nil, fmt.Errorf(
			"invalid coordinates: latitude=%f, longitude=%f",
			coordinates.Latitude,
			coordinates.Longitude,
		)
	}

	if level < 0 || level > 30 {
		return nil, fmt.Errorf("invalid level: %d, level must be between 0 and 30", level)
	}

	latLonAngles := s2.LatLngFromDegrees(coordinates.Latitude, coordinates.Longitude)
	cell := s2.CellFromLatLng(latLonAngles)
	return &S2GeoHash{hashID: cell.ID(), level: level}, nil
}

func NewHashesFromRadiusCenter(
	c geo.Coordinates,
	radius float64,
	level int,
	coverer *s2.RegionCoverer,
) ([]S2GeoHash, error) {
	if !isValidLatLon(c.Latitude, c.Longitude) {
		return nil, fmt.Errorf("invalid search center: lat=%f, lon=%f", c.Latitude, c.Longitude)
	}
	angle := s1.Angle(radius / earthRadiusMeter)
	centerPoint := pointFromCords(c)
	region := s2.CapFromCenterAngle(centerPoint, angle)
	if coverer == nil {
		coverer = &defaultAreaCoverer
	}
	covering := coverer.Covering(region)
	return newGeoHashesFromCells(covering, level), nil
}

func NewHashesFromBbox(
	ne, sw geo.Coordinates,
	level int,
	coverer *s2.RegionCoverer,
) ([]S2GeoHash, error) {
	if !isValidLatLon(ne.Latitude, ne.Longitude) || !isValidLatLon(sw.Latitude, sw.Longitude) {
		return nil, fmt.Errorf(
			"invalid bounding box: ne.lat=%f, ne.lon=%f, sw.lat=%f, sw.lon=%f",
			ne.Latitude,
			ne.Longitude,
			sw.Latitude,
			sw.Longitude,
		)
	}
	bounder := s2.NewRectBounder()
	bounder.AddPoint(pointFromCords(ne))
	bounder.AddPoint(pointFromCords(sw))
	if coverer == nil {
		coverer = &defaultAreaCoverer
	}
	covering := coverer.Covering(bounder.RectBound())
	return newGeoHashesFromCells(covering, level), nil
}

func NewHashesFromRoute(
	path []geo.Coordinates,
	level int,
	coverer *s2.RegionCoverer,
) ([]S2GeoHash, error) {
	if len(path) < 2 {
		return nil, fmt.Errorf("invalid path: length=%d", len(path))
	}

	// Validate coordinates before processing
	for _, p := range path {
		if !isValidLatLon(p.Latitude, p.Longitude) {
			return nil, fmt.Errorf(
				"invalid coordinates for route: lat=%f, lon:=%f",
				p.Latitude,
				p.Longitude,
			)
		}
	}

	// Pre-allocate slice with exact capacity
	latLngs := make([]s2.LatLng, len(path))
	for i, p := range path {
		latLngs[i] = s2.LatLngFromDegrees(p.Latitude, p.Longitude)
	}

	polyline := s2.PolylineFromLatLngs(latLngs)
	if coverer == nil {
		coverer = &defaultPolylineCoverer
	}
	covering := coverer.Covering(polyline)
	return newGeoHashesFromCells(covering, level), nil
}

func newGeoHashesFromCells(cells []s2.CellID, level int) []S2GeoHash {
	hashes := make([]S2GeoHash, len(cells))
	for i, v := range cells {
		hashes[i] = S2GeoHash{hashID: v, level: level}
	}
	return hashes
}

func pointFromCords(c geo.Coordinates) s2.Point {
	return s2.PointFromLatLng(s2.LatLngFromDegrees(c.Latitude, c.Longitude))
}

func isValidLatLon(lat, lon float64) bool {
	return lat >= minLatitude && lat <= maxLatitude &&
		lon >= minLongitude && lon <= maxLongitude
}
