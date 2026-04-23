//go:build linux

package handlers

import "syscall"

// diskGBRoot returns the total size (in GB) of the filesystem mounted at "/".
// Inside a Docker container this reports the container's rootfs — typically
// backed by the host's /var/lib/docker overlay, which is a reasonable proxy
// for "how much disk does this host have".
func diskGBRoot() int {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return 0
	}
	total := stat.Blocks * uint64(stat.Bsize)
	return int(total / (1024 * 1024 * 1024))
}
