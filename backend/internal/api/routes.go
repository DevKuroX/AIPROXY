package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/api/admin"
	"github.com/DevKuroX/AIPROXY/internal/api/chat"
	"github.com/DevKuroX/AIPROXY/internal/api/middleware"
	"github.com/DevKuroX/AIPROXY/internal/api/v1"
	"github.com/DevKuroX/AIPROXY/internal/router"
	"github.com/DevKuroX/AIPROXY/internal/storage"
)

type Router struct {
	jwtSecret         string
	users             admin.UserStore
	keyStore          middleware.KeyStore
	apiSecret         string
	analytics         admin.AnalyticsStore
	nodes             admin.NodeStore
	accounts          admin.AccountStore
	combos            admin.ComboStore
	aliases           admin.AliasStore
	keys              admin.KeyStore
	db                *storage.DB
	githubClientID    string
	githubClientSecret string
}

func NewRouter(jwtSecret string, users admin.UserStore, keyStore middleware.KeyStore, apiSecret string, analytics admin.AnalyticsStore, nodes admin.NodeStore, accounts admin.AccountStore, combos admin.ComboStore, aliases admin.AliasStore, keys admin.KeyStore, db *storage.DB, githubClientID, githubClientSecret string) *Router {
	return &Router{
		jwtSecret:         jwtSecret,
		users:             users,
		keyStore:          keyStore,
		apiSecret:         apiSecret,
		analytics:         analytics,
		nodes:             nodes,
		accounts:          accounts,
		combos:            combos,
		aliases:           aliases,
		keys:              keys,
		db:                db,
		githubClientID:    githubClientID,
		githubClientSecret: githubClientSecret,
	}
}

