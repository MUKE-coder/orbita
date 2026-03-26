package orchestrator

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/models"
)

type CgroupManager struct{}

func NewCgroupManager() *CgroupManager {
	return &CgroupManager{}
}

func (m *CgroupManager) EnsureOrgSlice(orgSlug string, plan *models.ResourcePlan) error {
	// TODO: real impl
	// 1. Create /sys/fs/cgroup/orbita.slice/orbita-org-{slug}.slice/
	// 2. Write memory.max = plan.MaxRAMMB * 1024 * 1024
	// 3. Write cpu.weight = proportional shares
	// 4. Write cgroup.subtree_control = "+cpu +memory"

	slicePath := fmt.Sprintf("orbita.slice/orbita-org-%s.slice", orgSlug)
	log.Info().
		Str("slice", slicePath).
		Int("max_ram_mb", plan.MaxRAMMB).
		Int("max_cpu_cores", plan.MaxCPUCores).
		Msg("Cgroup slice ensured (stub)")

	return nil
}

func (m *CgroupManager) UpdateOrgSlice(orgSlug string, plan *models.ResourcePlan) error {
	// TODO: real impl — update limits in place
	log.Info().Str("org", orgSlug).Msg("Cgroup slice updated (stub)")
	return nil
}

func (m *CgroupManager) RemoveOrgSlice(orgSlug string) error {
	// TODO: real impl — remove cgroup directory
	log.Info().Str("org", orgSlug).Msg("Cgroup slice removed (stub)")
	return nil
}

func GetCgroupParent(orgSlug string) string {
	return fmt.Sprintf("orbita.slice/orbita-org-%s.slice", orgSlug)
}
