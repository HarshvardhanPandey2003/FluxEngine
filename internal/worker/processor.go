package worker

import (
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"math/big"
	"sort"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/HarshvardhanPandey2003/FluxEngine/internal/models"
)

// Worker represents a background processor for CPU-intensive jobs.
// In Phase 1, we have a single worker. In Phase 3, we'll spawn multiple workers.
type Worker struct {
	id       int                   // Worker identifier (useful for logging)
	jobQueue <-chan *models.Job    // Receive-only channel
}

// NewWorker creates a worker instance.
func NewWorker(id int, jobQueue <-chan *models.Job) *Worker {
	return &Worker{
		id:       id,
		jobQueue: jobQueue,
	}
}

// Start begins the worker's processing loop.
// This function blocks until the job channel is closed.
func (w *Worker) Start() {
	log.Printf("🔧 Worker #%d started and waiting for CPU-intensive jobs...", w.id)

	// Range over the channel: blocks until a job arrives or channel closes
	for job := range w.jobQueue {
		w.processJob(job)
	}

	log.Printf("🛑 Worker #%d stopped (channel closed)", w.id)
}

// processJob routes the job to the appropriate CPU-intensive handler.
func (w *Worker) processJob(job *models.Job) {
	startTime := time.Now()
	log.Printf("⚙️  Worker #%d processing job %s (type: %s)", w.id, job.ID, job.Type)

	job.Status = "processing"

	var result interface{}
	var err error

	// Route to CPU-intensive workload based on job type
	switch job.Type {
	case "password_hash":
		result, err = w.processPasswordHash(job)
	case "report_generation":
		result, err = w.processReportGeneration(job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.Type)
	}

	duration := time.Since(startTime)

	if err != nil {
		job.Status = "failed"
		log.Printf("❌ Worker #%d failed job %s: %v (took %v)", w.id, job.ID, err, duration)
		return
	}

	job.Status = "completed"
	log.Printf("✅ Worker #%d completed job %s in %v", w.id, job.ID, duration)

	// Log result summary (in Phase 4, we'd persist this to a database)
	w.logResult(job.ID, result, duration)
}

// processPasswordHash performs secure password hashing using bcrypt.
// Bcrypt is intentionally slow (CPU-intensive) to defend against brute-force attacks.
func (w *Worker) processPasswordHash(job *models.Job) (interface{}, error) {
	password, ok := job.Payload["password"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid password field")
	}

	// Extract cost parameter (default to bcrypt.DefaultCost = 10)
	cost := bcrypt.DefaultCost
	if costVal, ok := job.Payload["cost"].(float64); ok {
		cost = int(costVal)
		// Bcrypt cost must be between 4 and 31
		if cost < 4 || cost > 31 {
			cost = bcrypt.DefaultCost
		}
	}

	log.Printf("   🔐 Hashing password with bcrypt (cost=%d)...", cost)

	// THIS IS CPU-INTENSIVE: bcrypt.GenerateFromPassword uses key stretching
	// Cost of 10 = 2^10 iterations, cost of 12 = 2^12 iterations, etc.
	// Higher cost = exponentially more CPU time
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt hashing failed: %w", err)
	}

	// Verify the hash (additional CPU work to demonstrate round-trip)
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return nil, fmt.Errorf("hash verification failed: %w", err)
	}

	return map[string]interface{}{
		"hash_length": len(hashedPassword),
		"cost":        cost,
		"algorithm":   "bcrypt",
		"verified":    true,
	}, nil
}

// processReportGeneration simulates large-scale data aggregation.
// Generates millions of random numbers and computes statistical measures.
// This is CPU-bound work that benefits from Go's lightweight goroutines.
func (w *Worker) processReportGeneration(job *models.Job) (interface{}, error) {
	// Extract data points count (default to 1 million)
	dataPoints := 1_000_000
	if dpVal, ok := job.Payload["data_points"].(float64); ok {
		dataPoints = int(dpVal)
	}

	// Cap at 10 million to prevent excessive memory usage in Phase 1
	if dataPoints > 10_000_000 {
		dataPoints = 10_000_000
	}

	log.Printf("   📊 Generating %d random values for statistical analysis...", dataPoints)

	// Generate random dataset
	// In production, this would be reading from a database or file
	dataset := make([]float64, dataPoints)
	for i := 0; i < dataPoints; i++ {
		// Generate cryptographically secure random float between 0 and 1000
		// This is more CPU-intensive than math/rand but demonstrates real work
		val, err := rand.Int(rand.Reader, big.NewInt(1000))
		if err != nil {
			// Fallback to timestamp-based pseudo-random if crypto/rand fails
			dataset[i] = float64(time.Now().UnixNano() % 1000)
		} else {
			dataset[i] = float64(val.Int64())
		}
	}

	log.Printf("   🧮 Computing statistics on %d values...", dataPoints)

	// Compute statistics (CPU-intensive operations)
	stats := computeStatistics(dataset)

	return stats, nil
}

// computeStatistics performs CPU-intensive statistical computations.
// These operations require iterating over large datasets multiple times.
func computeStatistics(data []float64) map[string]interface{} {
	n := len(data)
	if n == 0 {
		return map[string]interface{}{"error": "empty dataset"}
	}

	// 1. Sum and Mean (single pass)
	var sum float64
	for _, v := range data {
		sum += v
	}
	mean := sum / float64(n)

	// 2. Variance and Standard Deviation (requires mean first)
	var varianceSum float64
	for _, v := range data {
		diff := v - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(n)
	stdDev := math.Sqrt(variance)

	// 3. Min and Max (single pass)
	min, max := data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// 4. Median and Percentiles (requires sorting - O(n log n))
	// Sorting is CPU-intensive for large datasets
	sortedData := make([]float64, n)
	copy(sortedData, data)
	sort.Float64s(sortedData)

	median := sortedData[n/2]
	if n%2 == 0 {
		median = (sortedData[n/2-1] + sortedData[n/2]) / 2
	}

	p25 := sortedData[n/4]
	p75 := sortedData[3*n/4]
	p95 := sortedData[95*n/100]
	p99 := sortedData[99*n/100]

	return map[string]interface{}{
		"count":       n,
		"sum":         sum,
		"mean":        mean,
		"std_dev":     stdDev,
		"variance":    variance,
		"min":         min,
		"max":         max,
		"median":      median,
		"p25":         p25,
		"p75":         p75,
		"p95":         p95,
		"p99":         p99,
		"range":       max - min,
	}
}

// logResult displays the computation result in a readable format.
func (w *Worker) logResult(jobID string, result interface{}, duration time.Duration) {
	log.Printf("   📈 Job %s results (completed in %v):", jobID, duration)

	if resultMap, ok := result.(map[string]interface{}); ok {
		for key, value := range resultMap {
			log.Printf("      %s: %v", key, value)
		}
	}
}
