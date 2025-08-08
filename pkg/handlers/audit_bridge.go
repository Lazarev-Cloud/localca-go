package handlers

import "github.com/Lazarev-Cloud/localca-go/pkg/storage"

// enhancedAudit provides access to EnhancedStorage for unified audit logging
var enhancedAudit *storage.EnhancedStorage

// SetEnhancedStorage sets the global enhanced storage for audit logging
func SetEnhancedStorage(e *storage.EnhancedStorage) {
	enhancedAudit = e
}

// getEnhancedStorage returns the enhanced storage instance if available
func getEnhancedStorage() *storage.EnhancedStorage {
	return enhancedAudit
}
