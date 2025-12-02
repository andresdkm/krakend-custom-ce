package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ctx           = context.Background()
	clientOnce    sync.Once
	clientManager *ClientManager
)

type ClientManager struct {
	clients map[string]*redis.Client
	mu      sync.RWMutex
}

func GetClientManager() *ClientManager {
	clientOnce.Do(func() {
		clientManager = &ClientManager{
			clients: make(map[string]*redis.Client),
		}
	})
	return clientManager
}

func (cm *ClientManager) GetOrCreateClient(config *Config) (*redis.Client, error) {
	clientKey := fmt.Sprintf("%s:%d", config.RedisAddr, config.RedisDB)

	cm.mu.RLock()
	client, exists := cm.clients[clientKey]
	cm.mu.RUnlock()

	if exists && client != nil {
		if err := client.Ping(ctx).Err(); err == nil {
			return client, nil
		}
		cm.mu.Lock()
		delete(cm.clients, clientKey)
		cm.mu.Unlock()
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	client = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error connecting to Redis at %s: %v", config.RedisAddr, err)
	}

	cm.clients[clientKey] = client
	fmt.Printf("Redis connection established: %s (DB: %d)\n", config.RedisAddr, config.RedisDB)

	return client, nil
}

func (cm *ClientManager) CloseAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for key, client := range cm.clients {
		if err := client.Close(); err != nil {
			fmt.Printf("Error closing Redis connection %s: %v\n", key, err)
		}
	}
	cm.clients = make(map[string]*redis.Client)
}

type RedisClient struct {
	client *redis.Client
	config *Config
}

func NewRedisClient(config *Config) (*RedisClient, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}

	cm := GetClientManager()
	client, err := cm.GetOrCreateClient(config)
	if err != nil {
		return nil, err
	}

	return &RedisClient{
		client: client,
		config: config,
	}, nil
}

func (rc *RedisClient) Set(key string, value interface{}) error {
	var data []byte
	var err error

	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data, err = json.Marshal(value)
		if err != nil {
			return fmt.Errorf("error marshaling value: %v", err)
		}
	}

	fullKey := rc.config.KeyPrefix + key
	ttl := time.Duration(rc.config.KeyTTL) * time.Second

	return rc.client.Set(ctx, fullKey, data, ttl).Err()
}

func (rc *RedisClient) Get(key string) (string, error) {
	fullKey := rc.config.KeyPrefix + key
	return rc.client.Get(ctx, fullKey).Result()
}

func (rc *RedisClient) GetJSON(key string, dest interface{}) error {
	data, err := rc.Get(key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

func (rc *RedisClient) Delete(key string) error {
	fullKey := rc.config.KeyPrefix + key
	return rc.client.Del(ctx, fullKey).Err()
}

func (rc *RedisClient) Exists(key string) (bool, error) {
	fullKey := rc.config.KeyPrefix + key
	count, err := rc.client.Exists(ctx, fullKey).Result()
	return count > 0, err
}

func (rc *RedisClient) Expire(key string, ttl int) error {
	fullKey := rc.config.KeyPrefix + key
	return rc.client.Expire(ctx, fullKey, time.Duration(ttl)*time.Second).Err()
}

func (rc *RedisClient) Increment(key string) (int64, error) {
	fullKey := rc.config.KeyPrefix + key
	return rc.client.Incr(ctx, fullKey).Result()
}

func (rc *RedisClient) SetNX(key string, value interface{}, ttl int) (bool, error) {
	var data []byte
	var err error

	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data, err = json.Marshal(value)
		if err != nil {
			return false, fmt.Errorf("error marshaling value: %v", err)
		}
	}

	fullKey := rc.config.KeyPrefix + key
	duration := time.Duration(ttl) * time.Second

	return rc.client.SetNX(ctx, fullKey, data, duration).Result()
}

func (rc *RedisClient) Keys(pattern string) ([]string, error) {
	fullPattern := rc.config.KeyPrefix + pattern
	keys, err := rc.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return nil, err
	}

	prefixLen := len(rc.config.KeyPrefix)
	cleanKeys := make([]string, len(keys))
	for i, key := range keys {
		cleanKeys[i] = key[prefixLen:]
	}

	return cleanKeys, nil
}

func (rc *RedisClient) Ping() error {
	return rc.client.Ping(ctx).Err()
}

func (rc *RedisClient) GetClient() *redis.Client {
	return rc.client
}
