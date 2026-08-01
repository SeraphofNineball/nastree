package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"nastree/internal/api"
	"nastree/internal/run"
	"nastree/internal/store"
)

func main() {
	scanPath := envOr("SCAN_PATH", "/data")
	dbPath := envOr("DB_PATH", "/db/nastree.db")
	staticDir := envOr("STATIC_DIR", "./static")
	listenAddr := envOr("LISTEN_ADDR", ":8080")
	interval := envDuration("SCAN_INTERVAL", 6*time.Hour)
	retain := envInt("RETAIN_SCANS", 5)

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("failed to create db directory: %v", err)
	}

	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open store: %v", err)
	}
	defer st.Close()

	runner := run.New(scanPath, interval, retain, st)
	go runner.Start()

	apiServer := api.New(st, runner)

	mux := http.NewServeMux()
	mux.Handle("/api/", apiServer)
	mux.Handle("/", spaHandler(staticDir))

	log.Printf("nastree listening on %s (scanning %s every %s, retaining %d scans)", listenAddr, scanPath, interval, retain)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		log.Fatal(err)
	}
}

// spaHandler serves static files, falling back to index.html for any
// unmatched path so client-side routing works.
func spaHandler(dir string) http.Handler {
	fileServer := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(path); err != nil {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
