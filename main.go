package main

import (
	"go-microservice/handlers"
	"go-microservice/metrics"
	"go-microservice/services"
	"go-microservice/utils"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	userService := services.NewUserService()
	minioEndpoint := getEnv("MINIO_ENDPOINT", "minio:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	minioBucket := getEnv("MINIO_BUCKET", "users")
	minioUseSSL := strings.EqualFold(getEnv("MINIO_USE_SSL", "false"), "true")
	integrationService, err := services.NewIntegrationService(minioEndpoint, minioAccessKey, minioSecretKey, minioBucket, minioUseSSL)
	if err != nil {
		log.Fatalf("failed to initialize MinIO client: %v", err)
	}
	userHandler := handlers.NewUserHandler(userService, integrationService, minioBucket)
	integrationHandler := handlers.NewIntegrationHandler(integrationService)
	router := mux.NewRouter()
	metrics.RegisterMetricsHandler(router)
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	api.HandleFunc("/users/{id:[0-9]+}", userHandler.GetUserByID).Methods("GET")
	api.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	api.HandleFunc("/users/{id:[0-9]+}", userHandler.UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id:[0-9]+}", userHandler.DeleteUser).Methods("DELETE")
	api.HandleFunc("/storage/objects", integrationHandler.UploadObject).Methods("POST")
	api.HandleFunc("/storage/presign", integrationHandler.GetPresignedURL).Methods("POST")
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	rateLimiter := utils.NewFixedRateLimiter()
	handler := metrics.MetricsMiddleware(rateLimiter(router))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	log.Printf("API endpoints:")
	log.Printf("  GET    /api/users")
	log.Printf("  GET    /api/users/{id}")
	log.Printf("  POST   /api/users")
	log.Printf("  PUT    /api/users/{id}")
	log.Printf("  DELETE /api/users/{id}")
	log.Printf("  POST   /api/storage/objects")
	log.Printf("  POST   /api/storage/presign")
	log.Printf("  GET    /metrics")
	log.Printf("  GET    /health")
	server := &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Server configured with timeouts: Read=10s, Write=10s, Idle=120s")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return defaultValue
}