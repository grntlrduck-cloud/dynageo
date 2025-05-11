package dynamodb

import (
	"context"

	"github.com/grntlrduck-cloud/dynageo/geo"
)

type GeoRepository[T any] interface {
	GeoGetItem[T]
	GeoPutItem[T]
	GeoBatchPutItem[T]
	GeoGetItemsInRadius[T]
	GeoGetItemsInBBox[T]
	GeoGetItemsOnRoute[T]
}

type GeoGetItem[T any] interface {
	GetItemByGeoHash(ctx context.Context, geoHashKey, geoSortKey uint64) (*GeoItem[T], error)
}

type GeoPutItem[T any] interface {
	PutItem(ctx context.Context, item GeoItem[T]) error
}

type GeoBatchPutItem[T any] interface {
	BatchPutItem(ctx context.Context, items []GeoItem[T]) error
}

type GeoGetItemsInRadius[T any] interface {
	GetItemsInRadius(
		ctx context.Context,
		center geo.Coordinates,
		radius float64,
	) ([]GeoItem[T], error)
}

type GeoGetItemsInBBox[T any] interface {
	GetItemsInBbox(
		ctx context.Context,
		sw, ne geo.Coordinates,
	) ([]GeoItem[T], error)
}

type GeoGetItemsOnRoute[T any] interface {
	GetItemsOnRoute(
		ctx context.Context,
		path []geo.Coordinates,
	) ([]GeoItem[T], error)
}

type MultiIndexGeoRepository[T any] interface {
	MultiIndexGeoGetItem[T]
	MultiIndexGeoPutItem[T]
	MultiIndexGeoBatchPutItem[T]
	MultiIndexGeoGetItemsInRadius[T]
	MultiIndexGeoGetItemsInBBox[T]
	MultiIndexGeoGetItemsOnRoute[T]
}

type MultiIndexGeoGetItem[T any] interface {
	GetItemByGeoHash(
		ctx context.Context,
		geoHashKey, geoSortKey uint64,
	) (*MultiIndexGeoItem[T], error)
}

type MultiIndexGeoPutItem[T any] interface {
	PutItem(ctx context.Context, item MultiIndexGeoItem[T]) error
}

type MultiIndexGeoBatchPutItem[T any] interface {
	BatchPutItem(ctx context.Context, items []MultiIndexGeoItem[T]) error
}

type MultiIndexGeoGetItemsInRadius[T any] interface {
	GetItemsInRadius(
		ctx context.Context,
		center geo.Coordinates,
		radius float64,
	) ([]MultiIndexGeoItem[T], error)
}

type MultiIndexGeoGetItemsInBBox[T any] interface {
	GetItemsInBbox(
		ctx context.Context,
		sw, ne geo.Coordinates,
	) ([]MultiIndexGeoItem[T], error)
}

type MultiIndexGeoGetItemsOnRoute[T any] interface {
	GetItemsOnRoute(
		ctx context.Context,
		path []geo.Coordinates,
	) ([]MultiIndexGeoItem[T], error)
}
