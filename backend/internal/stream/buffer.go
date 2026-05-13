package stream

import (
	"context"
	"io"
	"sync"
)

// CircularBuffer implements a memory-efficient circular buffer for stream chunks
// ref: open-sse/utils/streamHelpers.js - createBuffer
type CircularBuffer struct {
	mu       sync.RWMutex
	buffer   []*StreamChunk
	capacity int
	head     int // Write position
	tail     int // Read position
	count    int // Number of items in buffer
	closed   bool
	notify   chan struct{} // Notify when data available
}

// NewCircularBuffer creates a new circular buffer with given capacity
// ref: open-sse/utils/streamHelpers.js - createBuffer
func NewCircularBuffer(capacity int) *CircularBuffer {
	if capacity <= 0 {
		capacity = 100 // Default capacity
	}
	return &CircularBuffer{
		buffer:   make([]*StreamChunk, capacity),
		capacity: capacity,
		notify:   make(chan struct{}, 1),
	}
}

// Write adds a chunk to the buffer
// ref: open-sse/utils/streamHelpers.js - writeToBuffer
func (b *CircularBuffer) Write(ctx context.Context, chunk *StreamChunk) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return io.ErrClosedPipe
	}

	// If buffer is full, wait for space (with context)
	for b.count >= b.capacity {
		b.mu.Unlock()
		select {
		case <-ctx.Done():
			b.mu.Lock()
			return ctx.Err()
		case <-b.notify: // Wait for reader to consume
		}
		b.mu.Lock()
		if b.closed {
			return io.ErrClosedPipe
		}
	}

	// Write to buffer
	b.buffer[b.head] = chunk
	b.head = (b.head + 1) % b.capacity
	b.count++

	// Notify waiting readers
	select {
	case b.notify <- struct{}{}:
	default:
	}

	return nil
}

// Read reads a chunk from the buffer (blocks until data available)
// ref: open-sse/utils/streamHelpers.js - readFromBuffer
func (b *CircularBuffer) Read(ctx context.Context) (*StreamChunk, error) {
	b.mu.Lock()

	// Wait for data (with context)
	for b.count == 0 && !b.closed {
		b.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-b.notify:
		}
		b.mu.Lock()
	}

	if b.closed && b.count == 0 {
		b.mu.Unlock()
		return nil, io.EOF
	}

	// Read from buffer
	chunk := b.buffer[b.tail]
	b.buffer[b.tail] = nil // Clear reference
	b.tail = (b.tail + 1) % b.capacity
	b.count--

	b.mu.Unlock()
	return chunk, nil
}

// TryRead attempts to read a chunk without blocking
// ref: open-sse/utils/streamHelpers.js - tryReadFromBuffer
func (b *CircularBuffer) TryRead() (*StreamChunk, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.count == 0 {
		return nil, false
	}

	chunk := b.buffer[b.tail]
	b.buffer[b.tail] = nil
	b.tail = (b.tail + 1) % b.capacity
	b.count--

	return chunk, true
}

// Peek returns the next chunk without removing it
// ref: open-sse/utils/streamHelpers.js - peekBuffer
func (b *CircularBuffer) Peek() (*StreamChunk, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.count == 0 {
		return nil, false
	}
	return b.buffer[b.tail], true
}

// Size returns the number of chunks in the buffer
// ref: open-sse/utils/streamHelpers.js - bufferSize
func (b *CircularBuffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.count
}

// Capacity returns the buffer capacity
func (b *CircularBuffer) Capacity() int {
	return b.capacity
}

// IsFull returns true if buffer is full
func (b *CircularBuffer) IsFull() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.count >= b.capacity
}

// IsEmpty returns true if buffer is empty
func (b *CircularBuffer) IsEmpty() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.count == 0
}

// Close closes the buffer
// ref: open-sse/utils/streamHelpers.js - closeBuffer
func (b *CircularBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true
	close(b.notify)

	// Clear references to help GC
	for i := range b.buffer {
		b.buffer[i] = nil
	}

	return nil
}

// Clear removes all chunks from the buffer
// ref: open-sse/utils/streamHelpers.js - clearBuffer
func (b *CircularBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := range b.buffer {
		b.buffer[i] = nil
	}
	b.head = 0
	b.tail = 0
	b.count = 0
}

// Drain reads all available chunks without blocking
// ref: open-sse/utils/streamHelpers.js - drainBuffer
func (b *CircularBuffer) Drain() []*StreamChunk {
	b.mu.Lock()
	defer b.mu.Unlock()

	chunks := make([]*StreamChunk, b.count)
	for i := 0; i < b.count; i++ {
		idx := (b.tail + i) % b.capacity
		chunks[i] = b.buffer[idx]
		b.buffer[idx] = nil
	}

	b.tail = b.head
	b.count = 0
	return chunks
}