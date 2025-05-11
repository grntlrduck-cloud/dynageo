package dynamodb

import (
	"fmt"

	"github.com/grntlrduck-cloud/dynageo/geo"
	"github.com/grntlrduck-cloud/dynageo/internal/hashing"
)

type S2GeoAttributes struct {
	Latitude  float64 `dynamodbav:"latitude"`
	Longitude float64 `dynamodbav:"longitude"`
	GeoAttributes[uint64]
}

func NewS2GeoAttributes(cfg geoIndexConfig, coordinates geo.Coordinates) (*S2GeoAttributes, error) {
	hash, err := hashing.NewS2GeoHash(coordinates, cfg.level)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to construct S2 geohash for index %s, and coordinates latitude=%f, longitude=%f: %w",
			cfg.geoIndexName,
			coordinates.Latitude,
			coordinates.Longitude,
			err,
		)
	}

	return &S2GeoAttributes{
		Longitude: coordinates.Longitude,
		Latitude:  coordinates.Latitude,
		GeoAttributes: newGeoAttributes(
			hash.Hash(),
			hash.Trimmed(),
			cfg,
		),
	}, nil
}

// S2MultiGeoAttributes represents a location with multiple geospatial indices.
// It implements the MultiIndexGeoItem interface by embedding MultiGeoAttributes.
type S2MultiGeoAttributes struct {
	Longitude float64 `dynamodbav:"longitude"`
	Latitude  float64 `dynamodbav:"latitude"`
	MultiGeoAttributes[uint64]
}

// NewS2MultiGeoAttributes creates a new S2MultiGeoAttributes instance with the given configurations and coordinates.
// This allows for storing multiple geo indices for a single item.
func NewS2MultiGeoAttributes(
	cfgs []geoIndexConfig,
	coordinates geo.Coordinates,
) (*S2MultiGeoAttributes, error) {
	if len(cfgs) > 10 {
		return nil, fmt.Errorf(
			"maximum number of geo indices exceeded, maximum is 10, but got %d",
			len(cfgs),
		)
	}

	geoAtributes := make(map[string]GeoAttributes[uint64])

	for i := range cfgs {
		hash, err := hashing.NewS2GeoHash(coordinates, cfgs[i].Level())
		if err != nil {
			return nil, fmt.Errorf(
				"failed to construct S2 geohash for index %s, and coordinates latitude=%f, longitude=%f: %w",
				cfgs[i].GeoIndexName(),
				coordinates.Latitude,
				coordinates.Longitude,
				err,
			)
		}

		geoAtributes[cfgs[i].GeoIndexName()] = newGeoAttributes(
			hash.Hash(),
			hash.Trimmed(),
			cfgs[i],
		)
	}

	return &S2MultiGeoAttributes{
		Longitude: coordinates.Longitude,
		Latitude:  coordinates.Latitude,
		MultiGeoAttributes: MultiGeoAttributes[uint64]{
			GeoAttributes: geoAtributes,
		},
	}, nil
}
