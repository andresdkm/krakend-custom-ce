package main

import (
	"fmt"
	"hotels-common/models"
	"hotels-common/transformers"
	"log"
	"math/rand"
	"redis"
	"strconv"
)

type ExpediaTransformer struct {
	transformers.BaseTransformer
}

func ExpediaTransformerImpl() *ExpediaTransformer {
	return &ExpediaTransformer{
		BaseTransformer: transformers.BaseTransformer{},
	}
}
func (t *ExpediaTransformer) PriceCheckStrategy(data map[string]interface{}, config models.TransformationConfig) string {
	redisConfig, err := redis.LoadConfig(config.ExtraConfig)
	if err != nil {
		_ = fmt.Errorf("error loading Redis config: %v", err)
	}
	redisClient, err := redis.NewRedisClient(redisConfig)
	if err != nil {
		_ = fmt.Errorf("error creating Redis client: %v", err)
	}
	key := redis.NewUUIDKey("expedia:pricecheck:")
	err = redisClient.Set(key, data)
	if err != nil {
		return ""
	}
	return strconv.Itoa(rand.Intn(10000000))
}

func (t *ExpediaTransformer) GetProvider() models.Provider {
	return models.ProviderExpedia
}

func (t *ExpediaTransformer) CanTransform(data map[string]interface{}) bool {
	_, hasHotelResult := data["results"]
	return hasHotelResult
}

func (t *ExpediaTransformer) Transform(data map[string]interface{}, config models.TransformationConfig) (models.StandardResponse, error) {
	response := models.StandardResponse{
		Hotels: []models.BaseHotel{},
	}

	if hotelResult, exists := data["results"]; exists {
		if hotelArray, ok := hotelResult.([]interface{}); ok {
			transformedHotels, err := t.transformHotelArray(hotelArray, config)
			if err != nil {
				return response, err
			}
			response.Hotels = transformedHotels
			log.Printf("[EXPEDIA-TRANSFORMER] Successfully transformed %d hotels", len(transformedHotels))
		}
	}
	if _, exists := data["property_id"]; exists {
		log.Printf("[EXPEDIA-TRANSFORMER] Detected single hotel result")
		transformedHotel := t.transformSingleHotel(data, config)
		transformedHotel.PriceCheckKey = t.PriceCheckStrategy(data, config)
		response.Hotels = append(response.Hotels, transformedHotel)
	}

	return response, nil
}

func (t *ExpediaTransformer) transformHotelArray(hotelArray []interface{}, config models.TransformationConfig) ([]models.BaseHotel, error) {
	hotels := make([]models.BaseHotel, 0, len(hotelArray))
	for _, property := range hotelArray {
		if propertyResult, exists := property.(map[string]interface{})["properties"]; exists {
			if propertyArray, ok := propertyResult.([]interface{}); ok {
				for _, hotel := range propertyArray {
					if hotelMap, ok := hotel.(map[string]interface{}); ok {
						transformedHotel := t.transformSingleHotel(hotelMap, config)
						hotels = append(hotels, transformedHotel)
					}
				}
			}
		}
	}
	return hotels, nil
}

func (t *ExpediaTransformer) transformSingleHotel(hotelMap map[string]interface{}, config models.TransformationConfig) models.BaseHotel {
	hotel := models.BaseHotel{
		Provider:    models.ProviderExpedia,
		ProductID:   transformers.GetStringValue(hotelMap, "property_id"),
		ProductName: transformers.GetStringValue(hotelMap, "title"),
		ProductType: nil,
		Images:      []interface{}{},
		Score:       transformers.GetStringValue(hotelMap, "stars"),
		Rank:        transformers.GetStringValue(hotelMap, "rank"),
		Refundable:  nil,
		Amenities:   []interface{}{},
		GeographicReference: models.GeographicReference{
			Latitude:  nil,
			Longitude: nil,
			Distance: models.Distance{
				Unit:  nil,
				Value: nil,
			},
		},
		StartDate: "2025-12-25T00:00:00.000Z", // TODO fechas del request
		EndDate:   "2025-12-27T00:00:00.000Z", // TODO fechas del request
		Days:      2,
		Rate: models.Rate{
			BasePrice:  nil,
			Taxes:      nil,
			TotalPrice: nil,
			Currency:   nil,
		},
		Fees:          []interface{}{},
		ProviderRates: []interface{}{},
	}

	return hotel
}
