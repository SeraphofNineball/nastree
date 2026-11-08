//go:build linux

package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// boundary keeps a scan on the disk it started on. Scanning "/" would otherwise walk into /proc (whose kcore file
// alone reports the whole address space, ~128 TB), /sys, /run, and every other mounted disk and network share.
// Directories where something else is mounted are kept in the tree, empty, but not entered.
type boundary struct {
	skip map[string]bool
	seen map[fileID]bool // hard-linked files already counted
}

type fileID struct{ dev, ino uint64 }

type mount struct {
	point, fsType, source string
}

func newBoundary(rootPath string) *boundary {
	b := &boundary{seen: make(map[fileID]bool)}
	data, err := os.ReadFile("/proc/self/mountinfo")
	if err == nil {
		b.skip = mountsToSkip(parseMountinfo(string(data)), filepath.Clean(rootPath))
	}
	return b
}

// skipDir reports whether the walk should not enter the directory at path.
func (b *boundary) skipDir(path string) bool { return b.skip[path] }

// countSize returns the size to count for a file. A file with several hard links is counted once, at the first
// link the walk reaches, as `du` does.
func (b *boundary) countSize(info fs.FileInfo) int64 {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || st.Nlink <= 1 {
		return info.Size()
	}
	id := fileID{uint64(st.Dev), uint64(st.Ino)}
	if b.seen[id] {
		return 0
	}
	b.seen[id] = true
	return info.Size()
}

// parseMountinfo reads /proc/self/mountinfo lines:
//
//	36 35 98:0 /mnt1 /mnt/parent rw,noatime master:1 - ext3 /dev/root rw,errors=continue
//
// Field 5 is the mount point; after the "-" separator come the filesystem type and its source.
func parseMountinfo(text string) []mount {
	var out []mount
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}
		sep := -1
		for i := 6; i < len(fields); i++ {
			if fields[i] == "-" {
				sep = i
				break
			}
		}
		if sep < 0 || sep+2 >= len(fields) {
			continue
		}
		out = append(out, mount{
			point:  unescapeMount(fields[4]),
			fsType: fields[sep+1],
			source: unescapeMount(fields[sep+2]),
		})
	}
	return out
}

// unescapeMount decodes the kernel's octal escapes in mount paths ("\040" is a space).
func unescapeMount(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+3 < len(s) {
			if v, err := strconv.ParseUint(s[i+1:i+4], 8, 8); err == nil {
				b.WriteByte(byte(v))
				i += 3
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// mountsToSkip returns the mount points under root that belong to something other than root's own filesystem.
// Mounts of the same source and type (btrfs subvolumes such as /home on Fedora, or bind mounts of the same disk)
// count as the same disk and are scanned.
func mountsToSkip(mounts []mount, root string) map[string]bool {
	var home *mount
	for i := range mounts {
		m := &mounts[i]
		if isUnder(root, m.point) && (home == nil || len(m.point) >= len(home.point)) {
			home = m
		}
	}
	skip := make(map[string]bool)
	for _, m := range mounts {
		if m.point == root || !isUnder(m.point, root) {
			continue
		}
		if home != nil && m.source == home.source && m.fsType == home.fsType && strings.HasPrefix(m.source, "/") {
			continue
		}
		skip[m.point] = true
	}
	return skip
}

// isUnder reports whether path is dir or inside it.
func isUnder(path, dir string) bool {
	if dir == "/" {
		return strings.HasPrefix(path, "/")
	}
	return path == dir || strings.HasPrefix(path, dir+"/")
}
