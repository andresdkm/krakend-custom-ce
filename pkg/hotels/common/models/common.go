package models

import "time"

type Provider string

const (
	ProviderTBO     Provider = "TBO"
	ProviderExpedia Provider = "Expedia"
)

type StandardResponse struct {
	Hotels []BaseHotel `json:"hotels"`
}

type BaseHotel struct {
	Provider            Provider            `json:"provider"`
	ProductID           interface{}         `json:"productId"`
	ProductName         interface{}         `json:"productName"`
	ProductType         interface{}         `json:"productType"`
	Images              []interface{}       `json:"images"`
	Score               interface{}         `json:"score"`
	Rank                interface{}         `json:"rank"`
	Refundable          interface{}         `json:"refundable"`
	Amenities           []interface{}       `json:"amenities"`
	GeographicReference GeographicReference `json:"geographicReference"`
	StartDate           string              `json:"startDate"`
	EndDate             string              `json:"endDate"`
	Days                interface{}         `json:"days"`
	Rate                Rate                `json:"rate"`
	Fees                []interface{}       `json:"fees"`
	ProviderRates       []interface{}       `json:"providerRates"`
	PriceCheckKey       string              `json:"priceCheckKey,omitempty"`
}

type GeographicReference struct {
	Latitude  interface{} `json:"latitude"`
	Longitude interface{} `json:"longitude"`
	Distance  Distance    `json:"distance"`
}

type Distance struct {
	Unit  interface{} `json:"unit"`
	Value interface{} `json:"value"`
}

type Rate struct {
	BasePrice  interface{} `json:"basePrice"`
	Taxes      interface{} `json:"taxes"`
	TotalPrice interface{} `json:"totalPrice"`
	Currency   interface{} `json:"currency"`
}

type TransformationConfig struct {
	Provider    Provider               `json:"provider"`
	StartDate   *time.Time             `json:"startDate,omitempty"`
	EndDate     *time.Time             `json:"endDate,omitempty"`
	ExtraConfig map[string]interface{} `json:"extraConfig,omitempty"`
}

type ResponseData map[string]interface{}

type ProviderConfig struct {
	Name     Provider               `json:"name"`
	Enabled  bool                   `json:"enabled"`
	Settings map[string]interface{} `json:"settings"`
}
