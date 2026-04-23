package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/orbita-sh/orbita/internal/docker"
	"github.com/orbita-sh/orbita/internal/middleware"
	"github.com/orbita-sh/orbita/internal/repository"
	"github.com/orbita-sh/orbita/internal/response"
)

type DashboardHandler struct {
	appRepo      *repository.AppRepository
	dbRepo       *repository.DBRepository
	cronRepo     *repository.CronRepository
	deployRepo   *repository.AppRepository
	dockerClient *docker.Client
}

func NewDashboardHandler(appRepo *repository.AppRepository, dbRepo *repository.DBRepository, cronRepo *repository.CronRepository, dockerClient *docker.Client) *DashboardHandler {
	return &DashboardHandler{
		appRepo:      appRepo,
		dbRepo:       dbRepo,
		cronRepo:     cronRepo,
		deployRepo:   appRepo,
		dockerClient: dockerClient,
	}
}

func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	ctx := c.Request.Context()

	apps, _ := h.appRepo.ListByOrgID(ctx, orgID)
	dbs, _ := h.dbRepo.ListByOrgID(ctx, orgID)
	cronJobs, _ := h.cronRepo.ListByOrgID(ctx, orgID)

	// Count by status
	runningApps := 0
	stoppedApps := 0
	for _, a := range apps {
		switch a.Status {
		case "running":
			runningApps++
		case "stopped":
			stoppedApps++
		}
	}

	runningDBs := 0
	for _, d := range dbs {
		if d.Status == "running" {
			runningDBs++
		}
	}

	activeCrons := 0
	for _, j := range cronJobs {
		if j.Enabled {
			activeCrons++
		}
	}

	// Recent deployments
	var recentDeploys []map[string]interface{}
	for _, a := range apps {
		deploys, _ := h.deployRepo.ListDeployments(ctx, a.ID, 3)
		for _, d := range deploys {
			recentDeploys = append(recentDeploys, map[string]interface{}{
				"id":         d.ID,
				"app_name":   a.Name,
				"version":    d.Version,
				"status":     d.Status,
				"started_at": d.StartedAt,
				"image_ref":  d.ImageRef,
			})
		}
	}

	// Limit to 10 most recent
	if len(recentDeploys) > 10 {
		recentDeploys = recentDeploys[:10]
	}

	response.Success(c, http.StatusOK, gin.H{
		"total_apps":        len(apps),
		"running_apps":      runningApps,
		"stopped_apps":      stoppedApps,
		"total_databases":   len(dbs),
		"running_databases": runningDBs,
		"total_cron_jobs":   len(cronJobs),
		"active_cron_jobs":  activeCrons,
		"recent_deploys":    recentDeploys,
	})
}

// GetMetricsOverview returns the org's quota as the "total" (source of truth
// for billing + planning) and actual container usage as the "used" number.
func (h *DashboardHandler) GetMetricsOverview(c *gin.Context) {
	org := middleware.GetOrgFromContext(c)
	if org == nil {
		response.InternalError(c, "Missing org context")
		return
	}
	ctx := c.Request.Context()

	totalMem := int64(org.EffectiveRAMMB()) * 1024 * 1024
	totalDisk := int64(org.EffectiveDiskGB()) * 1024 * 1024 * 1024
	cpuCores := org.EffectiveCPUCores()

	// Aggregate real container stats across this org's apps
	var usedMem int64
	var cpuSum float64
	var rxSum, txSum uint64
	appsSeen := 0

	apps, _ := h.appRepo.ListByOrgID(ctx, org.ID)
	for _, app := range apps {
		if app.DockerServiceID == nil || *app.DockerServiceID == "" {
			continue
		}
		appsSeen++
		// Best-effort: fetch stats with a short timeout so a slow daemon
		// doesn't stall the whole dashboard.
		statsCtx, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
		stats, err := h.dockerClient.GetContainerStats(statsCtx, *app.DockerServiceID)
		cancel()
		if err != nil {
			continue
		}
		if v, ok := stats["memory_usage"].(uint64); ok {
			usedMem += int64(v)
		}
		if v, ok := stats["cpu_percent"].(float64); ok {
			cpuSum += v
		}
		if v, ok := stats["network_rx"].(uint64); ok {
			rxSum += v
		}
		if v, ok := stats["network_tx"].(uint64); ok {
			txSum += v
		}
	}

	// Normalize CPU percent against the org's CPU allotment (0 - 100)
	cpuPct := 0.0
	if cpuCores > 0 {
		cpuPct = cpuSum / float64(cpuCores)
		if cpuPct > 100 {
			cpuPct = 100
		}
	}

	response.Success(c, http.StatusOK, gin.H{
		"cpu_percent":  cpuPct,
		"cpu_cores":    cpuCores,
		"memory_used":  usedMem,
		"memory_total": totalMem,
		"disk_used":    0, // per-container disk usage isn't tracked yet
		"disk_total":   totalDisk,
		"network_rx":   rxSum,
		"network_tx":   txSum,
	})
}
