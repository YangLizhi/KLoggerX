package main

import (
	"fmt"
	"log"

	"kloggerx-server/api/v1"
	"kloggerx-server/config"
	"kloggerx-server/internal/middleware"
	"kloggerx-server/internal/pkg/logger"
	miniosvc "kloggerx-server/internal/pkg/minio"
	"kloggerx-server/internal/repository/mysql"
	"kloggerx-server/internal/repository/qdrant"
	"kloggerx-server/internal/repository/redis"
	"kloggerx-server/internal/service"
	"kloggerx-server/internal/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Init("config/config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger.Init(config.Cfg.Log)

	if err := mysql.Init(config.Cfg.Database); err != nil {
		log.Fatalf("Failed to connect MySQL: %v", err)
	}

	if err := redis.Init(config.Cfg.Redis); err != nil {
		log.Fatalf("Failed to connect Redis: %v", err)
	}

	// Initialize MinIO (optional - will log warning if failed)
	if err := miniosvc.Init(config.Cfg.MinIO); err != nil {
		log.Printf("Warning: Failed to connect MinIO: %v (file uploads may fail)", err)
	}

	// Initialize Qdrant (optional - will log warning if failed)
	if config.Cfg.Qdrant.Host != "" {
		if err := qdrant.Init(config.Cfg.Qdrant); err != nil {
			log.Printf("Warning: Failed to connect Qdrant: %v (vector search disabled)", err)
		} else {
			qdrant.InitDefaultVectorClient(config.Cfg.Qdrant.Collection)
		}
	} else {
		log.Println("Warning: Qdrant not configured, vector search disabled")
	}

	// Initialize embedding and retrieval services (optional - requires AI settings)
	if qdrant.DefaultVectorClient != nil {
		// Get AI settings for embedding service
		embedBaseURL, embedAPIKey, embedModel, err := service.GetEmbeddingModelSettings()
		if err == nil && embedAPIKey != "" {
			// Initialize embedding service with the configured embedding model
			service.InitEmbeddingService(embedBaseURL, embedAPIKey, embedModel)
			// Set vector dimension based on the embedding model
			vectorDim := service.GetEmbeddingDimensionForModel(embedModel)
			qdrant.SetVectorSize(vectorDim)
			// Initialize retrieval service
			service.InitRetrievalService(service.GetEmbeddingService(), qdrant.DefaultVectorClient)
			// Initialize RAPTOR service
			service.InitRaptorService(service.GetEmbeddingService(), qdrant.DefaultVectorClient)
			log.Printf("Vector search initialized with embedding model: %s (dimension: %d)", embedModel, vectorDim)
		} else {
			log.Printf("Warning: Embedding model settings not found, vector embeddings disabled: %v", err)
		}
	}

	mysql.AutoMigrate()

	// Start storage policy background jobs
	service.StartStoragePolicyJobs()

	hub := ws.NewHub()
	go hub.Run()

	gin.SetMode(config.Cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())

	// Serve uploaded files statically
	r.Static("/uploads", "./uploads")

	v1.RegisterRoutes(r, hub)

	addr := fmt.Sprintf(":%d", config.Cfg.Server.Port)
	logger.Info("Server starting on " + addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
