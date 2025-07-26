package scraper

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// ConcurrencyConfig configures concurrent request handling
type ConcurrencyConfig struct {
	MaxConcurrent     int           `json:"max_concurrent"`     // Maximum concurrent requests
	WorkerPoolSize    int           `json:"worker_pool_size"`   // Number of worker goroutines
	RequestBuffer     int           `json:"request_buffer"`     // Size of request queue buffer
	RateLimitPerSec   int           `json:"rate_limit_per_sec"` // Max requests per second
	ConnectionTimeout time.Duration `json:"connection_timeout"` // Connection timeout
	IdleTimeout       time.Duration `json:"idle_timeout"`       // Idle connection timeout
	MaxIdleConns      int           `json:"max_idle_conns"`     // Max idle connections
	MaxConnsPerHost   int           `json:"max_conns_per_host"` // Max connections per host
}

// RequestJob represents a single scraping request
type RequestJob struct {
	ID         string                 `json:"id"`
	URL        string                 `json:"url"`
	Method     string                 `json:"method"`
	Headers    map[string]string      `json:"headers"`
	Data       []byte                 `json:"data"`
	Options    map[string]interface{} `json:"options"`
	ResultChan chan *JobResult        `json:"-"`
	CreatedAt  time.Time              `json:"created_at"`
	Priority   int                    `json:"priority"` // Higher numbers = higher priority
}

// JobResult contains the result of a request job
type JobResult struct {
	ID          string        `json:"id"`
	Response    *Response     `json:"response"`
	Error       error         `json:"error"`
	Duration    time.Duration `json:"duration"`
	CompletedAt time.Time     `json:"completed_at"`
	WorkerID    int           `json:"worker_id"`
}

// WorkerPool manages concurrent request workers
type WorkerPool struct {
	config         ConcurrencyConfig
	scraper        *AdvancedScraper
	jobQueue       chan *RequestJob
	resultQueue    chan *JobResult
	workers        []*Worker
	rateLimiter    *TokenBucket
	connectionPool *ConnectionPool
	stats          *WorkerPoolStats
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	mutex          sync.RWMutex
	running        bool
}

// Worker represents a single worker goroutine
type Worker struct {
	ID       int
	pool     *WorkerPool
	jobQueue chan *RequestJob
	quit     chan bool
	stats    *WorkerStats
}

