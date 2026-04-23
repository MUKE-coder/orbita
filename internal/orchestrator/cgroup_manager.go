package orchestrator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog/log"
)

// CgroupManager creates per-org cgroup v2 slices and writes memory/CPU limits.
// All operations gracefully no-op when /sys/fs/cgroup isn't writable — this is
// the expected state when Orbita runs without privileged mode, or on cgroup v1
// hosts. A single warning is logged and deploys continue using Docker's per-
// container limits instead of the slice-level cap.
type CgroupManager struct {
	// cgroupRoot is the cgroup v2 mount point. Defaults to /sys/fs/cgroup.
	cgroupRoot string
	// enforcementEnabled is determined at init based on whether we can write.
	enforcementEnabled bool
}

// NewCgroupManager detects whether cgroup v2 is writable and ready to use.
func NewCgroupManager() *CgroupManager {
	root := os.Getenv("CGROUP_ROOT")
	if root == "" {
		root = "/sys/fs/cgroup"
	}

	m := &CgroupManager{cgroupRoot: root}

	// Probe: cgroup v2 uses a single hierarchy; a reliable marker is the file
	// `cgroup.controllers` at the root.
	controllersPath := filepath.Join(root, "cgroup.controllers")
	if _, err := os.Stat(controllersPath); err != nil {
		log.Warn().Err(err).Str("root", root).Msg("cgroup v2 not detected; per-org resource enforcement disabled")
		return m
	}

	// Can we write? Try creating a test directory inside /sys/fs/cgroup/orbita.slice.
	testDir := filepath.Join(root, "orbita.slice")
	if err := os.MkdirAll(testDir, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		log.Warn().Err(err).Str("path", testDir).Msg("cannot write to cgroup v2; per-org enforcement disabled (run container with --privileged + cgroup mount rw)")
		return m
	}

	// Enable +memory +cpu in the parent so children inherit those controllers.
	_ = writeFile(filepath.Join(root, "cgroup.subtree_control"), []byte("+memory +cpu"))
	_ = writeFile(filepath.Join(testDir, "cgroup.subtree_control"), []byte("+memory +cpu"))

	m.enforcementEnabled = true
	log.Info().Str("root", root).Msg("cgroup v2 enforcement enabled")
	return m
}

// GetCgroupParent returns the Docker cgroup-parent path for an org.
// Docker runs containers under this slice, and kernel enforces the limits we set.
func GetCgroupParent(orgSlug string) string {
	return fmt.Sprintf("orbita.slice/orbita-org-%s.slice", orgSlug)
}

// Enforced reports whether cgroup writes will actually take effect.
func (m *CgroupManager) Enforced() bool {
	return m.enforcementEnabled
}

// EnsureOrgSlice creates (or updates) an org's cgroup slice with the given
// CPU/RAM caps. cpuCores = 0 means unlimited; ramMB = 0 means unlimited.
func (m *CgroupManager) EnsureOrgSlice(orgSlug string, cpuCores, ramMB int) error {
	if !m.enforcementEnabled {
		log.Debug().Str("org", orgSlug).Msg("cgroup enforcement disabled, skipping EnsureOrgSlice")
		return nil
	}

	slice := m.sliceDir(orgSlug)
	if err := os.MkdirAll(slice, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("EnsureOrgSlice: mkdir: %w", err)
	}

	// Enable controllers in this slice so children can be constrained.
	_ = writeFile(filepath.Join(slice, "cgroup.subtree_control"), []byte("+memory +cpu"))

	return m.applyLimits(slice, cpuCores, ramMB, orgSlug)
}

// UpdateOrgSlice changes the limits on an existing slice.
func (m *CgroupManager) UpdateOrgSlice(orgSlug string, cpuCores, ramMB int) error {
	if !m.enforcementEnabled {
		return nil
	}

	slice := m.sliceDir(orgSlug)
	if _, err := os.Stat(slice); err != nil {
		// Slice doesn't exist — create it.
		return m.EnsureOrgSlice(orgSlug, cpuCores, ramMB)
	}
	return m.applyLimits(slice, cpuCores, ramMB, orgSlug)
}

// RemoveOrgSlice removes an org's cgroup slice. Docker will complain if
// containers are still running under it; stop those first.
func (m *CgroupManager) RemoveOrgSlice(orgSlug string) error {
	if !m.enforcementEnabled {
		return nil
	}

	slice := m.sliceDir(orgSlug)
	if err := os.Remove(slice); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("RemoveOrgSlice: %w", err)
	}
	log.Info().Str("org", orgSlug).Msg("Removed cgroup slice")
	return nil
}

func (m *CgroupManager) sliceDir(orgSlug string) string {
	return filepath.Join(m.cgroupRoot, "orbita.slice", fmt.Sprintf("orbita-org-%s.slice", orgSlug))
}

// applyLimits writes memory.max and cpu.max for a slice. Uses "max" (cgroup
// v2 keyword) when the limit is 0.
func (m *CgroupManager) applyLimits(slice string, cpuCores, ramMB int, orgSlug string) error {
	// memory.max (bytes or "max")
	memVal := "max"
	if ramMB > 0 {
		memVal = strconv.Itoa(ramMB * 1024 * 1024)
	}
	if err := writeFile(filepath.Join(slice, "memory.max"), []byte(memVal)); err != nil {
		log.Warn().Err(err).Str("org", orgSlug).Msg("failed to write memory.max")
	}

	// cpu.max: "<quota> <period>". 1 core = 100000/100000. 2 cores = 200000/100000.
	cpuVal := "max 100000"
	if cpuCores > 0 {
		quota := cpuCores * 100000
		cpuVal = fmt.Sprintf("%d 100000", quota)
	}
	if err := writeFile(filepath.Join(slice, "cpu.max"), []byte(cpuVal)); err != nil {
		log.Warn().Err(err).Str("org", orgSlug).Msg("failed to write cpu.max")
	}

	log.Info().
		Str("org", orgSlug).
		Int("cpu_cores", cpuCores).
		Int("ram_mb", ramMB).
		Msg("Applied cgroup limits")
	return nil
}

// writeFile writes data to an existing cgroup control file. Unlike
// os.WriteFile, we don't create the file — cgroup controls must already exist.
func writeFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}
