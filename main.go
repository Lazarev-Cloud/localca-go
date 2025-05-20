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
	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/handlers"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"

	"github.com/gin-gonic/gin"
)

func getSecureTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
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
			tls.CurveP256,
			tls.X25519,
		},
	}
}

func main() {
	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	store, err := storage.NewStorage(cfg.StoragePath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize certificate service
	certSvc, err := certificates.NewCertificateService(cfg, store)
	if err != nil {
		log.Fatalf("Failed to initialize certificate service: %v", err)
	}

	// Check if CA exists, create if it doesn't
	exists, err := certSvc.CAExists()
	if err != nil {
		log.Fatalf("Failed to check CA existence: %v", err)
	}

	if !exists {
		log.Println("Creating new CA certificate...")
		if err := certSvc.CreateCA(); err != nil {
			log.Fatalf("Failed to create CA: %v", err)
		}
		log.Println("CA certificate created successfully")
	} else {
		log.Println("Using existing CA certificate")
	}

	// Initialize router
	router := gin.Default()

	// Configure static file serving
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	// Setup routes
	handlers.SetupRoutes(router, certSvc, store, cfg)

	// Configure server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
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
		if err := acme.StartACMEServer(ctx, certSvc, store, ":8555", getSecureTLSConfig()); err != nil {
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
				}
				
				log.Println("HTTPS server starting on port 8443...")
				if err := httpsServer.ListenAndServeTLS(serviceCert, serviceKey); err != nil && err != http.ErrServerClosed {
					log.Printf("HTTPS server error: %v", err)
				}
			}()
		}
	}

	// Start HTTP server
	log.Println("HTTP server starting on port 8080...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}