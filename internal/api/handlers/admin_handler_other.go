//go:build !linux

package handlers

// diskGBRoot is a no-op on non-Linux builds (Windows/Mac dev machines). The
// release binary always runs on Linux where the real implementation is used.
func diskGBRoot() int {
	return 0
}
