package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/localca-go/pkg/certificates"
	"github.com/yourusername/localca-go/pkg/config"
	"github.com/yourusername/localca-go/pkg/handlers"
	"github.com/yourusername/localca-go/pkg/storage"

	"github.com/gin-gonic/gin"
)

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

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
	}()

	// Start HTTPS server if TLS is enabled
	if cfg.TLSEnabled {
		// Make sure we have a CA cert and key
		caCertPath := store.GetCAPublicKeyPath()
		caKeyPath := store.GetCAPrivateKeyPath()
		
		// Create a self-signed certificate for the service itself
		serviceCert := "service.crt"
		serviceKey := "service.key"
		
		if _, err := os.Stat(serviceCert); os.IsNotExist(err) {
			log.Println("Creating service certificate for HTTPS...")
			if err := certSvc.CreateServiceCertificate(); err != nil {
				log.Printf("Warning: Failed to create service certificate: %v. HTTPS will not be available.", err)
			} else {
				// Start HTTPS server
				go func() {
					httpsServer := &http.Server{
						Addr:    ":8443",
						Handler: router,
					}
					
					log.Println("HTTPS server starting on port 8443...")
					if err := httpsServer.ListenAndServeTLS(serviceCert, serviceKey); err != nil && err != http.ErrServerClosed {
						log.Printf("HTTPS server error: %v", err)
					}
				}()
			}
		}
	}

	// Start HTTP server
	log.Println("HTTP server starting on port 8080...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}