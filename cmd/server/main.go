package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HarshvardhanPandey2003/FluxEngine/internal/handler"
	"github.com/HarshvardhanPandey2003/FluxEngine/internal/models"
	"github.com/HarshvardhanPandey2003/FluxEngine/internal/worker"
)

func main() {
	// Phase 1: Single unbuffered channel for job communication and we use 
	// This is the main entry point(Starting point for accepting the requests) for the FluxEngine server application.
	jobQueue := make(chan *models.Job)

	// Start a single worker goroutine
	// This worker will perform CPU-intensive operations (bcrypt, statistics)
	worker := worker.NewWorker(1, jobQueue) // Worker ID = 1
	go worker.Start()

	// Initialize HTTP handler with access to the job queue
	jobHandler := handler.NewJobHandler(jobQueue)

	// Route registration
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/submit", jobHandler.HandleSubmitJob)

	// Configure HTTP server settings
	// We use & here instead of hhtp.Server directly  
	server := &http.Server{
		Addr:         ":8080",
		Handler:      nil, // Use DefaultServeMux
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background goroutine
	go func() {
		log.Println("🚀 FluxEngine Phase 1: CPU-Intensive Job Processor")
		log.Println("📌 Server running on http://localhost:8080")
		log.Println("📌 Endpoints:")
		log.Println("   GET  /health           - Health check")
		log.Println("   POST /submit           - Submit CPU-intensive job")
		log.Println("")
		log.Println("📊 Worker initialized (1 worker, unbuffered queue)")
		log.Println("⚙️  Ready to process password hashing & report generation")
		log.Println("")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	// Graceful shutdown: Wait for SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutdown signal received...")
	close(jobQueue) // Signal worker to stop after finishing current job
	time.Sleep(2 * time.Second) // Give worker time to finish (crude version)
	log.Println("✅ FluxEngine stopped")
}

// healthCheckHandler responds with server status.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":  "healthy",
		"service": "fluxengine",
		"phase":   "1-skeleton",
		"workers": 1,
		"time":    time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}
