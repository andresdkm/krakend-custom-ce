package main

import (
	"hotels-common/models"
	"hotels-common/transformers"
	"log"
)

type TBOTransformer struct {
	transformers.BaseTransformer
}

func NewTBOTransformer() *TBOTransformer {
	return &TBOTransformer{
		BaseTransformer: transformers.BaseTransformer{},
	}
}

func (t *TBOTransformer) PriceCheckStrategy(data map[string]interface{}, config models.TransformationConfig) string {
	return "TBO_PRICE_CHECK"
}

func (t *TBOTransformer) GetProvider() models.Provider {
	return models.ProviderTBO
}

func (t *TBOTransformer) CanTransform(data map[string]interface{}) bool {
	_, hasHotelResult := data["HotelResult"]
	return hasHotelResult
}

func (t *TBOTransformer) Transform(data map[string]interface{}, config models.TransformationConfig) (models.StandardResponse, error) {
	log.Printf("[TBO-TRANSFORMER] Starting TBO transformation...")

	response := models.StandardResponse{
		Hotels: []models.BaseHotel{},
	}

	if hotelResult, exists := data["HotelResult"]; exists {
		if hotelArray, ok := hotelResult.([]interface{}); ok {
			transformedHotels, err := t.transformHotelArray(hotelArray, config)
			if err != nil {
				return response, err
			}
			response.Hotels = transformedHotels
			log.Printf("[TBO-TRANSFORMER] Successfully transformed %d hotels", len(transformedHotels))
		}
	}

	return response, nil
}

func (t *TBOTransformer) transformHotelArray(hotelArray []interface{}, config models.TransformationConfig) ([]models.BaseHotel, error) {
	hotels := make([]models.BaseHotel, 0, len(hotelArray))

	for _, hotel := range hotelArray {
		if hotelMap, ok := hotel.(map[string]interface{}); ok {
			transformedHotel := t.transformSingleHotel(hotelMap, config)
			hotels = append(hotels, transformedHotel)
		}
	}

	return hotels, nil
}

func (t *TBOTransformer) transformSingleHotel(hotelMap map[string]interface{}, config models.TransformationConfig) models.BaseHotel {
	hotel := models.BaseHotel{
		Provider:    models.ProviderTBO,
		ProductID:   transformers.GetStringValue(hotelMap, "HotelCode"),
		ProductName: nil,
		ProductType: nil,
		Images:      []interface{}{},
		Score:       nil,
		Rank:        nil,
		Refundable:  t.extractRefundableFromRooms(hotelMap),
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
			Currency:   transformers.GetStringValue(hotelMap, "Currency"),
		},
		Fees:          []interface{}{},
		ProviderRates: []interface{}{},
	}

	return hotel
}

func (t *TBOTransformer) extractRefundableFromRooms(hotelMap map[string]interface{}) interface{} {
	if roomsArray := transformers.GetArrayValue(hotelMap, "Rooms"); len(roomsArray) > 0 {
		if firstRoom, ok := roomsArray[0].(map[string]interface{}); ok {
			if refundable, exists := firstRoom["refundable"]; exists {
				return refundable
			}
		}
	}
	return true
}
