package model

import "time"

// Node is a single file or directory entry as returned by the API.
type Node struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	IsDir     bool   `json:"isDir"`
	Size      int64  `json:"size"`      // recursive size for dirs, file size for files
	Files     int64  `json:"files"`     // recursive file count (dirs only)
	Dirs      int64  `json:"dirs"`      // recursive dir count (dirs only)
	Ext       string `json:"ext,omitempty"`
	ModTime   int64  `json:"modTime"` // unix seconds
}

// FileTypeStat aggregates size/count by extension across a scan.
type FileTypeStat struct {
	Ext   string `json:"ext"`
	Size  int64  `json:"size"`
	Count int64  `json:"count"`
}

// ScanStatus describes the most recent scan.
type ScanStatus struct {
	ScanID      int64     `json:"scanId"`
	RootPath    string    `json:"rootPath"`
	StartedAt   time.Time `json:"startedAt"`
	FinishedAt  time.Time `json:"finishedAt"`
	DurationMs  int64     `json:"durationMs"`
	TotalSize   int64     `json:"totalSize"`
	TotalFiles  int64     `json:"totalFiles"`
	TotalDirs   int64     `json:"totalDirs"`
	DiskTotal   uint64    `json:"diskTotal"`
	DiskFree    uint64    `json:"diskFree"`
	Running     bool      `json:"running"`
	Error       string    `json:"error,omitempty"`
}
