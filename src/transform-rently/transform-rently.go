package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

var PluginName = "transform-rently"

var ModifierRegisterer = registerer(PluginName)

type registerer string

func init() {
	fmt.Println(string(ModifierRegisterer), "loaded!!!")
}

type ResponseWrapper interface {
	Data() map[string]interface{}
	Io() io.Reader
	IsComplete() bool
	StatusCode() int
	Headers() map[string][]string
}

func (r registerer) RegisterModifiers(f func(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
)) {
	f(
		PluginName,
		r.modifierFactory,
		false,
		true,
	)
}

func (r registerer) modifierFactory(config map[string]interface{}) func(interface{}) (interface{}, error) {
	fmt.Printf("[PLUGIN: %s] Configurando plugin\n", PluginName)

	return func(input interface{}) (interface{}, error) {
		fmt.Println("[PLUGIN] Ejecutando transformación de respuesta")

		resp, ok := input.(ResponseWrapper)
		if !ok {
			fmt.Println("[PLUGIN] No es un ResponseWrapper válido")
			return input, nil
		}

		// Verificar que el reader no sea nil antes de leer
		reader := resp.Io()
		var body []byte
		var err error

		if reader != nil {
			// Leer el cuerpo de la respuesta si el reader no es nil
			body, err = io.ReadAll(reader)
			if err != nil {
				fmt.Printf("[PLUGIN] Error leyendo cuerpo: %v\n", err)
				return input, nil
			}
		} else {
			// Reader es nil, intentar usar Data() en su lugar
			fmt.Println("[PLUGIN] Reader es nil, usando Data()")
			return transformUsingData(resp), nil
		}

		// Si no hay cuerpo, devolver la respuesta original
		if len(body) == 0 {
			fmt.Println("[PLUGIN] Cuerpo vacío")
			return input, nil
		}

		fmt.Printf("[PLUGIN] Longitud del cuerpo: %d bytes\n", len(body))

		// Intentar transformar el cuerpo JSON
		modifiedBody, transformErr := transformBody(body)
		if transformErr != nil {
			fmt.Printf("[PLUGIN] Error transformando cuerpo: %v\n", transformErr)
			return input, nil
		}

		fmt.Printf("[PLUGIN] Transformación exitosa. Nuevo cuerpo: %d bytes\n", len(modifiedBody))

		// Devolver un nuevo ResponseWrapper con los datos modificados
		return &modifiedResponseWrapper{
			data:       parseJSON(modifiedBody),
			body:       modifiedBody,
			statusCode: resp.StatusCode(),
			headers:    resp.Headers(),
			isComplete: resp.IsComplete(),
		}, nil
	}
}

// transformBody transforma el cuerpo JSON
func transformBody(body []byte) ([]byte, error) {
	// Primero intentar parsear como objeto con clave "collection"
	var responseObj map[string]interface{}
	if err := json.Unmarshal(body, &responseObj); err == nil {
		if collection, exists := responseObj["collection"]; exists {
			if cars, ok := collection.([]interface{}); ok {
				completeResponse := buildCompleteResponse(cars)
				return json.Marshal(completeResponse)
			}
		}
	}

	// Si falla, intentar parsear como array directo
	var originalData []map[string]interface{}
	if err := json.Unmarshal(body, &originalData); err == nil {
		var cars []interface{}
		for _, car := range originalData {
			cars = append(cars, car)
		}
		completeResponse := buildCompleteResponse(cars)
		return json.Marshal(completeResponse)
	}

	return nil, fmt.Errorf("no se pudo parsear el JSON en el formato esperado")
}

// buildCompleteResponse construye la respuesta completa con pickup, dropoff y results
func buildCompleteResponse(collection []interface{}) map[string]interface{} {
	// Extraer información de pickup y dropoff del primer vehículo (si existe)
	var pickup, dropoff map[string]interface{}

	if len(collection) > 0 {
		if firstCar, ok := collection[0].(map[string]interface{}); ok {
			pickup = extractLocationInfo(firstCar, "deliveryPlace", "fromDate")
			dropoff = extractLocationInfo(firstCar, "returnPlace", "toDate")
		}
	}

	// Transformar la colección de vehículos
	results := transformCollection(collection)

	return map[string]interface{}{
		"pickup":  pickup,
		"dropoff": dropoff,
		"results": results,
	}
}

