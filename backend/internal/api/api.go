package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"nastree/internal/model"
	"nastree/internal/store"
)

// Runner is the subset of scan-orchestration behavior the API needs.
type Runner interface {
	Trigger() bool // returns false if a scan is already running
	Running() bool
}

type Server struct {
	store  *store.Store
	runner Runner
	mux    *http.ServeMux
}

func New(st *store.Store, runner Runner) *Server {
	s := &Server{store: st, runner: runner, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /api/node", s.handleNode)
	s.mux.HandleFunc("GET /api/children", s.handleChildren)
	s.mux.HandleFunc("GET /api/tree", s.handleTree)
	s.mux.HandleFunc("GET /api/filetypes", s.handleFileTypes)
	s.mux.HandleFunc("GET /api/files", s.handleFiles)
	s.mux.HandleFunc("POST /api/scan/trigger", s.handleTrigger)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.store.Status()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	running := s.runner.Running()
	if status == nil {
		writeJSON(w, map[string]any{"running": running, "scanId": nil})
		return
	}
	status.Running = running
	writeJSON(w, status)
}

func (s *Server) handleNode(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	var n *model.Node
	var err error
	if path == "" {
		n, err = s.store.RootNode()
	} else {
		n, err = s.store.NodeAt(path)
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if n == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, n)
}

func (s *Server) handleChildren(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	children, err := s.store.Children(path)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, children)
}

const treeMaxNodes = 2000

func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	tree, err := s.store.Tree(path, treeMaxNodes)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tree == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, tree)
}

func (s *Server) handleFileTypes(w http.ResponseWriter, r *http.Request) {
	stats, err := s.store.FileTypes(30)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, stats)
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	params := store.FileSearchParams{
		Query:          q.Get("q"),
		MatchPath:      q.Get("matchPath") == "true",
		FoldersOnly:    q.Get("foldersOnly") == "true",
		DuplicatesOnly: q.Get("duplicatesOnly") == "true",
		DupMode:        q.Get("dupMode"),
		Limit:          limit,
	}
	results, err := s.store.SearchFiles(params)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, results)
}

func (s *Server) handleTrigger(w http.ResponseWriter, r *http.Request) {
	if !s.runner.Trigger() {
		writeErr(w, http.StatusConflict, "scan already running")
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
