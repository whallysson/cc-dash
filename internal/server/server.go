package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/whallysson/cc-dash/internal/index"
)

//go:embed all:static
var staticFS embed.FS

// Server is the cc-dash HTTP server.
type Server struct {
	idx      *index.Index
	mux      *http.ServeMux
	port     int
	srv      *http.Server
	wsHub    *WSHub
}

// New creates a new server.
func New(idx *index.Index, port int) *Server {
	s := &Server{
		idx:   idx,
		mux:   http.NewServeMux(),
		port:  port,
		wsHub: NewWSHub(),
	}
	s.registerRoutes()
	return s
}

// GetWSHub returns the WebSocket hub for external use (watcher).
func (s *Server) GetWSHub() *WSHub {
	return s.wsHub
}

// registerRoutes registers all API routes.
func (s *Server) registerRoutes() {
	// API routes
	s.mux.HandleFunc("GET /api/stats", s.handleStats)
	s.mux.HandleFunc("GET /api/sessions", s.handleSessions)
	s.mux.HandleFunc("GET /api/sessions/{id}", s.handleSessionDetail)
	s.mux.HandleFunc("GET /api/sessions/{id}/replay", s.handleSessionReplay)
	s.mux.HandleFunc("GET /api/projects", s.handleProjects)
	s.mux.HandleFunc("GET /api/projects/{slug}", s.handleProjectDetail)
	s.mux.HandleFunc("GET /api/costs", s.handleCosts)
	s.mux.HandleFunc("GET /api/tools", s.handleTools)
	s.mux.HandleFunc("GET /api/activity", s.handleActivity)
	s.mux.HandleFunc("GET /api/history", s.handleHistory)
	s.mux.HandleFunc("GET /api/memory", s.handleMemory)
	s.mux.HandleFunc("PATCH /api/memory", s.handleMemoryUpdate)
	s.mux.HandleFunc("GET /api/plans", s.handlePlans)
	s.mux.HandleFunc("GET /api/todos", s.handleTodos)
	s.mux.HandleFunc("GET /api/settings", s.handleSettings)
	s.mux.HandleFunc("GET /api/efficiency", s.handleEfficiency)
	s.mux.HandleFunc("POST /api/export", s.handleExport)
	s.mux.HandleFunc("GET /ws", s.wsHub.HandleWS)

	// SPA: serve static frontend with index.html fallback
	s.mux.Handle("/", s.spaHandler())
}

// spaHandler serves static files with index.html fallback (SPA routing).
func (s *Server) spaHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Printf("[server] warning: static frontend not found: %v", err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<!DOCTYPE html><html><body>
				<h1>cc-dash</h1>
				<p>Frontend not found. Run <code>make build-frontend</code> first.</p>
				<p>API available at <a href="/api/stats">/api/stats</a></p>
			</body></html>`)
		})
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the static file
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists
		f, err := sub.Open(path[1:]) // remove leading /
		if err != nil {
			// Fallback: serve index.html for SPA routes
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	})
}

// Start starts the HTTP server.
func (s *Server) Start(openBrowser bool) error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.port)

	s.srv = &http.Server{
		Addr:         addr,
		Handler:      s.withMiddleware(s.mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d in use: %w", s.port, err)
	}

	url := fmt.Sprintf("http://localhost:%d", s.port)
	log.Printf("[server] listening on %s", url)

	if openBrowser {
		go func() {
			time.Sleep(300 * time.Millisecond)
			openURL(url)
		}()
	}

	return s.srv.Serve(ln)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

// withMiddleware applies global middleware.
func (s *Server) withMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS para dev
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// openURL opens a URL in the default browser.
func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

// FindFreePort finds a free port starting from the given port.
func FindFreePort(startPort int) (int, error) {
	for port := startPort; port < startPort+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			continue
		}
		ln.Close()
		return port, nil
	}
	return 0, fmt.Errorf("no free port found between %d and %d", startPort, startPort+100)
}
