package dynamodb

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

type GeoItem[T any] interface {
	GeoHash() T
	TrimmedGeoHash() T
	GeoIndexName() string
	GeoIndexLevel() int
	MarshalDynamoDBAttributeValue() (types.AttributeValue, error)
	UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error
}

type MultiIndexGeoItem[T any] interface {
	GeoHash(index string) any
	TrimmedGeoHash(index string) any
	GeoIndices() []string
	GeoIndexLevel(index string) int
	MarshalDynamoDBAttributeValue() (types.AttributeValue, error)
	UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error
}
