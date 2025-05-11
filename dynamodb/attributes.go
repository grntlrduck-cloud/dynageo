package dynamodb

import (
	"fmt"
	"maps"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type geoIndexConfig struct {
	geoHashKeyAttributeName string
	geoSortKeyAttributeName string
	geoIndexName            string
	level                   int
}

// func newGeoIndexConfig(
// 	geoHashKeyAttributeName, geoSortKeyAttributeName, geoIndexName string,
// 	level int,
// ) geoIndexConfig {
// 	return geoIndexConfig{
// 		geoHashKeyAttributeName: geoHashKeyAttributeName,
// 		geoSortKeyAttributeName: geoSortKeyAttributeName,
// 		geoIndexName:            geoIndexName,
// 		level:                   level,
// 	}
// }

func (g *geoIndexConfig) GeoHashKeyAttrName() string {
	return g.geoHashKeyAttributeName
}

func (g *geoIndexConfig) GeoSortKeyAttrName() string {
	return g.geoSortKeyAttributeName
}

func (g *geoIndexConfig) GeoIndexName() string {
	return g.geoIndexName
}

func (g *geoIndexConfig) Level() int {
	return g.level
}

type GeoAttributes[T any] struct {
	geoHashKey     T `dynamodbav:"-"`
	geoHasSortKey  T `dynamodbav:"-"`
	geoIndexConfig `  dynamodbav:"-"`
}

func newGeoAttributes[T any](
	geoHashKey, geoHasSortKey T,
	geoIndexConfig geoIndexConfig,
) GeoAttributes[T] {
	return GeoAttributes[T]{
		geoHashKey:     geoHashKey,
		geoHasSortKey:  geoHasSortKey,
		geoIndexConfig: geoIndexConfig,
	}
}

func (g *GeoAttributes[T]) GeoHash() T {
	return g.geoHashKey
}

func (g *GeoAttributes[T]) TrimmedGeoHash() T {
	return g.geoHasSortKey
}

func (g *GeoAttributes[T]) GeoIndexName() string {
	return g.geoIndexConfig.GeoIndexName()
}

func (g *GeoAttributes[T]) GeoIndexLevel() int {
	return g.Level()
}

// MarshalDynamoDBAttributeValue implements custom marshaling for GeoAttributes
func (g GeoAttributes[T]) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	attrs := make(map[string]types.AttributeValue)

	// Marshal hash key with configured attribute name
	if g.geoHashKeyAttributeName != "" {
		hashValue, err := marshalGenericValue(g.geoHashKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal geo hash key: %w", err)
		}
		attrs[g.geoHashKeyAttributeName] = hashValue
	}

	// Marshal sort key with configured attribute name
	if g.geoSortKeyAttributeName != "" {
		sortKeyValue, err := marshalGenericValue(g.geoHasSortKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal geo sort key: %w", err)
		}
		attrs[g.geoSortKeyAttributeName] = sortKeyValue
	}

	return &types.AttributeValueMemberM{Value: attrs}, nil
}

// UnmarshalDynamoDBAttributeValue implements custom unmarshaling for GeoAttributes
func (g *GeoAttributes[T]) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	m, ok := av.(*types.AttributeValueMemberM)
	if !ok {
		return fmt.Errorf("expected map attribute value, got %T", av)
	}

	// Unmarshal hash key
	if g.geoHashKeyAttributeName != "" {
		if hashAv, ok := m.Value[g.geoHashKeyAttributeName]; ok {
			hash, err := unmarshalGenericValue[T](hashAv)
			if err == nil {
				g.geoHashKey = hash
			}
		}
	}

	// Unmarshal sort key
	if g.geoSortKeyAttributeName != "" {
		if sortKeyAv, ok := m.Value[g.geoSortKeyAttributeName]; ok {
			sortKey, err := unmarshalGenericValue[T](sortKeyAv)
			if err == nil {
				g.geoHasSortKey = sortKey
			}
		}
	}

	return nil
}

// Helper functions for generic marshaling/unmarshaling
func marshalGenericValue[T any](value T) (types.AttributeValue, error) {
	switch v := any(value).(type) {
	case uint64:
		return &types.AttributeValueMemberN{
			Value: strconv.FormatUint(v, 10),
		}, nil
	case int64:
		return &types.AttributeValueMemberN{
			Value: strconv.FormatInt(v, 10),
		}, nil
	case string:
		return &types.AttributeValueMemberS{
			Value: v,
		}, nil
	case []byte:
		return &types.AttributeValueMemberB{
			Value: v,
		}, nil
	default:
		// Use the standard attributevalue marshaler for other types
		return attributevalue.Marshal(value)
	}
}

