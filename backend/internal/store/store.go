package store

import (
	"database/sql"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"nastree/internal/model"
	"nastree/internal/scanner"
)

type Store struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS scans (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	root_path TEXT NOT NULL,
	started_at INTEGER NOT NULL,
	finished_at INTEGER NOT NULL,
	duration_ms INTEGER NOT NULL,
	total_size INTEGER NOT NULL,
	total_files INTEGER NOT NULL,
	total_dirs INTEGER NOT NULL,
	disk_total INTEGER NOT NULL,
	disk_free INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS nodes (
	scan_id INTEGER NOT NULL,
	path TEXT NOT NULL,
	parent_path TEXT NOT NULL,
	name TEXT NOT NULL,
	is_dir INTEGER NOT NULL,
	size INTEGER NOT NULL,
	files INTEGER NOT NULL,
	dirs INTEGER NOT NULL,
	ext TEXT NOT NULL DEFAULT '',
	mod_time INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_nodes_scan_parent ON nodes(scan_id, parent_path);
CREATE INDEX IF NOT EXISTS idx_nodes_scan_ext ON nodes(scan_id, ext);
CREATE INDEX IF NOT EXISTS idx_nodes_scan_path ON nodes(scan_id, path);
CREATE INDEX IF NOT EXISTS idx_nodes_scan_name_size ON nodes(scan_id, name, size);
`

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // modernc sqlite: keep writes serialized, simplest for our access pattern
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// SaveScan persists a completed scan and prunes older scans beyond retain.
func (s *Store) SaveScan(res *scanner.Result, diskTotal, diskFree uint64, startedAt, finishedAt time.Time, retain int) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO scans (root_path, started_at, finished_at, duration_ms, total_size, total_files, total_dirs, disk_total, disk_free)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filepath.Clean(res.RootPath), startedAt.Unix(), finishedAt.Unix(), finishedAt.Sub(startedAt).Milliseconds(),
		res.TotalSize, res.TotalFiles, res.TotalDirs, diskTotal, diskFree,
	)
	if err != nil {
		return 0, err
	}
	scanID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	stmt, err := tx.Prepare(
		`INSERT INTO nodes (scan_id, path, parent_path, name, is_dir, size, files, dirs, ext, mod_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return 0, err
	}
	for _, r := range res.Records() {
		if _, err := stmt.Exec(scanID, r.Path, r.ParentPath, r.Name, r.IsDir, r.Size, r.Files, r.Dirs, r.Ext, r.ModTime); err != nil {
			stmt.Close()
			return 0, err
		}
	}
	stmt.Close()

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	if err := s.prune(retain); err != nil {
		return scanID, err
	}
	return scanID, nil
}

func (s *Store) prune(retain int) error {
	if retain <= 0 {
		retain = 1
	}
	rows, err := s.db.Query(`SELECT id FROM scans ORDER BY id DESC LIMIT -1 OFFSET ?`, retain)
	if err != nil {
		return err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()

	for _, id := range ids {
		if _, err := s.db.Exec(`DELETE FROM nodes WHERE scan_id = ?`, id); err != nil {
			return err
		}
		if _, err := s.db.Exec(`DELETE FROM scans WHERE id = ?`, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) latestScanRow() (id int64, rootPath string, err error) {
	err = s.db.QueryRow(`SELECT id, root_path FROM scans ORDER BY id DESC LIMIT 1`).Scan(&id, &rootPath)
	return
}

// Status returns metadata for the most recent persisted scan.
// It returns (nil, nil) if no scan has completed yet.
func (s *Store) Status() (*model.ScanStatus, error) {
	row := s.db.QueryRow(
		`SELECT id, root_path, started_at, finished_at, duration_ms, total_size, total_files, total_dirs, disk_total, disk_free
		 FROM scans ORDER BY id DESC LIMIT 1`,
	)
	var st model.ScanStatus
	var startedAt, finishedAt int64
	var diskTotal, diskFree int64
	err := row.Scan(&st.ScanID, &st.RootPath, &startedAt, &finishedAt, &st.DurationMs, &st.TotalSize, &st.TotalFiles, &st.TotalDirs, &diskTotal, &diskFree)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	st.StartedAt = time.Unix(startedAt, 0)
	st.FinishedAt = time.Unix(finishedAt, 0)
	st.DiskTotal = uint64(diskTotal)
	st.DiskFree = uint64(diskFree)
	return &st, nil
}

// RootNode returns the root entry of the most recent scan.
func (s *Store) RootNode() (*model.Node, error) {
	scanID, rootPath, err := s.latestScanRow()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s.nodeAt(scanID, rootPath)
}

func (s *Store) nodeAt(scanID int64, path string) (*model.Node, error) {
	row := s.db.QueryRow(
		`SELECT path, name, is_dir, size, files, dirs, ext, mod_time FROM nodes WHERE scan_id = ? AND path = ?`,
		scanID, path,
	)
	var n model.Node
	var isDir int
	if err := row.Scan(&n.Path, &n.Name, &isDir, &n.Size, &n.Files, &n.Dirs, &n.Ext, &n.ModTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	n.IsDir = isDir != 0
	return &n, nil
}

// NodeAt returns a single node by exact path within the latest scan.
func (s *Store) NodeAt(path string) (*model.Node, error) {
	scanID, _, err := s.latestScanRow()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s.nodeAt(scanID, path)
}

// Children returns the immediate children of path within the latest scan,
// sorted largest-first. If path is empty, the scan root's children are returned.
func (s *Store) Children(path string) ([]model.Node, error) {
	scanID, rootPath, err := s.latestScanRow()
	if err != nil {
		if err == sql.ErrNoRows {
			return []model.Node{}, nil
		}
		return nil, err
	}
	if path == "" {
		path = rootPath
	}
	return s.childrenByScan(scanID, path)
}

func (s *Store) childrenByScan(scanID int64, path string) ([]model.Node, error) {
	rows, err := s.db.Query(
		`SELECT path, name, is_dir, size, files, dirs, ext, mod_time FROM nodes
		 WHERE scan_id = ? AND parent_path = ? ORDER BY size DESC`,
		scanID, path,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.Node{}
	for rows.Next() {
		var n model.Node
		var isDir int
		if err := rows.Scan(&n.Path, &n.Name, &isDir, &n.Size, &n.Files, &n.Dirs, &n.Ext, &n.ModTime); err != nil {
			return nil, err
		}
		n.IsDir = isDir != 0
		out = append(out, n)
	}
	return out, rows.Err()
}

// Tree returns a nested subtree rooted at path, for treemap rendering.
// It greedily expands the largest not-yet-expanded directory first, so the
// biggest branches get the most visual detail, until maxNodes is reached -
// smaller branches are left as a single unexpanded box.
func (s *Store) Tree(path string, maxNodes int) (*model.Node, error) {
	scanID, rootPath, err := s.latestScanRow()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if path == "" {
		path = rootPath
	}
	root, err := s.nodeAt(scanID, path)
	if err != nil || root == nil {
		return root, err
	}

	count := 1
	frontier := []*model.Node{root}
	for len(frontier) > 0 && count < maxNodes {
		best := -1
		for i, n := range frontier {
			if n.IsDir && n.Dirs+n.Files > 0 && (best == -1 || n.Size > frontier[best].Size) {
				best = i
			}
		}
		if best == -1 {
			break
		}
		node := frontier[best]
		frontier = append(frontier[:best], frontier[best+1:]...)

		kids, err := s.childrenByScan(scanID, node.Path)
		if err != nil {
			return nil, err
		}
		node.Children = make([]*model.Node, len(kids))
		for i := range kids {
			k := kids[i]
			node.Children[i] = &k
			count++
			frontier = append(frontier, node.Children[i])
		}
	}
	return root, nil
}

// FileTypes returns size/count aggregated by extension for the latest scan.
func (s *Store) FileTypes(limit int) ([]model.FileTypeStat, error) {
	scanID, _, err := s.latestScanRow()
	if err != nil {
		if err == sql.ErrNoRows {
			return []model.FileTypeStat{}, nil
		}
		return nil, err
	}
	if limit <= 0 {
		limit = 30
	}
	rows, err := s.db.Query(
		`SELECT ext, SUM(size) AS total, COUNT(*) AS cnt FROM nodes
		 WHERE scan_id = ? AND is_dir = 0 GROUP BY ext ORDER BY total DESC LIMIT ?`,
		scanID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.FileTypeStat{}
	for rows.Next() {
		var fs model.FileTypeStat
		if err := rows.Scan(&fs.Ext, &fs.Size, &fs.Count); err != nil {
			return nil, err
		}
		out = append(out, fs)
	}
	return out, rows.Err()
}

// FileSearchParams configures SearchFiles.
type FileSearchParams struct {
	Query          string
	MatchPath      bool   // match Query against the full path instead of just the name
	FoldersOnly    bool   // show only directories, excluding files (default: files only)
	DuplicatesOnly bool   // only return files that share a duplicate key with another file
	DupMode        string // "name_size" (default) or "name_size_date"
	Limit          int
}

// SearchFiles does a flat, scan-wide file search (like WizTree's File View),
// with optional duplicate detection. Duplicates are matched only among files
// (never folders) by name+size, or name+size+modTime when DupMode is "name_size_date".
func (s *Store) SearchFiles(p FileSearchParams) ([]model.Node, error) {
	scanID, _, err := s.latestScanRow()
	if err != nil {
		if err == sql.ErrNoRows {
			return []model.Node{}, nil
		}
		return nil, err
	}

	limit := p.Limit
	if limit <= 0 || limit > 5000 {
		limit = 1000
	}

	dupCols := "name, size"
	dupJoin := "d.name = n.name AND d.size = n.size"
	if p.DupMode == "name_size_date" {
		dupCols = "name, size, mod_time"
		dupJoin += " AND d.mod_time = n.mod_time"
	}

	query := `
		SELECT n.path, n.name, n.is_dir, n.size, n.files, n.dirs, n.ext, n.mod_time, COALESCE(d.cnt, 0) AS dup_count
		FROM nodes n
		LEFT JOIN (
			SELECT ` + dupCols + `, COUNT(*) AS cnt
			FROM nodes
			WHERE scan_id = ? AND is_dir = 0
			GROUP BY ` + dupCols + `
			HAVING COUNT(*) > 1
		) d ON ` + dupJoin + `
		WHERE n.scan_id = ?`
	args := []any{scanID, scanID}

	if p.FoldersOnly {
		query += " AND n.is_dir = 1"
	} else {
		query += " AND n.is_dir = 0"
	}
	if p.DuplicatesOnly {
		query += " AND d.cnt IS NOT NULL"
	}
	if p.Query != "" {
		like := "%" + p.Query + "%"
		if p.MatchPath {
			query += " AND n.path LIKE ?"
		} else {
			query += " AND n.name LIKE ?"
		}
		args = append(args, like)
	}
	query += " ORDER BY n.size DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.Node{}
	for rows.Next() {
		var n model.Node
		var isDir int
		if err := rows.Scan(&n.Path, &n.Name, &isDir, &n.Size, &n.Files, &n.Dirs, &n.Ext, &n.ModTime, &n.DupCount); err != nil {
			return nil, err
		}
		n.IsDir = isDir != 0
		if n.DupCount > 0 {
			n.DupSize = n.Size * n.DupCount
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
