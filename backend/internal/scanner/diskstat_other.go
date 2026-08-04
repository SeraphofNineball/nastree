//go:build !linux && !windows

package scanner

// DiskUsage is unsupported on this build target; this exists so local
// `go vet`/editors on other platforms still type-check.
func DiskUsage(path string) (total, free uint64, err error) {
	return 0, 0, nil
}
