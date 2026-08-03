package web

import (
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Options configures the web server.
type Options struct {
	Host string
	Port int
	Open bool
}

// newMux builds the HTTP router with every API route plus the SPA handler.
// Extracted from Serve so tests can exercise handlers without a socket.
func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Profiles
	mux.HandleFunc("GET /api/profiles", handleListProfiles)
	mux.HandleFunc("POST /api/profiles", handleCreateProfile)
	mux.HandleFunc("GET /api/profiles/{name}", handleGetProfile)
	mux.HandleFunc("PUT /api/profiles/{name}", handleSaveProfile)
	mux.HandleFunc("DELETE /api/profiles/{name}", handleDeleteProfile)
	mux.HandleFunc("POST /api/profiles/{name}/rename", handleRenameProfile)
	mux.HandleFunc("POST /api/profiles/{name}/activate", handleActivateProfile)
	mux.HandleFunc("GET /api/profiles/{name}/export", handleExportProfile)

	// Active / diff / import / validate / schema
	mux.HandleFunc("GET /api/active", handleGetActive)
	mux.HandleFunc("GET /api/diff", handleDiff)
	mux.HandleFunc("POST /api/import", handleImport)
	mux.HandleFunc("POST /api/validate", handleValidate)
	mux.HandleFunc("GET /api/schema", handleSchema)
	mux.HandleFunc("GET /api/document-schema", handleDocumentSchema)
	mux.HandleFunc("GET /api/schema-check", handleSchemaCheck)

	// Models (specific catalog route before the wildcard provider route)
	mux.HandleFunc("GET /api/models", handleListModels)
	mux.HandleFunc("POST /api/models", handleCreateModel)
	mux.HandleFunc("GET /api/models/catalog", handleModelsCatalog)
	mux.HandleFunc("PUT /api/models/{provider}/{modelId}", handleUpdateModel)
	mux.HandleFunc("DELETE /api/models/{provider}/{modelId}", handleDeleteModel)

	// SPA (root) — must be registered last.
	mux.HandleFunc("/", spaHandler())

	return mux
}

// spaHandler serves the embedded SPA with client-route fallback. If the frontend
// has not been built, it serves a plain-text placeholder for non-API paths.
func spaHandler() http.HandlerFunc {
	sub, err := fs.Sub(distFS, "frontend/dist")
	built := err == nil
	if built {
		if _, statErr := fs.Stat(sub, "index.html"); statErr != nil {
			built = false
		}
	}

	if !built {
		return func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				writeErr(w, http.StatusNotFound, "not found")
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Web UI not built. Run: make web-build\n"))
		}
	}

	fileServer := http.FileServer(http.FS(sub))
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}

		// Serve the static file when it exists; otherwise fall back to
		// index.html so deep client routes survive a refresh.
		reqPath := strings.TrimPrefix(r.URL.Path, "/")
		if reqPath == "" {
			serveIndex(w, r, sub)
			return
		}
		if _, statErr := fs.Stat(sub, reqPath); statErr != nil {
			serveIndex(w, r, sub)
			return
		}
		fileServer.ServeHTTP(w, r)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request, sub fs.FS) {
	data, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "index.html missing")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// Serve starts the web server, binding host:port and serving the API + SPA.
func Serve(opts Options) error {
	addr := net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port))

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	url := "http://" + addr
	fmt.Printf("omo-profiler (omo.json profiles) \u2192 %s\n", url)

	if opts.Open {
		go func() {
			time.Sleep(300 * time.Millisecond)
			_ = openBrowser(url)
		}()
	}

	return http.Serve(listener, newMux())
}
