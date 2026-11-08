//go:build linux

package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

// A Fedora layout: / and /home are btrfs subvolumes of one partition; /boot is its own partition; a NAS is mounted
// at /mnt/nas; a USB drive's label contains a space.
const sampleMountinfo = `22 1 0:21 / /proc rw,nosuid,nodev,noexec,relatime shared:5 - proc proc rw
23 1 0:22 / /sys rw,nosuid,nodev,noexec,relatime shared:6 - sysfs sysfs rw
24 1 0:5 / /dev rw,nosuid shared:2 - devtmpfs devtmpfs rw,size=8000000k
44 1 0:36 /root / rw,relatime shared:1 - btrfs /dev/nvme0n1p3 rw,compress=zstd:1,subvol=/root
25 44 0:23 / /run rw,nosuid,nodev shared:12 - tmpfs tmpfs rw,mode=755
70 44 0:36 /home /home rw,relatime shared:40 - btrfs /dev/nvme0n1p3 rw,compress=zstd:1,subvol=/home
71 44 259:2 / /boot rw,relatime shared:42 - ext4 /dev/nvme0n1p2 rw
72 71 259:1 / /boot/efi rw,relatime shared:44 - vfat /dev/nvme0n1p1 rw,fmask=0077
80 44 0:50 / /mnt/nas rw,relatime shared:50 - cifs //nineball-nas.local/personal_folder rw,vers=3.1.1
81 25 8:17 / /run/media/me/USB\040Drive rw,nosuid,nodev shared:60 - exfat /dev/sdb1 rw
`

func TestParseMountinfo(t *testing.T) {
	mounts := parseMountinfo(sampleMountinfo)
	if len(mounts) != 10 {
		t.Fatalf("parsed %d mounts, want 10", len(mounts))
	}
	got := mounts[9]
	if got.point != "/run/media/me/USB Drive" || got.fsType != "exfat" || got.source != "/dev/sdb1" {
		t.Fatalf("escaped mount parsed as %+v", got)
	}
}

func TestMountsToSkipFromRoot(t *testing.T) {
	skip := mountsToSkip(parseMountinfo(sampleMountinfo), "/")
	for _, p := range []string{"/proc", "/sys", "/dev", "/run", "/boot", "/boot/efi", "/mnt/nas", "/run/media/me/USB Drive"} {
		if !skip[p] {
			t.Errorf("%s should be skipped when scanning /", p)
		}
	}
	// Same partition, another btrfs subvolume: part of the disk being scanned.
	if skip["/home"] {
		t.Errorf("/home is on the same disk and must be scanned")
	}
	if skip["/"] {
		t.Errorf("the scan root itself must never be skipped")
	}
}

func TestMountsToSkipFromHome(t *testing.T) {
	skip := mountsToSkip(parseMountinfo(sampleMountinfo), "/home")
	if len(skip) != 0 {
		t.Errorf("nothing is mounted inside /home, got %v", skip)
	}
}

func TestMountsToSkipFromNetworkShare(t *testing.T) {
	// Scanning a share itself works as before.
	skip := mountsToSkip(parseMountinfo(sampleMountinfo), "/mnt/nas")
	if len(skip) != 0 {
		t.Errorf("nothing is mounted inside the share, got %v", skip)
	}
}

func TestHardLinksCountedOnce(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "a.bin")
	if err := os.WriteFile(original, make([]byte, 1000), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(original, filepath.Join(dir, "b.bin")); err != nil {
		t.Skipf("hard links not supported here: %v", err)
	}
	res, err := Walk(dir)
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalSize != 1000 {
		t.Errorf("total size %d, want 1000 (the hard-linked copy is the same data)", res.TotalSize)
	}
	if res.TotalFiles != 2 {
		t.Errorf("total files %d, want 2 (both names are listed)", res.TotalFiles)
	}
}
