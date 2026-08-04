//go:build windows

package scanner

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

// DiskUsage returns total and free bytes for the volume containing path.
// GetDiskFreeSpaceEx requires a volume root (e.g. "C:\"), not an arbitrary
// subpath, so we derive the volume name first.
func DiskUsage(path string) (total, free uint64, err error) {
	volume := filepath.VolumeName(path) + `\`

	volumePtr, err := windows.UTF16PtrFromString(volume)
	if err != nil {
		return 0, 0, err
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	if err := windows.GetDiskFreeSpaceEx(volumePtr, &freeBytesAvailable, &totalBytes, &totalFreeBytes); err != nil {
		return 0, 0, err
	}

	return totalBytes, freeBytesAvailable, nil
}
