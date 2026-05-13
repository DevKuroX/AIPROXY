package api

import (
	"net/http"
	"net/http/pprof"

	"github.com/DevKuroX/AIPROXY/internal/api/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (r *Router) metricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		promhttp.Handler().ServeHTTP(w, req)
	}
}

func pprofHandler(name string) http.HandlerFunc {
	switch name {
	case "index":
		return pprof.Index
	case "cmdline":
		return pprof.Cmdline
	case "profile":
		return pprof.Profile
	case "symbol":
		return pprof.Symbol
	case "trace":
		return pprof.Trace
	case "goroutine":
		return pprof.Handler("goroutine").ServeHTTP
	case "heap":
		return pprof.Handler("heap").ServeHTTP
	case "threadcreate":
		return pprof.Handler("threadcreate").ServeHTTP
	case "block":
		return pprof.Handler("block").ServeHTTP
	case "allocs":
		return pprof.Handler("allocs").ServeHTTP
	case "mutex":
		return pprof.Handler("mutex").ServeHTTP
	default:
		return pprof.Index
	}
}

func requireAdminAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := middleware.GetClaimsFromContext(r.Context())
			if claims == nil || !claims.IsAdmin {
				http.Error(w, "admin access required", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