func (r *Router) Routes() http.Handler {
	mux := http.NewServeMux()

	authHandler := admin.NewHandler(r.jwtSecret, r.users)

	mux.HandleFunc("POST /api/login", authHandler.Login)
	mux.HandleFunc("POST /api/logout", authHandler.Logout)

	authMiddleware := middleware.RequireAuth(r.jwtSecret)
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /api/me", authHandler.Me)
	protectedMux.HandleFunc("GET /api/providers", v1.HandleListProviders)

	mux.Handle("/api/me", authMiddleware(protectedMux))
	mux.Handle("/api/providers", authMiddleware(protectedMux))

	analyticsHandler := admin.NewAnalyticsHandler(r.analytics)
	nodeHandler := admin.NewNodeHandler(r.nodes)
	accountHandler := admin.NewAccountHandler(r.accounts)
	comboHandler := admin.NewComboHandler(r.combos)
	aliasHandler := admin.NewAliasHandler(r.aliases)
	keyHandler := admin.NewKeyHandler(r.keys, r.apiSecret)

	adminMux := http.NewServeMux()

	// Usage
	adminMux.HandleFunc("GET /api/admin/usage", analyticsHandler.ListUsage)
	adminMux.HandleFunc("GET /api/admin/usage/stats", analyticsHandler.GetUsageStats)
	adminMux.HandleFunc("GET /api/admin/pricing", analyticsHandler.ListPricing)
	adminMux.HandleFunc("POST /api/admin/pricing", analyticsHandler.CreatePricing)
	adminMux.HandleFunc("POST /api/admin/pricing/{id}", analyticsHandler.UpdatePricing)
	adminMux.HandleFunc("DELETE /api/admin/pricing/{id}", analyticsHandler.DeletePricing)

	// Provider nodes
	adminMux.HandleFunc("GET /api/provider-nodes", nodeHandler.ListNodes)
	adminMux.HandleFunc("POST /api/provider-nodes", nodeHandler.CreateNode)
	adminMux.HandleFunc("PATCH /api/provider-nodes/{id}", nodeHandler.UpdateNode)
	adminMux.HandleFunc("DELETE /api/provider-nodes/{id}", nodeHandler.DeleteNode)
	adminMux.HandleFunc("POST /api/provider-nodes/{id}/test", nodeHandler.TestNode)

	// Accounts
	adminMux.HandleFunc("GET /api/admin/accounts", accountHandler.List)
	adminMux.HandleFunc("POST /api/admin/accounts", accountHandler.Create)
	adminMux.HandleFunc("GET /api/admin/accounts/{id}", accountHandler.Get)
	adminMux.HandleFunc("PUT /api/admin/accounts/{id}", accountHandler.Update)
	adminMux.HandleFunc("DELETE /api/admin/accounts/{id}", accountHandler.Delete)

	// Combos
	adminMux.HandleFunc("GET /api/admin/combos", comboHandler.List)
	adminMux.HandleFunc("POST /api/admin/combos", comboHandler.Create)
	adminMux.HandleFunc("PUT /api/admin/combos/{id}", comboHandler.Update)
	adminMux.HandleFunc("DELETE /api/admin/combos/{id}", comboHandler.Delete)

	// Aliases
	adminMux.HandleFunc("GET /api/admin/aliases", aliasHandler.ListAliases)
	adminMux.HandleFunc("POST /api/admin/aliases", aliasHandler.CreateAlias)
	adminMux.HandleFunc("DELETE /api/admin/aliases/{id}", aliasHandler.DeleteAlias)

	// API Keys
	adminMux.HandleFunc("GET /api/admin/keys", keyHandler.List)
	adminMux.HandleFunc("POST /api/admin/keys", keyHandler.Create)
	adminMux.HandleFunc("DELETE /api/admin/keys/{id}", keyHandler.Delete)

	rateLimiter := middleware.NewRateLimiter()
	mux.Handle("/api/admin/", authMiddleware(rateLimiter.RateLimitByIP(adminMux)))

	chatHandler := chat.NewHandler(r.db, r.apiSecret)
	chatMux := http.NewServeMux()
	chatMux.HandleFunc("GET /api/chat/sessions", chatHandler.ListSessions)
	chatMux.HandleFunc("POST /api/chat/sessions", chatHandler.CreateSession)
	chatMux.HandleFunc("DELETE /api/chat/sessions/{id}", chatHandler.DeleteSession)
	chatMux.HandleFunc("GET /api/chat/sessions/{id}/messages", chatHandler.ListMessages)
	chatMux.HandleFunc("POST /api/chat/sessions/{id}/messages", chatHandler.SaveMessage)
	chatMux.HandleFunc("GET /api/chat/artifacts/{id}", chatHandler.GetArtifact)
	chatMux.HandleFunc("POST /api/chat/artifacts", chatHandler.CreateArtifact)
	chatMux.HandleFunc("POST /api/chat/completions", chatHandler.StreamCompletion)
	chatMux.HandleFunc("POST /api/chat/sessions/{id}/generate-title", chatHandler.GenerateTitle)
	chatMux.HandleFunc("POST /api/chat/files", chatHandler.UploadFile)
	chatMux.HandleFunc("GET /api/chat/models", chatHandler.ListProviderModels)

	githubHandler := chat.NewGitHubHandler(r.db, r.githubClientID, r.githubClientSecret)
	chatMux.HandleFunc("POST /api/chat/github/auth/start", githubHandler.StartAuth)
	chatMux.HandleFunc("POST /api/chat/github/auth/poll", githubHandler.PollAuth)
	chatMux.HandleFunc("POST /api/chat/github/api", githubHandler.ProxyAPI)
	mux.Handle("/api/chat/", authMiddleware(rateLimiter.RateLimitByIP(chatMux)))

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

	// DCP — Context Deduplication/Pruning
	proxyMux.HandleFunc("POST /v1/dcp", router.HandleDCP)

	// Model discovery (v1 handlers support kind-based filtering)
	proxyMux.HandleFunc("GET /v1/models", v1.HandleModels)
	proxyMux.HandleFunc("GET /v1/models/info", v1.HandleModelInfo)
	proxyMux.HandleFunc("GET /v1/models/{kind}", v1.HandleModelsByKind)

	rateLimitedMux := rateLimiter.RateLimitByIP(proxyMux)
	mux.Handle("/v1/", apiKeyMiddleware(rateLimitedMux))

	mux.Handle("GET /api/translator/console-logs/stream", authMiddleware(http.HandlerFunc(handleConsoleLogStream)))
	mux.Handle("DELETE /api/translator/console-logs", authMiddleware(http.HandlerFunc(handleConsoleLogClear)))

	// Usage/quota endpoint for Kiro
	usageHandler := NewUsageHandler(router.GetGlobalPool())
	mux.Handle("GET /api/usage/kiro", apiKeyMiddleware(http.HandlerFunc(usageHandler.GetKiroUsage)))

	if p := router.GetProxyAPI(); p != nil {
		proxyAPI := p.(*ProxyAPI)
		proxyMgmtMux := http.NewServeMux()
		proxyMgmtMux.HandleFunc("GET /api/proxies", proxyAPI.ListProxies)
		proxyMgmtMux.HandleFunc("POST /api/proxies", proxyAPI.AddProxy)
		proxyMgmtMux.HandleFunc("DELETE /api/proxies/{id}", proxyAPI.DeleteProxy)
		proxyMgmtMux.HandleFunc("GET /api/proxy-pools", proxyAPI.ListPools)
		proxyMgmtMux.HandleFunc("POST /api/proxy-pools", proxyAPI.CreatePool)
		proxyMgmtMux.HandleFunc("DELETE /api/proxy-pools/{id}", proxyAPI.DeletePool)
		proxyMgmtMux.HandleFunc("POST /api/proxy-pools/test-all", proxyAPI.TestAllPools)
		proxyMgmtMux.HandleFunc("POST /api/proxy-pools/{id}/test", proxyAPI.TestPool)
		proxyMgmtMux.HandleFunc("GET /api/proxy/settings", proxyAPI.GetSettings)
		proxyMgmtMux.HandleFunc("POST /api/proxy/settings", proxyAPI.UpdateSettings)
		proxyMgmtMux.HandleFunc("POST /api/scraper/start", proxyAPI.StartScraper)
		proxyMgmtMux.HandleFunc("POST /api/scraper/webshare", proxyAPI.ScrapeWebshare)
		proxyMgmtMux.HandleFunc("GET /api/scraper/progress", proxyAPI.ScraperProgress)
		rateLimitedProxyMgmt := rateLimiter.RateLimitByIP(proxyMgmtMux)
		mux.Handle("/api/proxies", apiKeyMiddleware(rateLimitedProxyMgmt))
		mux.Handle("/api/proxies/", apiKeyMiddleware(rateLimitedProxyMgmt))
		mux.Handle("/api/proxy-pools", apiKeyMiddleware(rateLimitedProxyMgmt))
		mux.Handle("/api/proxy-pools/", apiKeyMiddleware(rateLimitedProxyMgmt))
		mux.Handle("/api/proxy/", apiKeyMiddleware(rateLimitedProxyMgmt))
		mux.Handle("/api/scraper/", apiKeyMiddleware(rateLimitedProxyMgmt))
	}

	return middleware.CORS(mux)
}

func handleConsoleLogStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, _ := w.(http.Flusher)

	logFiles := []string{"/tmp/aiproxy-run.log", "/home/ubuntu/ai_proxy/backend/bin/server.log"}
	var logFile string
	for _, f := range logFiles {
		if _, err := os.Stat(f); err == nil {
			logFile = f
			break
		}
	}

	if logFile == "" {
		msg, _ := json.Marshal(map[string]string{"type": "init"})
		fmt.Fprintf(w, "data: %s\n\n", msg)
		flusher.Flush()
		<-r.Context().Done()
		return
	}

	f, _ := os.Open(logFile)
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if len(lines) > 200 {
		lines = lines[len(lines)-200:]
	}

	initMsg, _ := json.Marshal(map[string]interface{}{"type": "init", "logs": lines})
	fmt.Fprintf(w, "data: %s\n\n", initMsg)
	flusher.Flush()

	lastSize := len(lines)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			f2, _ := os.Open(logFile)
			if f2 == nil {
				continue
			}
			fi, _ := f2.Stat()
			if fi == nil {
				f2.Close()
				continue
			}
			if fi.Size() == 0 {
				f2.Close()
				continue
			}
			f2.Seek(0, 0)
			scanner2 := bufio.NewScanner(f2)
			var allLines []string
			for scanner2.Scan() {
				allLines = append(allLines, scanner2.Text())
			}
			f2.Close()
			if len(allLines) > lastSize {
				for _, line := range allLines[lastSize:] {
					msg, _ := json.Marshal(map[string]string{"type": "line", "line": line})
					fmt.Fprintf(w, "data: %s\n\n", msg)
				}
				flusher.Flush()
				lastSize = len(allLines)
			}
		}
	}
}

func handleConsoleLogClear(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}


