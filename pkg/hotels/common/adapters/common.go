package adapters

import (
	"encoding/json"
	"kraken-builder-plugins/pkg/hotels/common/models"
	"log"
	"math/rand"
	"reflect"
	"strconv"
)

type ResponseExtractor struct{}

func NewResponseExtractor() *ResponseExtractor {
	return &ResponseExtractor{}
}

func (re *ResponseExtractor) ExtractData(input interface{}) (models.ResponseData, bool) {
	if bytes, ok := input.([]byte); ok {
		return re.extractFromBytes(bytes)
	}

	if mapData, ok := input.(map[string]interface{}); ok {
		return models.ResponseData(mapData), true
	}

	if data := re.extractUsingReflection(input); data != nil {
		return models.ResponseData(data), true
	}

	return nil, false
}

func (re *ResponseExtractor) extractFromBytes(data []byte) (models.ResponseData, bool) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("[COMMON-EXTRACTOR] Error unmarshaling bytes: %v", err)
		return nil, false
	}
	return models.ResponseData(result), true
}

func (re *ResponseExtractor) extractUsingReflection(input interface{}) map[string]interface{} {
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if data := re.callDataMethod(val); data != nil {
		return data
	}

	if data := re.findDataField(val); data != nil {
		return data
	}

	return nil
}

func (re *ResponseExtractor) callDataMethod(val reflect.Value) map[string]interface{} {
	dataMethod := val.MethodByName("Data")
	if dataMethod.IsValid() {
		results := dataMethod.Call(nil)
		if len(results) > 0 {
			if dataMap, ok := results[0].Interface().(map[string]interface{}); ok {
				return dataMap
			}
		}
	}
	return nil
}

func (re *ResponseExtractor) findDataField(val reflect.Value) map[string]interface{} {
	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			fieldValue := val.Field(i)

			if field.Name == "Data" || field.Name == "data" {
				if fieldValue.CanInterface() {
					if dataMap, ok := fieldValue.Interface().(map[string]interface{}); ok {
						return dataMap
					}
				}
			}
		}
	}
	return nil
}

type OutputAdapter struct{}

func NewOutputAdapter() *OutputAdapter {
	return &OutputAdapter{}
}

func (oa *OutputAdapter) AdaptOutput(response models.StandardResponse, originalInput interface{}) interface{} {
	outputMap := map[string]interface{}{
		"hotels": response.Hotels,
	}

	val := reflect.ValueOf(originalInput)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		dataMethod := val.MethodByName("Data")
		if dataMethod.IsValid() {
			results := dataMethod.Call(nil)
			if len(results) > 0 {
				if dataMapInterface := results[0].Interface(); dataMapInterface != nil {
					if dataMap, ok := dataMapInterface.(map[string]interface{}); ok {
						for k := range dataMap {
							delete(dataMap, k)
						}
						for _, v := range outputMap {
							dataMap[strconv.Itoa(rand.Intn(100))] = v
						}
						return originalInput
					}
				}
			}
		}
	}

	if _, wasBytes := originalInput.([]byte); wasBytes {
		if jsonBytes, err := json.Marshal(outputMap); err == nil {
			return jsonBytes
		}
	}
	return outputMap
}
