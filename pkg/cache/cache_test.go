package cache

import (
	"context"
	"testing"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestNewCache_NoOp(t *testing.T) {
	cfg := &config.Config{
		CacheEnabled: false,
	}

	cache, err := NewCache(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cache)
	assert.IsType(t, &NoOpCache{}, cache)
}

func TestNoOpCache_Set(t *testing.T) {
	cache := &NoOpCache{}
	ctx := context.Background()

	// NoOp cache should not return error on set
	err := cache.Set(ctx, "test-key", "test-value", time.Minute)
	assert.NoError(t, err)
}

func TestNoOpCache_Get(t *testing.T) {
	cache := &NoOpCache{}
	ctx := context.Background()

	// NoOp cache should always return cache miss
	var value string
	err := cache.Get(ctx, "test-key", &value)
	assert.Error(t, err)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestNoOpCache_Del(t *testing.T) {
	cache := &NoOpCache{}
	ctx := context.Background()

	// NoOp cache should not return error on delete
	err := cache.Del(ctx, "test-key")
	assert.NoError(t, err)
}

func TestNoOpCache_Exists(t *testing.T) {
	cache := &NoOpCache{}
	ctx := context.Background()

	// NoOp cache should always return 0 (no keys exist)
	count, err := cache.Exists(ctx, "test-key")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestNoOpCache_InvalidatePattern(t *testing.T) {
	cache := &NoOpCache{}
	ctx := context.Background()

	// NoOp cache should not return error on invalidate
	err := cache.InvalidatePattern(ctx, "test:*")
	assert.NoError(t, err)
}

func TestNoOpCache_Close(t *testing.T) {
	cache := &NoOpCache{}

	// NoOp cache should not return error on close
	err := cache.Close()
	assert.NoError(t, err)
}

func TestCacheKey(t *testing.T) {
	tests := []struct {
		prefix   string
		key      string
		expected string
	}{
		{"test:", "key1", "test:key1"},
		{"", "key2", "key2"},
		{"prefix", "", "prefix"},
		{"", "", ""},
	}

	for _, tt := range tests {
		result := CacheKey(tt.prefix, tt.key)
		assert.Equal(t, tt.expected, result)
	}
}

func TestCertificateCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"example.com", "cert:example.com"},
		{"", "cert:"},
		{"test-cert", "cert:test-cert"},
	}

	for _, tt := range tests {
		result := CertificateCacheKey(tt.name)
		assert.Equal(t, tt.expected, result)
	}
}

func TestAuthCacheKey(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"session:123", "auth:session:123"},
		{"", "auth:"},
		{"token", "auth:token"},
	}

	for _, tt := range tests {
		result := AuthCacheKey(tt.key)
		assert.Equal(t, tt.expected, result)
	}
}

func TestSettingsCacheKey(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"smtp", "settings:smtp"},
		{"", "settings:"},
		{"email.enabled", "settings:email.enabled"},
	}

	for _, tt := range tests {
		result := SettingsCacheKey(tt.key)
		assert.Equal(t, tt.expected, result)
	}
}

func TestErrCacheMiss(t *testing.T) {
	assert.NotNil(t, ErrCacheMiss)
	assert.Contains(t, ErrCacheMiss.Error(), "cache miss")
}

func TestCacheConstants(t *testing.T) {
	assert.Equal(t, "cert:", CertificatePrefix)
	assert.Equal(t, "ca:info", CAInfoKey)
	assert.Equal(t, "cert:list", CertListKey)
	assert.Equal(t, "auth:", AuthPrefix)
	assert.Equal(t, "settings:", SettingsPrefix)
}

// Test that the Cache interface is properly implemented
func TestCacheInterface(t *testing.T) {
	var _ Cache = &NoOpCache{}
	var _ Cache = &KeyDBCache{}
}

func TestKeyDBCache_InvalidConfig(t *testing.T) {
	cfg := &config.Config{
		CacheEnabled:  true,
		KeyDBHost:     "invalid-host",
		KeyDBPort:     6379,
		KeyDBPassword: "",
		KeyDBDB:       0,
		CacheTTL:      300,
	}

	// This should fail to connect to invalid host
	_, err := NewCache(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to KeyDB")
}
