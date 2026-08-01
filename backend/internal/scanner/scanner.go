package scanner

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

// node is the in-memory representation of a file or directory built during a walk.
type node struct {
	name     string
	isDir    bool
	size     int64 // file size; for dirs this is filled in by aggregate()
	modTime  int64
	children []*node
	files    int64 // recursive file count, dirs only
	dirs     int64 // recursive dir count, dirs only
}

// ExtStat accumulates size/count for one extension during a walk.
type ExtStat struct {
	Size  int64
	Count int64
}

// Result is the outcome of a completed scan.
type Result struct {
	Root       *node
	RootPath   string
	ExtStats   map[string]*ExtStat
	TotalSize  int64
	TotalFiles int64
	TotalDirs  int64
	Duration   time.Duration
}

// Walk scans rootPath and returns the aggregated tree plus extension stats.
// It does not touch the database - callers persist the Result separately.
func Walk(rootPath string) (*Result, error) {
	start := time.Now()

	nodes := make(map[string]*node, 1024)
	root := &node{name: filepath.Base(rootPath), isDir: true}
	nodes[filepath.Clean(rootPath)] = root

	extStats := make(map[string]*ExtStat)

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip unreadable entries (permission errors on NAS shares are common)
			// rather than aborting the whole scan.
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		clean := filepath.Clean(path)
		if clean == filepath.Clean(rootPath) {
			info, ierr := d.Info()
			if ierr == nil {
				root.modTime = info.ModTime().Unix()
			}
			return nil
		}

		info, ierr := d.Info()
		if ierr != nil {
			return nil
		}

		n := &node{
			name:    d.Name(),
			isDir:   d.IsDir(),
			modTime: info.ModTime().Unix(),
		}
		if !d.IsDir() {
			n.size = info.Size()
			ext := strings.ToLower(filepath.Ext(d.Name()))
			if ext == "" {
				ext = "(no extension)"
			}
			es, ok := extStats[ext]
			if !ok {
				es = &ExtStat{}
				extStats[ext] = es
			}
			es.Size += n.size
			es.Count++
		}

		nodes[clean] = n
		parent := nodes[filepath.Clean(filepath.Dir(clean))]
		if parent != nil {
			parent.children = append(parent.children, n)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	aggregate(root)

	return &Result{
		Root:       root,
		RootPath:   rootPath,
		ExtStats:   extStats,
		TotalSize:  root.size,
		TotalFiles: root.files,
		TotalDirs:  root.dirs,
		Duration:   time.Since(start),
	}, nil
}

// Record is a flat, exported view of one tree node for persistence.
type Record struct {
	Path       string
	ParentPath string
	Name       string
	IsDir      bool
	Size       int64
	Files      int64
	Dirs       int64
	Ext        string
	ModTime    int64
}

// Records flattens the scanned tree (including the root) into a slice
// suitable for bulk insertion, without exposing the internal node type.
func (r *Result) Records() []Record {
	out := make([]Record, 0, r.TotalFiles+r.TotalDirs+1)
	var walk func(n *node, path, parentPath string)
	walk = func(n *node, path, parentPath string) {
		rec := Record{
			Path:       path,
			ParentPath: parentPath,
			Name:       n.name,
			IsDir:      n.isDir,
			Size:       n.size,
			Files:      n.files,
			Dirs:       n.dirs,
			ModTime:    n.modTime,
		}
		if !n.isDir {
			ext := strings.ToLower(filepath.Ext(n.name))
			if ext == "" {
				ext = "(no extension)"
			}
			rec.Ext = ext
		}
		out = append(out, rec)
		for _, c := range n.children {
			walk(c, filepath.Join(path, c.name), path)
		}
	}
	walk(r.Root, filepath.Clean(r.RootPath), "")
	return out
}

// aggregate computes recursive size/file/dir counts bottom-up.
func aggregate(n *node) {
	if !n.isDir {
		return
	}
	var size, files, dirs int64
	for _, c := range n.children {
		if c.isDir {
			aggregate(c)
			dirs += c.dirs + 1
		} else {
			files++
		}
		size += c.size
		files += c.files
	}
	n.size = size
	n.files = files
	n.dirs = dirs
}
