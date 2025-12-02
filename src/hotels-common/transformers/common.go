package transformers

import (
	"hotels-common/models"
)

type HotelTransformer interface {
	Transform(data map[string]interface{}, config models.TransformationConfig) (models.StandardResponse, error)
	CanTransform(data map[string]interface{}) bool
	GetProvider() models.Provider
	PriceCheckStrategy(data map[string]interface{}, config models.TransformationConfig) string
}

type BaseTransformer struct {
	provider models.Provider
}

func (bt *BaseTransformer) GetProvider() models.Provider {
	return bt.provider
}

type TransformerRegistry struct {
	transformers map[models.Provider]HotelTransformer
}

func NewTransformerRegistry() *TransformerRegistry {
	return &TransformerRegistry{
		transformers: make(map[models.Provider]HotelTransformer),
	}
}

func (tr *TransformerRegistry) Register(transformer HotelTransformer) {
	tr.transformers[transformer.GetProvider()] = transformer
}

func (tr *TransformerRegistry) Get(provider models.Provider) (HotelTransformer, bool) {
	transformer, exists := tr.transformers[provider]
	return transformer, exists
}

func (tr *TransformerRegistry) FindByData(data map[string]interface{}) (HotelTransformer, bool) {
	for _, transformer := range tr.transformers {
		if transformer.CanTransform(data) {
			return transformer, true
		}
	}
	return nil, false
}

func (tr *TransformerRegistry) GetAll() map[models.Provider]HotelTransformer {
	return tr.transformers
}

func GetStringValue(m map[string]interface{}, key string) interface{} {
	if value, exists := m[key]; exists {
		return value
	}
	return nil
}

func GetArrayValue(m map[string]interface{}, key string) []interface{} {
	if value, exists := m[key]; exists {
		if array, ok := value.([]interface{}); ok {
			return array
		}
	}
	return []interface{}{}
}

func GetMapValue(m map[string]interface{}, key string) map[string]interface{} {
	if value, exists := m[key]; exists {
		if mapVal, ok := value.(map[string]interface{}); ok {
			return mapVal
		}
	}
	return nil
}
