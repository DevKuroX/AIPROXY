package api

import (
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/api/admin"
	"github.com/DevKuroX/AIPROXY/internal/api/middleware"
	"github.com/DevKuroX/AIPROXY/internal/api/v1"
	"github.com/DevKuroX/AIPROXY/internal/router"
)

type Router struct {
	jwtSecret string
	users     admin.UserStore
	keyStore  middleware.KeyStore
	apiSecret string
	analytics admin.AnalyticsStore
	nodes     admin.NodeStore
}

func NewRouter(jwtSecret string, users admin.UserStore, keyStore middleware.KeyStore, apiSecret string, analytics admin.AnalyticsStore, nodes admin.NodeStore) *Router {
	return &Router{
		jwtSecret: jwtSecret,
		users:     users,
		keyStore:  keyStore,
		apiSecret: apiSecret,
		analytics: analytics,
		nodes:     nodes,
	}
}

func (r *Router) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	authHandler := admin.NewHandler(r.jwtSecret, r.users)

	mux.HandleFunc("POST /api/login", authHandler.Login)
	mux.HandleFunc("POST /api/logout", authHandler.Logout)

	authMiddleware := middleware.RequireAuth(r.jwtSecret)
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /api/me", authHandler.Me)

	mux.Handle("/api/me", authMiddleware(protectedMux))

	analyticsHandler := admin.NewAnalyticsHandler(r.analytics)
	nodeHandler := admin.NewNodeHandler(r.nodes)
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("GET /api/admin/usage", analyticsHandler.ListUsage)
	adminMux.HandleFunc("GET /api/admin/usage/stats", analyticsHandler.GetUsageStats)
	adminMux.HandleFunc("GET /api/admin/pricing", analyticsHandler.ListPricing)
	adminMux.HandleFunc("POST /api/admin/pricing", analyticsHandler.CreatePricing)
	adminMux.HandleFunc("POST /api/admin/pricing/{id}", analyticsHandler.UpdatePricing)
	adminMux.HandleFunc("DELETE /api/admin/pricing/{id}", analyticsHandler.DeletePricing)
	adminMux.HandleFunc("GET /api/provider-nodes", nodeHandler.ListNodes)
	adminMux.HandleFunc("POST /api/provider-nodes", nodeHandler.CreateNode)
	adminMux.HandleFunc("PATCH /api/provider-nodes/{id}", nodeHandler.UpdateNode)
	adminMux.HandleFunc("DELETE /api/provider-nodes/{id}", nodeHandler.DeleteNode)
	adminMux.HandleFunc("POST /api/provider-nodes/{id}/test", nodeHandler.TestNode)

	mux.Handle("/api/admin/", authMiddleware(adminMux))

	adminAuthMiddleware := requireAdminAuth(r.jwtSecret)
	metricsMux := http.NewServeMux()
	metricsMux.HandleFunc("GET /metrics", r.metricsHandler())
	mux.Handle("/metrics", adminAuthMiddleware(authMiddleware(metricsMux)))

	debugMux := http.NewServeMux()
	debugMux.HandleFunc("GET /debug/pprof/", pprofHandler("index"))
	debugMux.HandleFunc("GET /debug/pprof/cmdline", pprofHandler("cmdline"))
	debugMux.HandleFunc("GET /debug/pprof/profile", pprofHandler("profile"))
	debugMux.HandleFunc("GET /debug/pprof/symbol", pprofHandler("symbol"))
	debugMux.HandleFunc("GET /debug/pprof/trace", pprofHandler("trace"))
	debugMux.HandleFunc("GET /debug/pprof/goroutine", pprofHandler("goroutine"))
	debugMux.HandleFunc("GET /debug/pprof/heap", pprofHandler("heap"))
	debugMux.HandleFunc("GET /debug/pprof/threadcreate", pprofHandler("threadcreate"))
	debugMux.HandleFunc("GET /debug/pprof/block", pprofHandler("block"))
	debugMux.HandleFunc("GET /debug/pprof/allocs", pprofHandler("allocs"))
	debugMux.HandleFunc("GET /debug/pprof/mutex", pprofHandler("mutex"))
	mux.Handle("/debug/pprof/", adminAuthMiddleware(authMiddleware(debugMux)))

	apiKeyMiddleware := middleware.RequireAPIKey(r.keyStore, r.apiSecret)
	proxyMux := http.NewServeMux()
	proxyMux.HandleFunc("POST /v1/chat/completions", router.HandleChatCompletions)
	proxyMux.HandleFunc("POST /v1/images/generations", v1.HandleImageGenerations)
	proxyMux.HandleFunc("POST /v1/audio/speech", v1.HandleTTSSpeech)
	proxyMux.HandleFunc("POST /v1/audio/transcriptions", v1.HandleAudioTranscriptions) // ref: open-sse/handlers/sttCore.js
	proxyMux.HandleFunc("POST /v1/search", v1.HandleSearch)                            // ref: open-sse/handlers/search/index.js
	proxyMux.HandleFunc("POST /v1/fetch", v1.HandleFetch)                             // ref: open-sse/handlers/fetch/index.js
	proxyMux.HandleFunc("POST /v1/embeddings", v1.HandleEmbeddings)                   // ref: open-sse/handlers/embeddingsCore.js

	// Anthropic /v1/messages endpoint
	proxyMux.HandleFunc("POST /v1/messages", router.HandleAnthropicMessages)

	// Conversation compact endpoint
	proxyMux.HandleFunc("POST /v1/responses/compact", router.HandleCompact)

	// Model discovery (v1 handlers support kind-based filtering)
	proxyMux.HandleFunc("GET /v1/models", v1.HandleModels)
	proxyMux.HandleFunc("GET /v1/models/info", v1.HandleModelInfo)
	proxyMux.HandleFunc("GET /v1/models/{kind}", v1.HandleModelsByKind)

	mux.Handle("/v1/", apiKeyMiddleware(proxyMux))

	// Usage/quota endpoint for Kiro
	usageHandler := NewUsageHandler(router.GetGlobalPool())
	mux.HandleFunc("GET /api/usage/kiro", usageHandler.GetKiroUsage)

	// Proxy API
	if p := router.GetProxyAPI(); p != nil {
		proxyAPI := p.(*ProxyAPI)
		mux.HandleFunc("GET /api/proxies", proxyAPI.ListProxies)
		mux.HandleFunc("POST /api/proxies", proxyAPI.AddProxy)
		mux.HandleFunc("DELETE /api/proxies/{id}", proxyAPI.DeleteProxy)
		mux.HandleFunc("GET /api/proxy-pools", proxyAPI.ListPools)
		mux.HandleFunc("POST /api/proxy-pools", proxyAPI.CreatePool)
		mux.HandleFunc("DELETE /api/proxy-pools/{id}", proxyAPI.DeletePool)
		mux.HandleFunc("POST /api/proxy-pools/{id}/test", proxyAPI.TestPool)
		mux.HandleFunc("GET /api/proxy/settings", proxyAPI.GetSettings)
		mux.HandleFunc("POST /api/proxy/settings", proxyAPI.UpdateSettings)
		mux.HandleFunc("POST /api/scraper/start", proxyAPI.StartScraper)
		mux.HandleFunc("GET /api/scraper/progress", proxyAPI.ScraperProgress)
	}

	return mux
}
