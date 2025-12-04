package redis

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type KeyGenerator interface {
	Generate() string
}

type TimestampKeyGenerator struct {
	Prefix string
}

func (kg *TimestampKeyGenerator) Generate() string {
	timestamp := time.Now().UnixNano()
	if kg.Prefix != "" {
		return fmt.Sprintf("%s%d", kg.Prefix, timestamp)
	}
	return fmt.Sprintf("%d", timestamp)
}

type UUIDKeyGenerator struct {
	Prefix string
}

func (kg *UUIDKeyGenerator) Generate() string {
	id := uuid.New().String()
	if kg.Prefix != "" {
		return kg.Prefix + id
	}
	return id
}

type RandomKeyGenerator struct {
	Prefix string
	Length int
}

func (kg *RandomKeyGenerator) Generate() string {
	length := kg.Length
	if length == 0 {
		length = 16
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp if random fails
		return (&TimestampKeyGenerator{Prefix: kg.Prefix}).Generate()
	}

	key := base64.URLEncoding.EncodeToString(bytes)
	if kg.Prefix != "" {
		return kg.Prefix + key
	}
	return key
}

type CompositeKeyGenerator struct {
	Parts []string
}

func (kg *CompositeKeyGenerator) Generate() string {
	if len(kg.Parts) == 0 {
		return (&TimestampKeyGenerator{}).Generate()
	}

	result := kg.Parts[0]
	for i := 1; i < len(kg.Parts); i++ {
		result += ":" + kg.Parts[i]
	}
	return result
}

func NewTimestampKey(prefix string) string {
	kg := &TimestampKeyGenerator{Prefix: prefix}
	return kg.Generate()
}

func NewUUIDKey(prefix string) string {
	kg := &UUIDKeyGenerator{Prefix: prefix}
	return kg.Generate()
}

func NewRandomKey(prefix string, length int) string {
	kg := &RandomKeyGenerator{Prefix: prefix, Length: length}
	return kg.Generate()
}

func NewCompositeKey(parts ...string) string {
	kg := &CompositeKeyGenerator{Parts: parts}
	return kg.Generate()
}
