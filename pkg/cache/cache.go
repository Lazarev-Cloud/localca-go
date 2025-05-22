package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/redis/go-redis/v9"
)

// Cache interface defines the caching operations
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	InvalidatePattern(ctx context.Context, pattern string) error
	Close() error
}

// KeyDBCache implements the Cache interface using KeyDB
type KeyDBCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NoOpCache is a no-operation cache implementation for when caching is disabled
type NoOpCache struct{}

// NewCache creates a new cache instance based on configuration
func NewCache(cfg *config.Config) (Cache, error) {
	if !cfg.CacheEnabled {
		log.Println("Cache is disabled, using NoOpCache")
		return &NoOpCache{}, nil
	}

	// Create KeyDB client
	client := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%d", cfg.KeyDBHost, cfg.KeyDBPort),
		Password:    cfg.KeyDBPassword,
		DB:          cfg.KeyDBDB,
		PoolSize:    10,
		PoolTimeout: 30 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to KeyDB: %w", err)
	}

	log.Printf("Connected to KeyDB at %s:%d", cfg.KeyDBHost, cfg.KeyDBPort)

	return &KeyDBCache{
		client: client,
		ttl:    time.Duration(cfg.CacheTTL) * time.Second,
	}, nil
}

// KeyDBCache implementations

// Set stores a value in the cache with the specified TTL
func (c *KeyDBCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.ttl
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a value from the cache and unmarshals it into dest
func (c *KeyDBCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return nil
}

// Del removes one or more keys from the cache
func (c *KeyDBCache) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists checks if keys exist in the cache
func (c *KeyDBCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// InvalidatePattern removes all keys matching the pattern
func (c *KeyDBCache) InvalidatePattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find keys with pattern %s: %w", pattern, err)
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Close closes the KeyDB connection
func (c *KeyDBCache) Close() error {
	return c.client.Close()
}

// NoOpCache implementations (all operations are no-ops)

func (c *NoOpCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (c *NoOpCache) Get(ctx context.Context, key string, dest interface{}) error {
	return ErrCacheMiss
}

func (c *NoOpCache) Del(ctx context.Context, keys ...string) error {
	return nil
}

func (c *NoOpCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return 0, nil
}

func (c *NoOpCache) InvalidatePattern(ctx context.Context, pattern string) error {
	return nil
}

func (c *NoOpCache) Close() error {
	return nil
}

// Cache errors
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)

// Cache key constants and utilities
const (
	CertificatePrefix = "cert:"
	CAInfoKey         = "ca:info"
	CertListKey       = "cert:list"
	AuthPrefix        = "auth:"
	SettingsPrefix    = "settings:"
)

// CacheKey generates a cache key with prefix
func CacheKey(prefix, key string) string {
	return prefix + key
}

// CertificateCacheKey generates a cache key for certificates
func CertificateCacheKey(name string) string {
	return CacheKey(CertificatePrefix, name)
}

// AuthCacheKey generates a cache key for authentication data
func AuthCacheKey(key string) string {
	return CacheKey(AuthPrefix, key)
}

// SettingsCacheKey generates a cache key for settings
func SettingsCacheKey(key string) string {
	return CacheKey(SettingsPrefix, key)
}
