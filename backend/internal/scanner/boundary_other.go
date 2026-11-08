//go:build !linux

package scanner

import "io/fs"

// boundary keeps a scan on the disk it started on. Windows volumes are separate drive letters already, so there is
// nothing to skip there; macOS is not handled yet.
type boundary struct{}

func newBoundary(rootPath string) *boundary { return &boundary{} }

func (b *boundary) skipDir(path string) bool { return false }

func (b *boundary) countSize(info fs.FileInfo) int64 { return info.Size() }
