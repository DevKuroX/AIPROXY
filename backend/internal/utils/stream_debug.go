// ref: _ref/9router/open-sse/utils/streamHelpers.js
package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

var streamDebugMu sync.Mutex
var streamDebugEnabled = os.Getenv("STREAM_DEBUG") == "1"

func IsStreamDebugEnabled() bool {
	return streamDebugEnabled
}

func EnableStreamDebug(enabled bool) {
	streamDebugMu.Lock()
	defer streamDebugMu.Unlock()
	streamDebugEnabled = enabled
}

type StreamDebugger struct {
	prefix    string
	startTime time.Time
	enabled   bool
}

func NewStreamDebugger(prefix string) *StreamDebugger {
	return &StreamDebugger{
		prefix:    prefix,
		startTime: time.Now(),
		enabled:   streamDebugEnabled,
	}
}

func (d *StreamDebugger) Log(format string, args ...interface{}) {
	if !d.enabled {
		return
	}
	streamDebugMu.Lock()
	defer streamDebugMu.Unlock()
	elapsed := time.Since(d.startTime).Milliseconds()
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("[%s][%dms] %s\n", d.prefix, elapsed, msg)
}

func (d *StreamDebugger) LogChunk(data []byte) {
	if !d.enabled {
		return
	}
	d.Log("chunk(%d bytes): %s", len(data), strings.TrimSpace(string(data)))
}

func (d *StreamDebugger) LogJSON(label string, data interface{}) {
	if !d.enabled {
		return
	}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		d.Log("%s: <marshal error: %v>", label, err)
		return
	}
	d.Log("%s: %s", label, string(jsonData))
}

type DebugWriter struct {
	writer io.Writer
	debugger *StreamDebugger
	label   string
}

func NewDebugWriter(writer io.Writer, debugger *StreamDebugger, label string) *DebugWriter {
	return &DebugWriter{
		writer:   writer,
		debugger: debugger,
		label:    label,
	}
}

func (w *DebugWriter) Write(p []byte) (n int, err error) {
	if w.debugger != nil && w.debugger.enabled {
		w.debugger.LogChunk(p)
	}
	return w.writer.Write(p)
}

type DebugReader struct {
	reader   io.Reader
	debugger *StreamDebugger
	label    string
}

func NewDebugReader(reader io.Reader, debugger *StreamDebugger, label string) *DebugReader {
	return &DebugReader{
		reader:   reader,
		debugger: debugger,
		label:    label,
	}
}

func (r *DebugReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if r.debugger != nil && r.debugger.enabled && n > 0 {
		r.debugger.Log("read %d bytes: %s", n, strings.TrimSpace(string(p[:n])))
	}
	return n, err
}

type SSEStreamReader struct {
	reader   *bufio.Reader
	buffer   string
	debugger *StreamDebugger
}

func NewSSEStreamReader(reader io.Reader) *SSEStreamReader {
	return &SSEStreamReader{
		reader: bufio.NewReader(reader),
	}
}

func NewSSEStreamReaderWithDebug(reader io.Reader, debugger *StreamDebugger) *SSEStreamReader {
	return &SSEStreamReader{
		reader:   bufio.NewReader(reader),
		debugger: debugger,
	}
}

func (r *SSEStreamReader) ReadLine() (string, error) {
	line, err := r.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	if r.debugger != nil && r.debugger.enabled && line != "" {
		r.debugger.Log("line: %s", line)
	}

	return line, err
}

func (r *SSEStreamReader) ReadEvent() (map[string]interface{}, error) {
	for {
		line, err := r.ReadLine()
		if err == io.EOF {
			return nil, io.EOF
		}
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimSpace(data)

		if data == "[DONE]" {
			return map[string]interface{}{"done": true}, nil
		}

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			if r.debugger != nil && r.debugger.enabled {
				r.debugger.Log("parse error: %v for data: %s", err, data)
			}
			continue
		}

		return result, nil
	}
}

func (r *SSEStreamReader) ReadAll() ([]map[string]interface{}, error) {
	var events []map[string]interface{}
	for {
		event, err := r.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			return events, err
		}
		events = append(events, event)
	}
	return events, nil
}

func DumpStreamInfo(label string, reader io.Reader) {
	if !streamDebugEnabled {
		return
	}

	debugger := NewStreamDebugger(label)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		debugger.Log("data: %s", scanner.Text())
	}
}
