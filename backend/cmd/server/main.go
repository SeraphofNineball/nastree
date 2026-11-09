package main

import (
	"io"
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

	// Run as a desktop app's sidecar, NAStree must not outlive the app, even when the app is killed or crashes and
	// gets no chance to stop it. The app keeps a pipe to NAStree's stdin open; the OS closes it however the app ends.
	// Opt-in, because in a container stdin is usually closed from the start.
	if os.Getenv("EXIT_ON_STDIN_CLOSE") == "1" {
		go exitWhenClosed(os.Stdin, func() {
			log.Printf("stdin closed: the parent app has exited, stopping")
			os.Exit(0)
		})
	}

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

// exitWhenClosed calls exit once r reaches end of input or fails. Anything written to it is ignored.
func exitWhenClosed(r io.Reader, exit func()) {
	_, _ = io.Copy(io.Discard, r)
	exit()
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
