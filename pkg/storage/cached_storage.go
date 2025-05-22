package storage

import (
	"context"
	"crypto/x509"
	"log"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/cache"
)

// CachedStorage wraps the regular storage with caching functionality
type CachedStorage struct {
	*Storage // Embedded storage for fallback
	cache    cache.Cache
	ctx      context.Context
}

// CertificateInfo represents cached certificate information
type CertificateInfo struct {
	Name         string            `json:"name"`
	Certificate  *x509.Certificate `json:"certificate"`
	ExpiryDate   time.Time         `json:"expiry_date"`
	SerialNumber string            `json:"serial_number"`
}

// CAInfoCache represents cached CA information
type CAInfoCache struct {
	CAName       string    `json:"ca_name"`
	CAKey        string    `json:"ca_key"`
	Organization string    `json:"organization"`
	Country      string    `json:"country"`
	CachedAt     time.Time `json:"cached_at"`
}

// EmailSettingsCache represents cached email settings
type EmailSettingsCache struct {
	Server      string    `json:"server"`
	Port        string    `json:"port"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	UseTLS      bool      `json:"use_tls"`
	UseStartTLS bool      `json:"use_start_tls"`
	CachedAt    time.Time `json:"cached_at"`
}

// NewCachedStorage creates a new cached storage instance
func NewCachedStorage(storage *Storage, cacheInstance cache.Cache) *CachedStorage {
	return &CachedStorage{
		Storage: storage,
		cache:   cacheInstance,
		ctx:     context.Background(),
	}
}

// ListCertificates returns cached certificate list or fetches from storage
func (cs *CachedStorage) ListCertificates() ([]string, error) {
	// Try to get from cache first
	var certList []string
	err := cs.cache.Get(cs.ctx, cache.CertListKey, &certList)
	if err == nil {
		log.Printf("Cache hit for certificate list (%d certificates)", len(certList))
		return certList, nil
	}

	// Cache miss, fetch from storage
	log.Println("Cache miss for certificate list, fetching from storage")
	certList, err = cs.Storage.ListCertificates()
	if err != nil {
		return nil, err
	}

	// Cache the result
	if cacheErr := cs.cache.Set(cs.ctx, cache.CertListKey, certList, 5*time.Minute); cacheErr != nil {
		log.Printf("Failed to cache certificate list: %v", cacheErr)
	}

	return certList, nil
}

// GetCAInfo returns cached CA info or fetches from storage
func (cs *CachedStorage) GetCAInfo() (string, string, string, string, error) {
	// Try to get from cache first
	var caInfo CAInfoCache
	err := cs.cache.Get(cs.ctx, cache.CAInfoKey, &caInfo)
	if err == nil {
		log.Println("Cache hit for CA info")
		return caInfo.CAName, caInfo.CAKey, caInfo.Organization, caInfo.Country, nil
	}

	// Cache miss, fetch from storage
	log.Println("Cache miss for CA info, fetching from storage")
	caName, caKey, organization, country, err := cs.Storage.GetCAInfo()
	if err != nil {
		return "", "", "", "", err
	}

	// Cache the result
	caInfo = CAInfoCache{
		CAName:       caName,
		CAKey:        caKey,
		Organization: organization,
		Country:      country,
		CachedAt:     time.Now(),
	}

	if cacheErr := cs.cache.Set(cs.ctx, cache.CAInfoKey, caInfo, 1*time.Hour); cacheErr != nil {
		log.Printf("Failed to cache CA info: %v", cacheErr)
	}

	return caName, caKey, organization, country, nil
}

// SaveCAInfo saves CA info and invalidates related cache
func (cs *CachedStorage) SaveCAInfo(caName, caKey, organization, country string) error {
	// Save to storage first
	err := cs.Storage.SaveCAInfo(caName, caKey, organization, country)
	if err != nil {
		return err
	}

	// Invalidate CA info cache
	if cacheErr := cs.cache.Del(cs.ctx, cache.CAInfoKey); cacheErr != nil {
		log.Printf("Failed to invalidate CA info cache: %v", cacheErr)
	}

	log.Println("CA info saved and cache invalidated")
	return nil
}

// GetEmailSettings returns cached email settings or fetches from storage
func (cs *CachedStorage) GetEmailSettings() (string, string, string, string, string, string, bool, bool, error) {
	// Try to get from cache first
	cacheKey := cache.SettingsCacheKey("email")
	var emailSettings EmailSettingsCache
	err := cs.cache.Get(cs.ctx, cacheKey, &emailSettings)
	if err == nil {
		log.Println("Cache hit for email settings")
		return emailSettings.Server, emailSettings.Port, emailSettings.Username,
			emailSettings.Password, emailSettings.From, emailSettings.To,
			emailSettings.UseTLS, emailSettings.UseStartTLS, nil
	}

	// Cache miss, fetch from storage
	log.Println("Cache miss for email settings, fetching from storage")
	server, port, username, password, from, to, useTLS, useStartTLS, err := cs.Storage.GetEmailSettings()
	if err != nil {
		return "", "", "", "", "", "", false, false, err
	}

	// Cache the result
	emailSettings = EmailSettingsCache{
		Server:      server,
		Port:        port,
		Username:    username,
		Password:    password,
		From:        from,
		To:          to,
		UseTLS:      useTLS,
		UseStartTLS: useStartTLS,
		CachedAt:    time.Now(),
	}

	if cacheErr := cs.cache.Set(cs.ctx, cacheKey, emailSettings, 30*time.Minute); cacheErr != nil {
		log.Printf("Failed to cache email settings: %v", cacheErr)
	}

	return server, port, username, password, from, to, useTLS, useStartTLS, nil
}

// SaveEmailSettings saves email settings and invalidates cache
func (cs *CachedStorage) SaveEmailSettings(server, port, username, password, from, to string, useTLS, useStartTLS bool) error {
	// Save to storage first
	err := cs.Storage.SaveEmailSettings(server, port, username, password, from, to, useTLS, useStartTLS)
	if err != nil {
		return err
	}

	// Invalidate email settings cache
	cacheKey := cache.SettingsCacheKey("email")
	if cacheErr := cs.cache.Del(cs.ctx, cacheKey); cacheErr != nil {
		log.Printf("Failed to invalidate email settings cache: %v", cacheErr)
	}

	log.Println("Email settings saved and cache invalidated")
	return nil
}

// DeleteCertificate deletes certificate and invalidates related cache
func (cs *CachedStorage) DeleteCertificate(name string) error {
	// Delete from storage first
	err := cs.Storage.DeleteCertificate(name)
	if err != nil {
		return err
	}

	// Invalidate certificate-related caches
	certKey := cache.CertificateCacheKey(name)
	if cacheErr := cs.cache.Del(cs.ctx, certKey, cache.CertListKey); cacheErr != nil {
		log.Printf("Failed to invalidate certificate cache: %v", cacheErr)
	}

	log.Printf("Certificate %s deleted and cache invalidated", name)
	return nil
}

// GetCertificateNameBySerial checks cache first, then falls back to storage
func (cs *CachedStorage) GetCertificateNameBySerial(serialNumber string) (string, error) {
	// Try to get from cache first
	cacheKey := cache.CacheKey("serial:", serialNumber)
	var certName string
	err := cs.cache.Get(cs.ctx, cacheKey, &certName)
	if err == nil {
		log.Printf("Cache hit for serial number %s", serialNumber)
		return certName, nil
	}

	// Cache miss, fetch from storage
	log.Printf("Cache miss for serial number %s, fetching from storage", serialNumber)
	certName, err = cs.Storage.GetCertificateNameBySerial(serialNumber)
	if err != nil {
		return "", err
	}

	// Cache the result for 2 hours (certificates don't change serial mapping often)
	if cacheErr := cs.cache.Set(cs.ctx, cacheKey, certName, 2*time.Hour); cacheErr != nil {
		log.Printf("Failed to cache serial mapping: %v", cacheErr)
	}

	return certName, nil
}

// SaveCertificateSerialMapping saves serial mapping and updates cache
func (cs *CachedStorage) SaveCertificateSerialMapping(serialNumber, certName string) error {
	// Save to storage first
	err := cs.Storage.SaveCertificateSerialMapping(serialNumber, certName)
	if err != nil {
		return err
	}

	// Update cache
	cacheKey := cache.CacheKey("serial:", serialNumber)
	if cacheErr := cs.cache.Set(cs.ctx, cacheKey, certName, 2*time.Hour); cacheErr != nil {
		log.Printf("Failed to cache serial mapping: %v", cacheErr)
	}

	// Invalidate certificate list cache since we have a new certificate
	if cacheErr := cs.cache.Del(cs.ctx, cache.CertListKey); cacheErr != nil {
		log.Printf("Failed to invalidate certificate list cache: %v", cacheErr)
	}

	log.Printf("Serial mapping saved and cached for certificate %s", certName)
	return nil
}

// InvalidateAllCaches removes all cached data
func (cs *CachedStorage) InvalidateAllCaches() error {
	patterns := []string{
		cache.CertificatePrefix + "*",
		cache.AuthPrefix + "*",
		cache.SettingsPrefix + "*",
		cache.CAInfoKey,
		cache.CertListKey,
	}

	for _, pattern := range patterns {
		if err := cs.cache.InvalidatePattern(cs.ctx, pattern); err != nil {
			log.Printf("Failed to invalidate cache pattern %s: %v", pattern, err)
		}
	}

	log.Println("All caches invalidated")
	return nil
}

// WarmUpCache pre-loads frequently accessed data into cache
func (cs *CachedStorage) WarmUpCache() error {
	log.Println("Warming up cache...")

	// Warm up certificate list
	if _, err := cs.ListCertificates(); err != nil {
		log.Printf("Failed to warm up certificate list: %v", err)
	}

	// Warm up CA info
	if _, _, _, _, err := cs.GetCAInfo(); err != nil {
		log.Printf("Failed to warm up CA info: %v", err)
	}

	// Warm up email settings (if they exist)
	if _, _, _, _, _, _, _, _, err := cs.GetEmailSettings(); err != nil {
		// This is expected if email settings haven't been configured yet
		log.Printf("Email settings not cached (likely not configured): %v", err)
	}

	log.Println("Cache warm-up completed")
	return nil
}

// CacheStats represents cache statistics
type CacheStats struct {
	CertificateCount int       `json:"certificate_count"`
	CAInfoCached     bool      `json:"ca_info_cached"`
	EmailCached      bool      `json:"email_cached"`
	LastUpdate       time.Time `json:"last_update"`
}

// GetCacheStats returns information about what's currently cached
func (cs *CachedStorage) GetCacheStats() (*CacheStats, error) {
	stats := &CacheStats{
		LastUpdate: time.Now(),
	}

	// Check if CA info is cached
	exists, err := cs.cache.Exists(cs.ctx, cache.CAInfoKey)
	if err == nil {
		stats.CAInfoCached = exists > 0
	}

	// Check if email settings are cached
	emailKey := cache.SettingsCacheKey("email")
	exists, err = cs.cache.Exists(cs.ctx, emailKey)
	if err == nil {
		stats.EmailCached = exists > 0
	}

	// Check certificate list
	exists, err = cs.cache.Exists(cs.ctx, cache.CertListKey)
	if err == nil && exists > 0 {
		var certList []string
		if cacheErr := cs.cache.Get(cs.ctx, cache.CertListKey, &certList); cacheErr == nil {
			stats.CertificateCount = len(certList)
		}
	}

	return stats, nil
}
