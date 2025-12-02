package core

import (
	"fmt"
	"hotels-common/adapters"
	"hotels-common/models"
	"hotels-common/transformers"
	"log"
)

type HotelPluginCore struct {
	registry  *transformers.TransformerRegistry
	extractor *adapters.ResponseExtractor
	adapter   *adapters.OutputAdapter
}

func NewHotelPluginCore() *HotelPluginCore {
	return &HotelPluginCore{
		registry:  transformers.NewTransformerRegistry(),
		extractor: adapters.NewResponseExtractor(),
		adapter:   adapters.NewOutputAdapter(),
	}
}

func (hpc *HotelPluginCore) RegisterTransformer(transformer transformers.HotelTransformer) {
	hpc.registry.Register(transformer)
}

func (hpc *HotelPluginCore) ProcessResponse(input interface{}, config models.TransformationConfig) (interface{}, error) {
	data, success := hpc.extractor.ExtractData(input)
	if !success {
		log.Printf("[HOTEL-CORE] Could not extract data, returning original input")
		return input, nil
	}

	var transformer transformers.HotelTransformer
	var found bool

	if config.Provider != "" {
		transformer, found = hpc.registry.Get(config.Provider)
	}

	if !found {
		transformer, found = hpc.registry.FindByData(map[string]interface{}(data))
	}

	if !found {
		log.Printf("[HOTEL-CORE] No transformer found, adding provider at root level")
		data["provider"] = config.Provider
		return hpc.adapter.AdaptOutput(models.StandardResponse{}, input), nil
	}

	log.Printf("[HOTEL-CORE] Applying transformation using %s transformer", transformer.GetProvider())
	standardResponse, err := transformer.Transform(map[string]interface{}(data), config)
	if err != nil {
		return nil, fmt.Errorf("error during transformation: %w", err)
	}

	return hpc.adapter.AdaptOutput(standardResponse, input), nil
}

func (hpc *HotelPluginCore) CreateModifierFactory(defaultProvider models.Provider) func(map[string]interface{}) func(interface{}) (interface{}, error) {
	return func(cfg map[string]interface{}) func(interface{}) (interface{}, error) {
		config := hpc.extractConfig(cfg, defaultProvider)

		log.Printf("[HOTEL-CORE] Plugin initialized with config: provider=%s, extra=%v",
			config.Provider, config.ExtraConfig)

		return func(input interface{}) (interface{}, error) {
			return hpc.ProcessResponse(input, config)
		}
	}
}

func (hpc *HotelPluginCore) extractConfig(cfg map[string]interface{}, defaultProvider models.Provider) models.TransformationConfig {
	config := models.TransformationConfig{
		Provider:    defaultProvider,
		ExtraConfig: make(map[string]interface{}),
	}

	if provider, ok := cfg["provider"]; ok {
		if providerStr, ok := provider.(string); ok {
			config.Provider = models.Provider(providerStr)
		}
	}

	for key, value := range cfg {
		if key != "provider" {
			config.ExtraConfig[key] = value
		}
	}

	return config
}

func (hpc *HotelPluginCore) GetRegisteredProviders() []models.Provider {
	providers := make([]models.Provider, 0)
	for provider := range hpc.registry.GetAll() {
		providers = append(providers, provider)
	}
	return providers
}