func unmarshalGenericValue[T any](av types.AttributeValue) (T, error) {
	var zero T
	var result T

	// Try type-specific unmarshaling first
	switch av := av.(type) {
	case *types.AttributeValueMemberN:
		switch any(result).(type) {
		case uint64:
			v, err := strconv.ParseUint(av.Value, 10, 64)
			if err != nil {
				return zero, err
			}
			return any(v).(T), nil
		case int64:
			v, err := strconv.ParseInt(av.Value, 10, 64)
			if err != nil {
				return zero, err
			}
			return any(v).(T), nil
		}
	case *types.AttributeValueMemberS:
		if _, ok := any(result).(string); ok {
			return any(av.Value).(T), nil
		}
	case *types.AttributeValueMemberB:
		if _, ok := any(result).([]byte); ok {
			return any(av.Value).(T), nil
		}
	}

	// Fall back to standard unmarshaler
	err := attributevalue.Unmarshal(av, &result)
	return result, err
}

// MultiGeoAttributes represents a collection of GeoAttributes indexed by their names.
// This allows for storing and accessing multiple geospatial indices for a single item.
type MultiGeoAttributes[T any] struct {
	GeoAttributes map[string]GeoAttributes[T] `dynamodbav:"-"`
}

// GeoHash returns the geohash value for the specified index
func (m *MultiGeoAttributes[T]) GeoHash(index string) any {
	if geoAttr, ok := m.GeoAttributes[index]; ok {
		return geoAttr.GeoHash()
	}
	return nil
}

// TrimmedGeoHash returns the trimmed geohash value for the specified index
func (m *MultiGeoAttributes[T]) TrimmedGeoHash(index string) any {
	if geoAttr, ok := m.GeoAttributes[index]; ok {
		return geoAttr.TrimmedGeoHash()
	}
	return nil
}

// GeoIndices returns a list of all geo index names stored in this MultiGeoAttributes
func (m *MultiGeoAttributes[T]) GeoIndices() []string {
	indices := make([]string, 0, len(m.GeoAttributes))
	for index := range m.GeoAttributes {
		indices = append(indices, index)
	}
	return indices
}

// GeoIndexLevel returns the level of the geohash for the specified index
func (m *MultiGeoAttributes[T]) GeoIndexLevel(index string) int {
	if geoAttr, ok := m.GeoAttributes[index]; ok {
		return geoAttr.Level()
	}
	return -1 // Return -1 to indicate an invalid index
}

// MarshalDynamoDBAttributeValue implements custom marshaling for MultiGeoAttributes
func (m MultiGeoAttributes[T]) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	// Create a map to store all attributes from all geo indices
	allAttrs := make(map[string]types.AttributeValue)

	// Marshal each GeoAttributes instance and add its attributes to the merged map
	for _, geoAttr := range m.GeoAttributes {
		attrValue, err := geoAttr.MarshalDynamoDBAttributeValue()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal geo attributes: %w", err)
		}

		// Get the map from the AttributeValueMemberM
		if mv, ok := attrValue.(*types.AttributeValueMemberM); ok {
			// Add each key-value pair to our merged map
			maps.Copy(allAttrs, mv.Value)
		}
	}

	return &types.AttributeValueMemberM{Value: allAttrs}, nil
}

// UnmarshalDynamoDBAttributeValue implements custom unmarshaling for MultiGeoAttributes
func (m *MultiGeoAttributes[T]) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	memberM, ok := av.(*types.AttributeValueMemberM)
	if !ok {
		return fmt.Errorf("expected map attribute value, got %T", av)
	}

	// Initialize the map if it's nil
	if m.GeoAttributes == nil {
		m.GeoAttributes = make(map[string]GeoAttributes[T])
	}

	// Extract attribute values based on known geo indices
	// This relies on the fact that we know which geo indices to expect
	// In a complete implementation, we might need to have this information stored elsewhere
	for indexName, geoAttr := range m.GeoAttributes {
		// Create a copy of the attribute values with only those relevant to this index
		indexAttrs := make(map[string]types.AttributeValue)

		if geoAttr.geoHashKeyAttributeName != "" {
			if hashAv, ok := memberM.Value[geoAttr.geoHashKeyAttributeName]; ok {
				indexAttrs[geoAttr.geoHashKeyAttributeName] = hashAv
			}
		}

		if geoAttr.geoSortKeyAttributeName != "" {
			if sortKeyAv, ok := memberM.Value[geoAttr.geoSortKeyAttributeName]; ok {
				indexAttrs[geoAttr.geoSortKeyAttributeName] = sortKeyAv
			}
		}

		// Unmarshal into the specific GeoAttributes
		if err := geoAttr.UnmarshalDynamoDBAttributeValue(&types.AttributeValueMemberM{Value: indexAttrs}); err != nil {
			return fmt.Errorf("failed to unmarshal geo attributes for index %s: %w", indexName, err)
		}

		// Update the map with the unmarshaled version
		m.GeoAttributes[indexName] = geoAttr
	}

	return nil
}