// WorkerStats tracks individual worker performance
type WorkerStats struct {
	JobsProcessed   int64         `json:"jobs_processed"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	Errors          int64         `json:"errors"`
	LastJobAt       time.Time     `json:"last_job_at"`
}

// WorkerPoolStats tracks overall pool performance
type WorkerPoolStats struct {
	TotalJobs      int64         `json:"total_jobs"`
	CompletedJobs  int64         `json:"completed_jobs"`
	FailedJobs     int64         `json:"failed_jobs"`
	AverageLatency time.Duration `json:"average_latency"`
	RequestsPerSec float64       `json:"requests_per_sec"`
	ActiveWorkers  int           `json:"active_workers"`
	QueueLength    int           `json:"queue_length"`
	StartTime      time.Time     `json:"start_time"`
	mutex          sync.RWMutex
}

// TokenBucket implements rate limiting
type TokenBucket struct {
	capacity   int
	tokens     int
	refillRate int // tokens per second
	lastRefill time.Time
	mutex      sync.Mutex
}

// ConnectionPool manages HTTP connections efficiently
type ConnectionPool struct {
	config    ConcurrencyConfig
	transport *http.Transport
	clients   map[string]*http.Client // Per-host clients
	mutex     sync.RWMutex
}

// GetDefaultConcurrencyConfig returns sensible defaults
func GetDefaultConcurrencyConfig() ConcurrencyConfig {
	return ConcurrencyConfig{
		MaxConcurrent:     runtime.NumCPU() * 10, // 10x CPU cores
		WorkerPoolSize:    runtime.NumCPU() * 2,  // 2x CPU cores
		RequestBuffer:     1000,                  // 1000 queued requests
		RateLimitPerSec:   100,                   // 100 req/sec default
		ConnectionTimeout: 30 * time.Second,
		IdleTimeout:       90 * time.Second,
		MaxIdleConns:      100,
		MaxConnsPerHost:   20,
	}
}

// NewWorkerPool creates a new concurrent worker pool
func NewWorkerPool(scraper *AdvancedScraper, config ConcurrencyConfig) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		config:         config,
		scraper:        scraper,
		jobQueue:       make(chan *RequestJob, config.RequestBuffer),
		resultQueue:    make(chan *JobResult, config.RequestBuffer),
		workers:        make([]*Worker, config.WorkerPoolSize),
		rateLimiter:    NewTokenBucket(config.RateLimitPerSec),
		connectionPool: NewConnectionPool(config),
		stats:          &WorkerPoolStats{StartTime: time.Now()},
		ctx:            ctx,
		cancel:         cancel,
	}

	// Create workers
	for i := 0; i < config.WorkerPoolSize; i++ {
		worker := &Worker{
			ID:       i,
			pool:     pool,
			jobQueue: pool.jobQueue,
			quit:     make(chan bool),
			stats:    &WorkerStats{},
		}
		pool.workers[i] = worker
	}

	return pool
}

// NewTokenBucket creates a new rate limiter
func NewTokenBucket(requestsPerSec int) *TokenBucket {
	return &TokenBucket{
		capacity:   requestsPerSec,
		tokens:     requestsPerSec,
		refillRate: requestsPerSec,
		lastRefill: time.Now(),
	}
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config ConcurrencyConfig) *ConnectionPool {
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxConnsPerHost,
		IdleConnTimeout:     config.IdleTimeout,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   true,
	}

	return &ConnectionPool{
		config:    config,
		transport: transport,
		clients:   make(map[string]*http.Client),
	}
}

// Start begins the worker pool
func (wp *WorkerPool) Start() error {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if wp.running {
		return fmt.Errorf("worker pool is already running")
	}

	// Start all workers
	for _, worker := range wp.workers {
		wp.wg.Add(1)
		go worker.start()
	}

	// Start stats updater
	wp.wg.Add(1)
	go wp.statsUpdater()

	wp.running = true
	return nil
}

// Stop gracefully shuts down the worker pool
func (wp *WorkerPool) Stop() error {
	wp.mutex.Lock()
	if !wp.running {
		wp.mutex.Unlock()
		return fmt.Errorf("worker pool is not running")
	}
	wp.running = false
	wp.mutex.Unlock()

	// Cancel context to signal shutdown
	wp.cancel()

	// Close job queue
	close(wp.jobQueue)

	// Stop all workers
	for _, worker := range wp.workers {
		close(worker.quit)
	}

	// Wait for all workers to finish
	wp.wg.Wait()

	// Close result queue
	close(wp.resultQueue)

	return nil
}

// SubmitJob adds a new job to the queue
func (wp *WorkerPool) SubmitJob(job *RequestJob) error {
	wp.mutex.RLock()
	running := wp.running
	wp.mutex.RUnlock()

	if !running {
		return fmt.Errorf("worker pool is not running")
	}

	wp.stats.mutex.Lock()
	wp.stats.TotalJobs++
	wp.stats.QueueLength = len(wp.jobQueue)
	wp.stats.mutex.Unlock()

	select {
	case wp.jobQueue <- job:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// GetResult retrieves a completed job result
func (wp *WorkerPool) GetResult() (*JobResult, error) {
	select {
	case result := <-wp.resultQueue:
		return result, nil
	case <-wp.ctx.Done():
		return nil, fmt.Errorf("worker pool is shutting down")
	}
}

// GetStats returns current worker pool statistics
func (wp *WorkerPool) GetStats() WorkerPoolStats {
	wp.stats.mutex.RLock()
	defer wp.stats.mutex.RUnlock()

	stats := *wp.stats
	stats.ActiveWorkers = len(wp.workers)
	stats.QueueLength = len(wp.jobQueue)

	// Calculate requests per second
	duration := time.Since(stats.StartTime)
	if duration.Seconds() > 0 {
		stats.RequestsPerSec = float64(stats.CompletedJobs) / duration.Seconds()
	}

	return stats
}

// Worker methods
func (w *Worker) start() {
	defer w.pool.wg.Done()

	for {
		select {
		case job := <-w.jobQueue:
			if job == nil {
				return // Channel closed
			}
			w.processJob(job)

		case <-w.quit:
			return

		case <-w.pool.ctx.Done():
			return
		}
	}
}

func (w *Worker) processJob(job *RequestJob) {
	startTime := time.Now()

	// Wait for rate limiter
	w.pool.rateLimiter.WaitForToken()

	// Get HTTP client from connection pool
	client := w.pool.connectionPool.GetClient(job.URL)

	// Create a new scraper instance for this job (to avoid race conditions)
	jobScraper := *w.pool.scraper
	jobScraper.client = client

	// Perform the request
	var response *Response
	var err error

	switch job.Method {
	case "GET":
		response, err = jobScraper.Get(job.URL)
	case "POST":
		response, err = jobScraper.Post(job.URL, job.Data)
	default:
		err = fmt.Errorf("unsupported method: %s", job.Method)
	}

	duration := time.Since(startTime)

	// Update worker stats
	w.stats.JobsProcessed++
	w.stats.TotalDuration += duration
	w.stats.AverageDuration = w.stats.TotalDuration / time.Duration(w.stats.JobsProcessed)
	w.stats.LastJobAt = time.Now()
	if err != nil {
		w.stats.Errors++
	}

	// Create result
	result := &JobResult{
		ID:          job.ID,
		Response:    response,
		Error:       err,
		Duration:    duration,
		CompletedAt: time.Now(),
		WorkerID:    w.ID,
	}

	// Update pool stats
	w.pool.stats.mutex.Lock()
	w.pool.stats.CompletedJobs++
	if err != nil {
		w.pool.stats.FailedJobs++
	}
	w.pool.stats.mutex.Unlock()

	// Send result
	select {
	case w.pool.resultQueue <- result:
	case <-w.pool.ctx.Done():
		return
	}

	// Send result to job's channel if available
	if job.ResultChan != nil {
		select {
		case job.ResultChan <- result:
		default:
			// Channel might be closed or full
		}
	}
}

// Token bucket methods
func (tb *TokenBucket) WaitForToken() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	timePassed := now.Sub(tb.lastRefill)
	tokensToAdd := int(timePassed.Seconds()) * tb.refillRate

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	// Wait if no tokens available
	for tb.tokens <= 0 {
		tb.mutex.Unlock()
		time.Sleep(100 * time.Millisecond)
		tb.mutex.Lock()

		// Refill again after waiting
		now = time.Now()
		timePassed = now.Sub(tb.lastRefill)
		tokensToAdd = int(timePassed.Seconds()) * tb.refillRate

		if tokensToAdd > 0 {
			tb.tokens += tokensToAdd
			if tb.tokens > tb.capacity {
				tb.tokens = tb.capacity
			}
			tb.lastRefill = now
		}
	}

	// Consume a token
	tb.tokens--
}

// Connection pool methods
func (cp *ConnectionPool) GetClient(url string) *http.Client {
	cp.mutex.RLock()
	client, exists := cp.clients[url]
	cp.mutex.RUnlock()

	if exists {
		return client
	}

	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	// Double-check pattern
	client, exists = cp.clients[url]
	if exists {
		return client
	}

	// Create new client
	client = &http.Client{
		Transport: cp.transport,
		Timeout:   cp.config.ConnectionTimeout,
	}

	cp.clients[url] = client
	return client
}

// Stats updater goroutine
func (wp *WorkerPool) statsUpdater() {
	defer wp.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wp.updateStats()
		case <-wp.ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) updateStats() {
	wp.stats.mutex.Lock()
	defer wp.stats.mutex.Unlock()

	// Calculate average latency from all workers
	var totalDuration time.Duration
	var totalJobs int64

	for _, worker := range wp.workers {
		totalDuration += worker.stats.TotalDuration
		totalJobs += worker.stats.JobsProcessed
	}

	if totalJobs > 0 {
		wp.stats.AverageLatency = totalDuration / time.Duration(totalJobs)
	}

	// Update queue length
	wp.stats.QueueLength = len(wp.jobQueue)
}

// MemoryOptimizer handles memory management for high-volume operations
type MemoryOptimizer struct {
	maxMemoryMB         int64
	gcThresholdMB       int64
	lastGC              time.Time
	mutex               sync.RWMutex
	memoryCheckInterval time.Duration
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer(maxMemoryMB int64) *MemoryOptimizer {
	if maxMemoryMB <= 0 {
		maxMemoryMB = 512 // Default 512MB limit
	}

	return &MemoryOptimizer{
		maxMemoryMB:         maxMemoryMB,
		gcThresholdMB:       maxMemoryMB / 2, // Trigger GC at 50% of max
		lastGC:              time.Now(),
		memoryCheckInterval: 30 * time.Second,
	}
}

// CheckMemoryUsage monitors memory usage and triggers optimization if needed
func (mo *MemoryOptimizer) CheckMemoryUsage() bool {
	mo.mutex.RLock()
	defer mo.mutex.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	currentMemoryMB := int64(m.Alloc / 1024 / 1024)

	// Check if we need to force garbage collection
	if currentMemoryMB > mo.gcThresholdMB {
		if time.Since(mo.lastGC) > mo.memoryCheckInterval {
			mo.forceGarbageCollection()
			return true
		}
	}

	return currentMemoryMB > mo.maxMemoryMB
}

// forceGarbageCollection triggers manual garbage collection
func (mo *MemoryOptimizer) forceGarbageCollection() {
	mo.mutex.Lock()
	defer mo.mutex.Unlock()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	runtime.GC()
	runtime.GC() // Run twice for better cleanup

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	mo.lastGC = time.Now()

	// Log memory cleanup results (could be enabled with verbose flag)
	beforeMB := before.Alloc / 1024 / 1024
	afterMB := after.Alloc / 1024 / 1024
	_ = beforeMB // Avoid unused variable
	_ = afterMB  // Avoid unused variable
}

// GetMemoryStats returns current memory usage statistics
func (mo *MemoryOptimizer) GetMemoryStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"allocated_mb":       m.Alloc / 1024 / 1024,
		"total_allocated_mb": m.TotalAlloc / 1024 / 1024,
		"sys_mb":             m.Sys / 1024 / 1024,
		"gc_runs":            m.NumGC,
		"last_gc":            time.Unix(0, int64(m.LastGC)),
		"max_memory_mb":      mo.maxMemoryMB,
		"gc_threshold_mb":    mo.gcThresholdMB,
	}
}

// IntelligentQueue provides priority-based request queuing with load balancing
type IntelligentQueue struct {
	highPriority   chan *RequestJob
	normalPriority chan *RequestJob
	lowPriority    chan *RequestJob
	maxQueueSize   int
	mutex          sync.RWMutex
	stats          QueueStats
}

// QueueStats tracks queue performance metrics
type QueueStats struct {
	HighPriorityJobs   int64 `json:"high_priority_jobs"`
	NormalPriorityJobs int64 `json:"normal_priority_jobs"`
	LowPriorityJobs    int64 `json:"low_priority_jobs"`
	DroppedJobs        int64 `json:"dropped_jobs"`
	TotalQueued        int64 `json:"total_queued"`
}

// NewIntelligentQueue creates a new priority-based queue
func NewIntelligentQueue(maxSize int) *IntelligentQueue {
	queueSize := maxSize / 3 // Divide among three priority levels

	return &IntelligentQueue{
		highPriority:   make(chan *RequestJob, queueSize),
		normalPriority: make(chan *RequestJob, queueSize),
		lowPriority:    make(chan *RequestJob, queueSize),
		maxQueueSize:   maxSize,
	}
}

// EnqueueJob adds a job to the appropriate priority queue
func (iq *IntelligentQueue) EnqueueJob(job *RequestJob) error {
	iq.mutex.Lock()
	defer iq.mutex.Unlock()

	// Determine priority queue based on job priority
	var targetQueue chan *RequestJob
	switch {
	case job.Priority >= 8: // High priority (8-10)
		targetQueue = iq.highPriority
		iq.stats.HighPriorityJobs++
	case job.Priority >= 4: // Normal priority (4-7)
		targetQueue = iq.normalPriority
		iq.stats.NormalPriorityJobs++
	default: // Low priority (1-3)
		targetQueue = iq.lowPriority
		iq.stats.LowPriorityJobs++
	}

	// Try to enqueue, drop if queue is full
	select {
	case targetQueue <- job:
		iq.stats.TotalQueued++
		return nil
	default:
		iq.stats.DroppedJobs++
		return fmt.Errorf("queue full, job dropped (priority: %d)", job.Priority)
	}
}

// DequeueJob retrieves the next job with priority ordering
func (iq *IntelligentQueue) DequeueJob() (*RequestJob, bool) {
	// Check high priority first, then normal, then low
	select {
	case job := <-iq.highPriority:
		return job, true
	default:
		select {
		case job := <-iq.normalPriority:
			return job, true
		default:
			select {
			case job := <-iq.lowPriority:
				return job, true
			default:
				return nil, false
			}
		}
	}
}

// GetQueueStats returns current queue statistics
func (iq *IntelligentQueue) GetQueueStats() QueueStats {
	iq.mutex.RLock()
	defer iq.mutex.RUnlock()

	stats := iq.stats
	stats.HighPriorityJobs = int64(len(iq.highPriority))
	stats.NormalPriorityJobs = int64(len(iq.normalPriority))
	stats.LowPriorityJobs = int64(len(iq.lowPriority))

	return stats
}

// Close closes all priority queues
func (iq *IntelligentQueue) Close() {
	close(iq.highPriority)
	close(iq.normalPriority)
	close(iq.lowPriority)
}
