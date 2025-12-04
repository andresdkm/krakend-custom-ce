package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
)

func main() {}

func init() {
	fmt.Println("hotels-provider plugin loaded!!!")
}

var ClientRegisterer = registerer("hotels-provider")
var HandlerRegisterer = handlerRegisterer("hotels-provider")
var ModifierRegisterer = registerer("hotels-provider")

type registerer string

func (r registerer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r registerer) registerClients(ctx context.Context, extra map[string]interface{}) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), nil
}

type handlerRegisterer string

func (hr handlerRegisterer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(hr), hr.registerHandlers)
}

func (hr handlerRegisterer) registerHandlers(ctx context.Context, extra map[string]interface{}) (http.Handler, error) {
	log.Printf("[HOTELS-PROVIDER] Handler registerer called")
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("[HOTELS-PROVIDER] Handler processing request")
		w.WriteHeader(http.StatusOK)
	}), nil
}

func (r registerer) RegisterModifiers(f func(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
)) {
	f(string(r), r.modifierFactory, false, true)
	fmt.Println(string(r), "registered!!!")
}

type ResponseWrapper interface {
	Data() map[string]interface{}
	Io() io.Reader
	IsComplete() bool
	StatusCode() int
	Headers() map[string][]string
}

func (r registerer) modifierFactory(cfg map[string]interface{}) func(interface{}) (interface{}, error) {
	providerName := "TBO"
	if provider, ok := cfg["provider"]; ok {
		if providerStr, ok := provider.(string); ok {
			providerName = providerStr
		}
	}

	log.Printf("[HOTELS-PROVIDER] Plugin initialized with provider: %s", providerName)

	return func(input interface{}) (interface{}, error) {
		var inputMap map[string]interface{}

		if bytes, ok := input.([]byte); ok {
			if err := json.Unmarshal(bytes, &inputMap); err != nil {
				log.Printf("[HOTELS-PROVIDER] Error unmarshaling bytes: %v", err)
				return input, nil
			}
		} else if mapData, ok := input.(map[string]interface{}); ok {
			inputMap = mapData
		} else if wrapper, ok := input.(ResponseWrapper); ok {
			inputMap = wrapper.Data()
		} else {
			if data := extractDataFromWrapper(input); data != nil {
				inputMap = data
			} else {
				return input, nil
			}
		}

		if hotelResult, exists := inputMap["HotelResult"]; exists {
			log.Printf("[HOTELS-PROVIDER] Found HotelResult, processing...")

			if hotelArray, ok := hotelResult.([]interface{}); ok {
				transformedHotels := make([]interface{}, 0, len(hotelArray))

				for _, hotel := range hotelArray {
					if hotelMap, ok := hotel.(map[string]interface{}); ok {
						transformedHotel := transformTBOHotel(hotelMap, providerName)
						transformedHotels = append(transformedHotels, transformedHotel)
					}
				}

				inputMap["hotels"] = transformedHotels

				delete(inputMap, "HotelResult")
				delete(inputMap, "Status")
				delete(inputMap, "provider")
			}
		} else {
			inputMap["provider"] = providerName
		}

		if _, wasBytes := input.([]byte); wasBytes {
			if result, err := json.Marshal(inputMap); err == nil {
				return result, nil
			}
		}

		return inputMap, nil
	}
}

func transformTBOHotel(hotelMap map[string]interface{}, providerName string) map[string]interface{} {
	// Create the specific structure for TBO
	transformed := map[string]interface{}{
		"provider":    providerName,
		"productId":   getStringValue(hotelMap, "productId"),
		"productName": nil,
		"productType": nil,
		"images":      []interface{}{},
		"score":       nil,
		"rank":        nil,
		"refundable":  getRefundableFromRooms(hotelMap),
		"amenities":   []interface{}{},
		"geographicReference": map[string]interface{}{
			"latitude":  nil,
			"longitude": nil,
			"distance": map[string]interface{}{
				"unit":  nil,
				"value": nil,
			},
		},
		"startDate": "2025-12-14T18:00:00.000-03:00", // TODO: extract from request
		"endDate":   "2025-12-17T18:00:00.000-03:00", // TODO: extract from request
		"days":      nil,
		"rate": map[string]interface{}{
			"basePrice":  nil,
			"taxes":      nil,
			"totalPrice": nil,
			"currency":   getStringValue(hotelMap, "currency"),
		},
		"fees":          []interface{}{map[string]interface{}{}},
		"providerRates": []interface{}{map[string]interface{}{}},
	}

	return transformed
}

func getStringValue(m map[string]interface{}, key string) interface{} {
	if value, exists := m[key]; exists {
		return value
	}
	return nil
}

func getRefundableFromRooms(hotelMap map[string]interface{}) interface{} {
	if rooms, exists := hotelMap["Rooms"]; exists {
		if roomsArray, ok := rooms.([]interface{}); ok && len(roomsArray) > 0 {
			if firstRoom, ok := roomsArray[0].(map[string]interface{}); ok {
				if refundable, exists := firstRoom["refundable"]; exists {
					return refundable
				}
			}
		}
	}
	return true
}

func extractDataFromWrapper(input interface{}) map[string]interface{} {
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Look for Data() method
	dataMethod := val.MethodByName("Data")
	if dataMethod.IsValid() {
		log.Printf("[HOTELS-PROVIDER] Found Data() method, calling it...")
		results := dataMethod.Call(nil)
		if len(results) > 0 {
			if dataMap, ok := results[0].Interface().(map[string]interface{}); ok {
				log.Printf("[HOTELS-PROVIDER] Successfully extracted data from Data() method")
				return dataMap
			}
		}
	}

	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			fieldValue := val.Field(i)

			if field.Name == "Data" || field.Name == "data" {
				if fieldValue.CanInterface() {
					if dataMap, ok := fieldValue.Interface().(map[string]interface{}); ok {
						log.Printf("[HOTELS-PROVIDER] Found data in field %s", field.Name)
						return dataMap
					}
				}
			}
		}
	}

	return nil
}
