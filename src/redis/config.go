package redis

import (
	"encoding/json"
	"fmt"
)

type Config struct {
	RedisAddr     string `json:"redis_addr"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`
	KeyPrefix     string `json:"key_prefix"`
	KeyTTL        int    `json:"key_ttl"`
}

func DefaultConfig() *Config {
	return &Config{
		RedisAddr:     "redis:6379",
		RedisPassword: "",
		RedisDB:       0,
		KeyPrefix:     "krakend:",
		KeyTTL:        3600,
	}
}

func LoadConfig(extra map[string]interface{}) (*Config, error) {
	config := DefaultConfig()

	configData, err := json.Marshal(extra)
	if err != nil {
		return nil, fmt.Errorf("error marshaling config: %v", err)
	}

	if err := json.Unmarshal(configData, config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	if config.RedisAddr == "" {
		config.RedisAddr = "localhost:6379"
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "krakend:"
	}
	if config.KeyTTL == 0 {
		config.KeyTTL = 3600
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.RedisAddr == "" {
		return fmt.Errorf("redis_addr is required")
	}
	if c.KeyTTL < 0 {
		return fmt.Errorf("key_ttl must be greater than or equal to 0")
	}
	if c.RedisDB < 0 || c.RedisDB > 15 {
		return fmt.Errorf("redis_db must be between 0 and 15")
	}
	return nil
}
