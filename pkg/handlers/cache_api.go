package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// apiCacheStatsHandler returns cache statistics
func apiCacheStatsHandler(storageManager *StorageManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := storageManager.GetCacheStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to get cache statistics",
			})
			return
		}

		if stats == nil {
			c.JSON(http.StatusOK, APIResponse{
				Success: true,
				Message: "Caching is disabled",
				Data: map[string]interface{}{
					"cache_enabled": false,
				},
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Cache statistics retrieved successfully",
			Data: map[string]interface{}{
				"cache_enabled":     true,
				"certificate_count": stats.CertificateCount,
				"ca_info_cached":    stats.CAInfoCached,
				"email_cached":      stats.EmailCached,
				"last_update":       stats.LastUpdate,
			},
		})
	}
}

// apiCacheClearHandler clears all cache data
func apiCacheClearHandler(storageManager *StorageManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := storageManager.InvalidateAllCaches()
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to clear cache",
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Cache cleared successfully",
		})
	}
}

// apiCacheInvalidateHandler invalidates specific cache keys
func apiCacheInvalidateHandler(storageManager *StorageManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Keys []string `json:"keys"`
		}

		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid request format",
			})
			return
		}

		if len(request.Keys) == 0 {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "No cache keys provided",
			})
			return
		}

		err := storageManager.InvalidateCache(request.Keys...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to invalidate cache keys",
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Cache keys invalidated successfully",
		})
	}
}
