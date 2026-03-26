package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/orbita-sh/orbita/internal/middleware"
	"github.com/orbita-sh/orbita/internal/repository"
	"github.com/orbita-sh/orbita/internal/response"
)

type DashboardHandler struct {
	appRepo    *repository.AppRepository
	dbRepo     *repository.DBRepository
	cronRepo   *repository.CronRepository
	deployRepo *repository.AppRepository
}

func NewDashboardHandler(appRepo *repository.AppRepository, dbRepo *repository.DBRepository, cronRepo *repository.CronRepository) *DashboardHandler {
	return &DashboardHandler{
		appRepo:    appRepo,
		dbRepo:     dbRepo,
		cronRepo:   cronRepo,
		deployRepo: appRepo,
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
		"total_apps":       len(apps),
		"running_apps":     runningApps,
		"stopped_apps":     stoppedApps,
		"total_databases":  len(dbs),
		"running_databases": runningDBs,
		"total_cron_jobs":  len(cronJobs),
		"active_cron_jobs": activeCrons,
		"recent_deploys":   recentDeploys,
	})
}

func (h *DashboardHandler) GetMetricsOverview(c *gin.Context) {
	// TODO: real impl — aggregate resource usage across all org containers
	response.Success(c, http.StatusOK, gin.H{
		"cpu_percent":    12.5,
		"memory_used":    536870912,
		"memory_total":   2147483648,
		"disk_used":      5368709120,
		"disk_total":     21474836480,
		"network_rx":     10240000,
		"network_tx":     5120000,
	})
}
