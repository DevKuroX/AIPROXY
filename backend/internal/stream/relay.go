package stream

import (
	"context"
	"io"
	"sync"
)

// Relay relays streams from a reader to multiple writers with tee functionality
// ref: open-sse/utils/streamHandler.js - pipeWithDisconnect
type Relay struct {
	reader     StreamReader
	writers    []StreamWriter
	teeWriters []io.Writer
	mu         sync.RWMutex
	closed     bool
}

// NewRelay creates a new stream relay
func NewRelay(reader StreamReader) *Relay {
	return &Relay{
		reader:  reader,
		writers: make([]StreamWriter, 0),
		closed:  false,
	}
}

// AddWriter adds a writer to the relay
func (r *Relay) AddWriter(writer StreamWriter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.writers = append(r.writers, writer)
}

// AddTeeWriter adds a tee writer for copying raw data
func (r *Relay) AddTeeWriter(writer io.Writer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.teeWriters = append(r.teeWriters, writer)
}

// Start starts the relay process
// ref: open-sse/utils/streamHandler.js - pipeWithDisconnect
func (r *Relay) Start(ctx context.Context) error {
	defer r.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		chunk, err := r.reader.ReadChunk(ctx)
		if err != nil {
			if err == io.EOF {
				// Stream completed normally
				return nil
			}
			return err
		}

		// Check for [DONE] marker
		if IsDoneChunk(chunk) {
			// Write [DONE] to all writers
			r.mu.RLock()
			for _, writer := range r.writers {
				if sseWriter, ok := writer.(*SSEWriter); ok {
					sseWriter.WriteDone()
				}
			}
			r.mu.RUnlock()
			return nil
		}

		// Relay chunk to all writers
		if err := r.relayChunk(ctx, chunk); err != nil {
			return err
		}
	}
}

// relayChunk relays a single chunk to all writers
func (r *Relay) relayChunk(ctx context.Context, chunk *StreamChunk) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Write to tee writers first (raw data)
	for _, teeWriter := range r.teeWriters {
		if _, err := teeWriter.Write(chunk.Data); err != nil {
			return err
		}
	}

	// Write to SSE writers
	var wg sync.WaitGroup
	errCh := make(chan error, len(r.writers))

	for _, writer := range r.writers {
		wg.Add(1)
		go func(w StreamWriter) {
			defer wg.Done()
			if err := w.WriteChunk(ctx, chunk); err != nil {
				select {
				case errCh <- err:
				default:
				}
			}
		}(writer)
	}

	wg.Wait()
	close(errCh)

	// Return first error if any
	for err := range errCh {
		return err
	}

	return nil
}

// Close closes all writers and the reader
func (r *Relay) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true

	// Close reader
	if r.reader != nil {
		r.reader.Close()
	}

	// Close all writers
	for _, writer := range r.writers {
		writer.Close()
	}

	return nil
}

// TeeReader creates a tee reader that copies data to multiple writers
// ref: open-sse/utils/stream.js - createPassthroughStreamWithLogger
func TeeReader(reader io.Reader, writers ...io.Writer) io.Reader {
	return &teeReader{
		reader:  reader,
		writers: writers,
	}
}

type teeReader struct {
	reader  io.Reader
	writers []io.Writer
}

func (t *teeReader) Read(p []byte) (n int, err error) {
	n, err = t.reader.Read(p)
	if n > 0 {
		for _, writer := range t.writers {
			writer.Write(p[:n])
		}
	}
	return n, err
}

// PipeWithContext pipes data from reader to writer with context cancellation
// ref: open-sse/utils/streamHandler.js - pipeWithDisconnect
func PipeWithContext(ctx context.Context, dst io.Writer, src io.Reader) error {
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		nr, err := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if ew != nil {
				return ew
			}
			if nr != nw {
				return io.ErrShortWrite
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}