// extractLocationInfo extrae información de ubicación y fecha con todos los detalles
func extractLocationInfo(car map[string]interface{}, placeKey, dateKey string) map[string]interface{} {
	locationInfo := make(map[string]interface{})

	// Extraer información del lugar (deliveryPlace o returnPlace)
	if place, exists := car[placeKey]; exists {
		if placeMap, ok := place.(map[string]interface{}); ok {
			// Información básica de ubicación
			if iata, exists := placeMap["iata"].(string); exists && iata != "" {
				locationInfo["location"] = iata
				locationInfo["iata"] = iata
			}

			// Información detallada del lugar
			locationInfo["id"] = getFloat(placeMap, "id")
			locationInfo["email"] = getString(placeMap, "email")
			locationInfo["city"] = getString(placeMap, "city")
			locationInfo["country"] = getString(placeMap, "country")
			locationInfo["address2"] = getString(placeMap, "address2")
			locationInfo["zipCode"] = getString(placeMap, "zipCode")
			locationInfo["type"] = getString(placeMap, "type")
			locationInfo["serviceType"] = getString(placeMap, "serviceType")
			locationInfo["pickupInstructions"] = getString(placeMap, "pickupInstructions")

			// Coordenadas
			locationInfo["latitude"] = getFloat(placeMap, "latitude")
			locationInfo["longitude"] = getFloat(placeMap, "longitude")
		}
	}

	// Extraer fecha y hora
	if dateStr, exists := car[dateKey].(string); exists {
		// Formatear fecha: "2025-11-29T11:00:00Z" -> "2025-11-29"
		dateParts := strings.Split(dateStr, "T")
		if len(dateParts) > 0 {
			locationInfo["date"] = dateParts[0]
		}

		// Formatear hora: "2025-11-29T11:00:00Z" -> "1100"
		if len(dateParts) > 1 {
			timeParts := strings.Split(dateParts[1], ":")
			if len(timeParts) >= 2 {
				locationInfo["time"] = timeParts[0] + timeParts[1]
			}
		}
	}

	// Si no se pudo extraer información esencial, usar valores por defecto
	if locationInfo["location"] == nil {
		locationInfo["location"] = "MIA"
		locationInfo["iata"] = "MIA"
	}
	if locationInfo["date"] == nil {
		if placeKey == "deliveryPlace" {
			locationInfo["date"] = "2025-11-21"
		} else {
			locationInfo["date"] = "2025-11-26"
		}
	}
	if locationInfo["time"] == nil {
		locationInfo["time"] = "1200"
	}

	return locationInfo
}

// transformUsingData transforma usando el método Data() cuando Io() es nil
func transformUsingData(resp ResponseWrapper) interface{} {
	data := resp.Data()
	if data == nil {
		return resp
	}

	fmt.Printf("[PLUGIN] Transformando desde Data(): %+v\n", data)

	// Intentar extraer datos de la respuesta
	if collection, exists := data["collection"]; exists {
		if cars, ok := collection.([]interface{}); ok {
			completeResponse := buildCompleteResponse(cars)

			// Convertir a JSON
			modifiedBody, err := json.Marshal(completeResponse)
			if err != nil {
				return resp
			}

			return &modifiedResponseWrapper{
				data:       completeResponse,
				body:       modifiedBody,
				statusCode: resp.StatusCode(),
				headers:    resp.Headers(),
				isComplete: resp.IsComplete(),
			}
		}
	}

	return resp
}

// transformCollection transforma la colección de vehículos
func transformCollection(collection []interface{}) []map[string]interface{} {
	var simplified []map[string]interface{}

	for _, item := range collection {
		if car, ok := item.(map[string]interface{}); ok {
			simplifiedCar := extractSimplifiedCarInfo(car)
			simplified = append(simplified, simplifiedCar)
		}
	}

	return simplified
}

// extractSimplifiedCarInfo extrae la información simplificada del auto
func extractSimplifiedCarInfo(car map[string]interface{}) map[string]interface{} {
	simplified := map[string]interface{}{
		"nombre":               getString(car, "model", "name"),
		"marca":                getString(car, "model", "brand"),
		"tipo":                 getString(car, "category", "name"),
		"tarifa":               getFloat(car, "price"),
		"tarifa_con_impuestos": getFloat(car, "customerPrice"),
		"tarifa_sin_impuestos": calculatePriceWithoutTax(car),
		"proveedor":            getString(car, "supplier", "name"),
		"franquicia":           getFloat(car, "franchise"),
		"dias_totales":         getDays(car, "totalDays"),
	}

	fmt.Printf("[PLUGIN] Auto transformado: %s - %.2f USD\n",
		getString(car, "model", "name"), getFloat(car, "customerPrice"))

	return simplified
}

// calculatePriceWithoutTax calcula el precio base sin impuestos
func calculatePriceWithoutTax(car map[string]interface{}) float64 {
	var basePrice float64

	if priceItems, ok := car["priceItems"].([]interface{}); ok {
		for _, item := range priceItems {
			if priceItem, ok := item.(map[string]interface{}); ok {
				if itemType, ok := priceItem["type"].(string); ok && itemType == "Booking" {
					if price, ok := priceItem["price"].(float64); ok {
						basePrice += price
					}
				}
			}
		}
	}

	return basePrice
}

// Helper functions para acceder de forma segura a los datos anidados
func getString(data map[string]interface{}, keys ...string) string {
	current := data
	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key].(string); ok {
				return val
			}
			return ""
		}
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return ""
		}
	}
	return ""
}

func getFloat(data map[string]interface{}, key string) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	return 0
}

func getDays(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func parseJSON(data []byte) interface{} {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

// modifiedResponseWrapper implementa ResponseWrapper para la respuesta modificada
type modifiedResponseWrapper struct {
	data       interface{}
	body       []byte
	statusCode int
	headers    map[string][]string
	isComplete bool
}

func (m *modifiedResponseWrapper) Data() map[string]interface{} {
	if m.data == nil {
		return map[string]interface{}{}
	}

	// Si ya es un map, devolverlo directamente
	if dataMap, ok := m.data.(map[string]interface{}); ok {
		return dataMap
	}

	// Envolver en un mapa
	return map[string]interface{}{
		"data": m.data,
	}
}

func (m *modifiedResponseWrapper) Io() io.Reader {
	if m.body == nil {
		return bytes.NewReader([]byte{})
	}
	return bytes.NewReader(m.body)
}

func (m *modifiedResponseWrapper) IsComplete() bool {
	return m.isComplete
}

func (m *modifiedResponseWrapper) StatusCode() int {
	return m.statusCode
}

func (m *modifiedResponseWrapper) Headers() map[string][]string {
	if m.headers == nil {
		return make(map[string][]string)
	}
	return m.headers
}

func main() {}
