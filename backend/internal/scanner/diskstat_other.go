//go:build !linux

package scanner

// DiskUsage is unsupported on non-Linux build targets; the app is only
// ever built and run for linux (Docker), this exists so local `go vet`/
// editors on other platforms still type-check.
func DiskUsage(path string) (total, free uint64, err error) {
	return 0, 0, nil
}
