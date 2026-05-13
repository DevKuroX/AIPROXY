package router

import (
	"sync"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/pricing"
)

// UsageRecord represents a complete usage record for logging/storage.
// ref: open-sse/services/usage.js
type UsageRecord struct {
	Model            string     `json:"model"`
	Provider         string     `json:"provider"`
	PromptTokens     int        `json:"prompt_tokens"`
	CompletionTokens int        `json:"completion_tokens"`
	TotalTokens      int        `json:"total_tokens"`
	Cost             float64    `json:"cost"`
	RTKSavings       int        `json:"rtk_savings"`
	BytesBefore      int        `json:"bytes_before"`
	BytesAfter       int        `json:"bytes_after"`
	Timestamp        time.Time  `json:"timestamp"`
	Duration         time.Duration `json:"duration,omitempty"`
}

// UsageRecorder handles async recording of usage data.
// ref: open-sse/services/usage.js
type UsageRecorder struct {
	mu       sync.RWMutex
	handlers []UsageHandler
	buffer   chan *UsageRecord
	wg       sync.WaitGroup
	stopCh   chan struct{}
}

// UsageHandler is called for each usage record.
type UsageHandler func(record *UsageRecord)

// NewUsageRecorder creates a new async usage recorder.
func NewUsageRecorder(bufferSize int) *UsageRecorder {
	if bufferSize <= 0 {
		bufferSize = 1000
	}

	r := &UsageRecorder{
		buffer:   make(chan *UsageRecord, bufferSize),
		stopCh:   make(chan struct{}),
		handlers: make([]UsageHandler, 0),
	}

	go r.processBuffer()

	return r
}

// AddHandler adds a usage handler that will be called for each record.
func (r *UsageRecorder) AddHandler(handler UsageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers = append(r.handlers, handler)
}

// Record queues a usage record for async processing (non-blocking).
// ref: open-sse/services/usage.js
func (r *UsageRecorder) Record(record *UsageRecord) {
	if record == nil {
		return
	}

	select {
	case r.buffer <- record:
	default:
		// Buffer full, drop record to avoid blocking
	}
}

// RecordUsage creates and records usage from TokenUsage.
func (r *UsageRecorder) RecordUsage(model, provider string, usage *TokenUsage, duration time.Duration) {
	if usage == nil {
		return
	}

	pricingInfo, _ := pricing.LookupPricing(model)
	cost := pricing.CalculateCost(usage.PromptTokens, usage.CompletionTokens, pricingInfo)

	record := &UsageRecord{
		Model:            model,
		Provider:         provider,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		Cost:             cost,
		RTKSavings:       usage.RTKSavings(),
		BytesBefore:      usage.BytesBefore,
		BytesAfter:       usage.BytesAfter,
		Timestamp:        time.Now(),
		Duration:         duration,
	}

	r.Record(record)
}

// processBuffer processes usage records from the buffer.
func (r *UsageRecorder) processBuffer() {
	for {
		select {
		case <-r.stopCh:
			return
		case record := <-r.buffer:
			r.callHandlers(record)
		}
	}
}

// callHandlers calls all registered handlers for a record.
func (r *UsageRecorder) callHandlers(record *UsageRecord) {
	r.mu.RLock()
	handlers := make([]UsageHandler, len(r.handlers))
	copy(handlers, r.handlers)
	r.mu.RUnlock()

	for _, handler := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Handler panicked, continue with other handlers
				}
			}()
			handler(record)
		}()
	}
}

// Stop stops the recorder and waits for pending records to be processed.
func (r *UsageRecorder) Stop() {
	close(r.stopCh)
	r.wg.Wait()
}

// Global usage recorder instance
var globalRecorder *UsageRecorder
var recorderOnce sync.Once

// GetGlobalRecorder returns the global usage recorder.
func GetGlobalRecorder() *UsageRecorder {
	recorderOnce.Do(func() {
		globalRecorder = NewUsageRecorder(10000)
	})
	return globalRecorder
}

// RecordUsage records usage using the global recorder (convenience function).
func RecordUsage(model, provider string, usage *TokenUsage, duration time.Duration) {
	GetGlobalRecorder().RecordUsage(model, provider, usage, duration)
}
