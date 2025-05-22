package handlers

import (
	"context"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/cache"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
)

// StorageManager manages both cached and non-cached storage operations
type StorageManager struct {
	baseStore   *storage.Storage
	cachedStore *storage.CachedStorage
	cache       cache.Cache
}

// NewStorageManager creates a new storage manager
func NewStorageManager(baseStore *storage.Storage, cachedStore *storage.CachedStorage, cacheInstance cache.Cache) *StorageManager {
	return &StorageManager{
		baseStore:   baseStore,
		cachedStore: cachedStore,
		cache:       cacheInstance,
	}
}

// ListCertificates returns certificates using cache if available
func (sm *StorageManager) ListCertificates() ([]string, error) {
	if sm.cachedStore != nil {
		return sm.cachedStore.ListCertificates()
	}
	return sm.baseStore.ListCertificates()
}

// GetCAInfo returns CA info using cache if available
func (sm *StorageManager) GetCAInfo() (string, string, string, string, error) {
	if sm.cachedStore != nil {
		return sm.cachedStore.GetCAInfo()
	}
	return sm.baseStore.GetCAInfo()
}

// SaveCAInfo saves CA info and invalidates cache if available
func (sm *StorageManager) SaveCAInfo(caName, caKey, organization, country string) error {
	if sm.cachedStore != nil {
		return sm.cachedStore.SaveCAInfo(caName, caKey, organization, country)
	}
	return sm.baseStore.SaveCAInfo(caName, caKey, organization, country)
}

// GetEmailSettings returns email settings using cache if available
func (sm *StorageManager) GetEmailSettings() (string, string, string, string, string, string, bool, bool, error) {
	if sm.cachedStore != nil {
		return sm.cachedStore.GetEmailSettings()
	}
	return sm.baseStore.GetEmailSettings()
}

// SaveEmailSettings saves email settings and invalidates cache if available
func (sm *StorageManager) SaveEmailSettings(server, port, username, password, from, to string, useTLS, useStartTLS bool) error {
	if sm.cachedStore != nil {
		return sm.cachedStore.SaveEmailSettings(server, port, username, password, from, to, useTLS, useStartTLS)
	}
	return sm.baseStore.SaveEmailSettings(server, port, username, password, from, to, useTLS, useStartTLS)
}

// DeleteCertificate deletes certificate and invalidates cache if available
func (sm *StorageManager) DeleteCertificate(name string) error {
	if sm.cachedStore != nil {
		return sm.cachedStore.DeleteCertificate(name)
	}
	return sm.baseStore.DeleteCertificate(name)
}

// GetCertificateNameBySerial gets certificate name by serial using cache if available
func (sm *StorageManager) GetCertificateNameBySerial(serialNumber string) (string, error) {
	if sm.cachedStore != nil {
		return sm.cachedStore.GetCertificateNameBySerial(serialNumber)
	}
	return sm.baseStore.GetCertificateNameBySerial(serialNumber)
}

// SaveCertificateSerialMapping saves serial mapping and updates cache if available
func (sm *StorageManager) SaveCertificateSerialMapping(serialNumber, certName string) error {
	if sm.cachedStore != nil {
		return sm.cachedStore.SaveCertificateSerialMapping(serialNumber, certName)
	}
	return sm.baseStore.SaveCertificateSerialMapping(serialNumber, certName)
}

// GetBaseStore returns the base storage for operations that need it
func (sm *StorageManager) GetBaseStore() *storage.Storage {
	return sm.baseStore
}

// InvalidateAllCaches invalidates all caches if caching is enabled
func (sm *StorageManager) InvalidateAllCaches() error {
	if sm.cachedStore != nil {
		return sm.cachedStore.InvalidateAllCaches()
	}
	return nil
}

// InvalidateCache invalidates specific cache keys
func (sm *StorageManager) InvalidateCache(keys ...string) error {
	if sm.cache != nil {
		ctx := context.Background()
		return sm.cache.Del(ctx, keys...)
	}
	return nil
}

// GetCacheStats returns cache statistics if caching is enabled
func (sm *StorageManager) GetCacheStats() (*storage.CacheStats, error) {
	if sm.cachedStore != nil {
		return sm.cachedStore.GetCacheStats()
	}
	return nil, nil
}

// CacheAuthToken caches an authentication token
func (sm *StorageManager) CacheAuthToken(token, username string, ttl int) error {
	if sm.cache != nil {
		ctx := context.Background()
		cacheKey := cache.AuthCacheKey(token)
		cacheTTL := time.Duration(ttl) * time.Second
		return sm.cache.Set(ctx, cacheKey, username, cacheTTL)
	}
	return nil
}

// GetCachedAuthToken retrieves a cached authentication token
func (sm *StorageManager) GetCachedAuthToken(token string) (string, error) {
	if sm.cache != nil {
		ctx := context.Background()
		cacheKey := cache.AuthCacheKey(token)
		var username string
		err := sm.cache.Get(ctx, cacheKey, &username)
		if err != nil {
			return "", err
		}
		return username, nil
	}
	return "", cache.ErrCacheMiss
}
