package dashboard

import (
	"net/http"
	"os"
	"path/filepath"
)

func Handler(apiPort int) http.Handler {
	mux := http.NewServeMux()
	frontendDir := findFrontendDir()
	fileServer := http.FileServer(http.Dir(frontendDir))
	apiTarget := "http://localhost:" + itoa(apiPort)

	mux.HandleFunc("/api/", proxyHandler(apiTarget))
	mux.HandleFunc("/v1/", proxyHandler(apiTarget))
	mux.HandleFunc("/health", proxyHandler(apiTarget))

	os.MkdirAll("/tmp/aiproxy-uploads", 0755)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("/tmp/aiproxy-uploads"))))

	mux.Handle("/", fileServer)
	return mux
}

func findFrontendDir() string {
	// Next.js static export output (out/) first, then fallback to frontend/
	candidates := []string{
		"frontend/out", "../frontend/out", "../../frontend/out",
		"frontend", "../frontend", "../../frontend",
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	cwd, _ := os.Getwd()
	return cwd
}

func proxyHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxyURL := target + r.URL.Path
		if r.URL.RawQuery != "" {
			proxyURL += "?" + r.URL.RawQuery
		}
		proxyReq, err := http.NewRequest(r.Method, proxyURL, r.Body)
		if err != nil {
			http.Error(w, "Proxy error", http.StatusInternalServerError)
			return
		}
		for key, values := range r.Header {
			for _, v := range values {
				proxyReq.Header.Add(key, v)
			}
		}
		proxyReq.Header.Del("Proxy-Connection")
		resp, err := http.DefaultTransport.RoundTrip(proxyReq)
		if err != nil {
			http.Error(w, "Proxy error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		for key, values := range resp.Header {
			for _, v := range values {
				w.Header().Add(key, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		flusher, canFlush := w.(http.Flusher)
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
				if canFlush && resp.Header.Get("Content-Type") == "text/event-stream" {
					flusher.Flush()
				}
			}
			if err != nil {
				break
			}
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "1432"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func CORS(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		for _, o := range allowedOrigins {
			if o == origin {
				allowed = true
				break
			}
		}
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
