package router

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func ForwardRequest(ctx context.Context, upstreamURL, apiKey string, body []byte, stream bool) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", upstreamURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}

	client := NewProxyClient()
	return client.Do(req)
}

func CopyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func StreamBody(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	return err
}

type LineReader struct {
	src io.Reader
	buf []byte
}

func NewLineReader(r io.Reader) *LineReader {
	return &LineReader{
		src: r,
		buf: make([]byte, 0, 4096),
	}
}

func (lr *LineReader) ReadLine() ([]byte, error) {
	for {
		for i := 0; i < len(lr.buf)-1; i++ {
			if lr.buf[i] == '\n' {
				line := lr.buf[:i]
				if i > 0 && lr.buf[i-1] == '\r' {
					line = lr.buf[:i-1]
				}
				lr.buf = lr.buf[i+1:]
				return line, nil
			}
		}

		tmp := make([]byte, 1024)
		n, err := lr.src.Read(tmp)
		if err != nil {
			if err == io.EOF && len(lr.buf) > 0 {
				line := lr.buf
				lr.buf = nil
				return line, nil
			}
			return nil, err
		}

		lr.buf = append(lr.buf, tmp[:n]...)
	}
}
