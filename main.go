package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/acme"
	"github.com/Lazarev-Cloud/localca-go/pkg/cache"
	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/handlers"
	"github.com/Lazarev-Cloud/localca-go/pkg/logging"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"

	"github.com/gin-gonic/gin"
)

func getSecureTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
		SessionTicketsDisabled: true,
		Renegotiation:          tls.RenegotiateNever,
	}
}

func main() {
	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize structured logger
	logger, err := logging.NewLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info("Starting LocalCA server with enhanced storage and logging")

	// Initialize cache
	cacheInstance, err := cache.NewCache(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize cache")
	}
	defer func() {
		if err := cacheInstance.Close(); err != nil {
			logger.WithError(err).Error("Failed to close cache")
		}
	}()

	// Initialize enhanced storage (with database and S3 support)
	enhancedStore, err := storage.NewEnhancedStorage(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize enhanced storage")
	}
	defer func() {
		if err := enhancedStore.Close(); err != nil {
			logger.WithError(err).Error("Failed to close enhanced storage")
		}
	}()

	// Log storage health status
	health := enhancedStore.Health()
	for backend, err := range health {
		if err != nil {
			logger.WithField("backend", backend).WithError(err).Warn("Storage backend not available")
		} else {
			logger.WithField("backend", backend).Info("Storage backend healthy")
		}
	}

	// Initialize file storage for backward compatibility
	baseStore, err := storage.NewStorage(cfg.DataDir)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize file storage")
	}

	// Initialize storage wrapper - either cached or regular
	var store storage.StorageInterface
	var cachedStore *storage.CachedStorage

	if cfg.CacheEnabled {
		cachedStore = storage.NewCachedStorage(baseStore, cacheInstance)
		store = cachedStore

		// Warm up the cache with frequently accessed data
		go func() {
			if err := cachedStore.WarmUpCache(); err != nil {
				logger.WithError(err).Error("Failed to warm up cache")
			}
		}()

		logger.Info("Cache-enabled storage initialized")
	} else {
		store = baseStore
		logger.Info("Cache-disabled storage initialized")
	}

	// Initialize certificate service with enhanced storage
	certSvc, err := certificates.NewCertificateService(cfg, enhancedStore)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize certificate service")
	}

	// Check if CA exists, create if it doesn't
	exists, err := certSvc.CAExists()
	if err != nil {
		logger.WithError(err).Fatal("Failed to check CA existence")
	}

	if !exists {
		logger.Info("Creating new CA certificate...")
		if err := certSvc.CreateCA(); err != nil {
			logger.WithError(err).Fatal("Failed to create CA")
		}
		logger.Info("CA certificate created successfully")

		// Log CA creation audit event
		enhancedStore.LogAudit("create", "ca", cfg.CAName, "system", "localca-server", "CA created during startup", true, "")
	} else {
		logger.Info("Using existing CA certificate")
	}

	// Load auth config and log setup token if setup is not completed
	authConfig, err := handlers.LoadAuthConfig(baseStore)
	if err != nil {
		log.Printf("Failed to load auth config: %v", err)
	} else if !authConfig.SetupCompleted {
		log.Println("==========================================================")
		log.Println("INITIAL SETUP REQUIRED")
		log.Println("Please visit /setup to complete the initial configuration")
		log.Println("Setup Token:", authConfig.SetupToken)
		log.Println("This token will expire in 24 hours")
		log.Println("==========================================================")
	}

	// Initialize router
	router := gin.Default()

	// Configure static file serving
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	// Setup routes - pass cachedStore if caching is enabled, otherwise baseStore
	if cfg.CacheEnabled && cachedStore != nil {
		handlers.SetupRoutes(router, certSvc, baseStore, cfg)
	} else {
		handlers.SetupRoutes(router, certSvc, baseStore, cfg)
	}

	// Configure server
	server := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: router,
		// Add timeouts to prevent slow client attacks
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Println("Shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		cancel() // Cancel the context for ACME server

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
	}()

	// Start ACME server
	go func() {
		log.Println("Starting ACME server on port 8555...")
		if err := acme.StartACMEServer(ctx, certSvc, baseStore, ":8555", getSecureTLSConfig()); err != nil {
			if err != http.ErrServerClosed {
				log.Printf("ACME server error: %v", err)
			}
		}
	}()

	// Start HTTPS server if TLS is enabled
	if cfg.TLSEnabled {
		// Make sure we have a CA cert and key
		caCertPath := store.GetCAPublicKeyPath()
		caKeyPath := store.GetCAPrivateKeyPath()

		// Certificate paths for the service
		serviceCert := filepath.Join(store.GetBasePath(), "service.crt")
		serviceKey := filepath.Join(store.GetBasePath(), "service.key")

		// Check if service certificate exists
		if _, err := os.Stat(serviceCert); os.IsNotExist(err) {
			log.Println("Creating service certificate for HTTPS...")
			// Use caCertPath and caKeyPath here
			if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
				log.Fatalf("CA certificate not found at %s", caCertPath)
			}
			if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
				log.Fatalf("CA private key not found at %s", caKeyPath)
			}
			if err := certSvc.CreateServiceCertificate(); err != nil {
				log.Printf("Warning: Failed to create service certificate: %v. HTTPS will not be available.", err)
			} else {
				// Start HTTPS server
				go func() {
					httpsServer := &http.Server{
						Addr:      ":8443",
						Handler:   router,
						TLSConfig: getSecureTLSConfig(),
						// Add timeouts to prevent slow client attacks
						ReadTimeout:  10 * time.Second,
						WriteTimeout: 30 * time.Second,
						IdleTimeout:  120 * time.Second,
					}

					log.Println("HTTPS server starting on port 8443...")
					if err := httpsServer.ListenAndServeTLS(serviceCert, serviceKey); err != nil && err != http.ErrServerClosed {
						log.Printf("HTTPS server error: %v", err)
					}
				}()
			}
		} else {
			// Start HTTPS server with existing certificate
			go func() {
				httpsServer := &http.Server{
					Addr:      ":8443",
					Handler:   router,
					TLSConfig: getSecureTLSConfig(),
					// Add timeouts to prevent slow client attacks
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  120 * time.Second,
				}

				log.Println("HTTPS server starting on port 8443...")
				if err := httpsServer.ListenAndServeTLS(serviceCert, serviceKey); err != nil && err != http.ErrServerClosed {
					log.Printf("HTTPS server error: %v", err)
				}
			}()
		}
	}

	// Start HTTP server
	log.Printf("HTTP server starting on %s...", cfg.ListenAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
