package orchestrator

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/docker"
	"github.com/orbita-sh/orbita/internal/models"
)

func (o *Orchestrator) ProvisionDatabase(ctx context.Context, mdb *models.ManagedDatabase, orgSlug string) error {
	images, ok := models.EngineImages[mdb.Engine]
	if !ok {
		return fmt.Errorf("ProvisionDatabase: unsupported engine %s", mdb.Engine)
	}
	image, ok := images[mdb.Version]
	if !ok {
		return fmt.Errorf("ProvisionDatabase: unsupported version %s for engine %s", mdb.Version, mdb.Engine)
	}

	// Generate strong password
	password := generatePassword(32)

	// Volume name
	volumeName := fmt.Sprintf("%s_%s_data", orgSlug, mdb.Name)
	mdb.VolumeName = &volumeName

	// Build env vars based on engine
	envVars := getDBEnvVars(mdb.Engine, mdb.Name, password)

	// Default ports
	port := getDefaultPort(mdb.Engine)
	mdb.Port = &port

	spec := docker.ServiceSpec{
		Name:         fmt.Sprintf("orbita-db-%s", mdb.ID.String()[:8]),
		Image:        image,
		Replicas:     1,
		Port:         port,
		EnvVars:      envVars,
		NetworkName:  docker.GetOrgNetworkName(orgSlug),
		CgroupParent: fmt.Sprintf("orbita-org-%s", orgSlug),
		Labels: map[string]string{
			"orbita.db.id":   mdb.ID.String(),
			"orbita.org":     orgSlug,
			"orbita.managed": "true",
		},
	}

	// Pull image
	reader, err := o.dockerClient.PullImage(ctx, image, "")
	if err != nil {
		return fmt.Errorf("ProvisionDatabase: pull: %w", err)
	}
	defer reader.Close()

	// Create service
	serviceID, err := o.dockerClient.CreateService(ctx, spec)
	if err != nil {
		return fmt.Errorf("ProvisionDatabase: create service: %w", err)
	}

	mdb.DockerServiceID = &serviceID
	mdb.Status = models.DBStatusRunning

	// Build connection string
	connStr := buildConnectionString(mdb.Engine, mdb.Name, password, port, orgSlug)
	mdb.ConnectionConfig = &connStr

	log.Info().
		Str("engine", mdb.Engine).
		Str("version", mdb.Version).
		Str("name", mdb.Name).
		Msg("Database provisioned")

	return nil
}

func (o *Orchestrator) RemoveDatabase(ctx context.Context, mdb *models.ManagedDatabase) error {
	if mdb.DockerServiceID != nil {
		if err := o.dockerClient.RemoveService(ctx, *mdb.DockerServiceID); err != nil {
			return fmt.Errorf("RemoveDatabase: %w", err)
		}
	}
	return nil
}

func (o *Orchestrator) RestartDatabase(ctx context.Context, mdb *models.ManagedDatabase) error {
	if mdb.DockerServiceID == nil {
		return fmt.Errorf("RestartDatabase: no service ID")
	}
	_ = o.dockerClient.ScaleService(ctx, *mdb.DockerServiceID, 0)
	return o.dockerClient.ScaleService(ctx, *mdb.DockerServiceID, 1)
}

func (o *Orchestrator) StopDatabase(ctx context.Context, mdb *models.ManagedDatabase) error {
	if mdb.DockerServiceID == nil {
		return nil
	}
	return o.dockerClient.ScaleService(ctx, *mdb.DockerServiceID, 0)
}

func (o *Orchestrator) StartDatabase(ctx context.Context, mdb *models.ManagedDatabase) error {
	if mdb.DockerServiceID == nil {
		return fmt.Errorf("StartDatabase: no service ID")
	}
	return o.dockerClient.ScaleService(ctx, *mdb.DockerServiceID, 1)
}

func generatePassword(length int) string {
	b := make([]byte, length/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func getDBEnvVars(engine, dbName, password string) map[string]string {
	switch engine {
	case models.EnginePostgres:
		return map[string]string{"POSTGRES_DB": dbName, "POSTGRES_USER": dbName, "POSTGRES_PASSWORD": password}
	case models.EngineMySQL:
		return map[string]string{"MYSQL_DATABASE": dbName, "MYSQL_USER": dbName, "MYSQL_PASSWORD": password, "MYSQL_ROOT_PASSWORD": password}
	case models.EngineMariaDB:
		return map[string]string{"MARIADB_DATABASE": dbName, "MARIADB_USER": dbName, "MARIADB_PASSWORD": password, "MARIADB_ROOT_PASSWORD": password}
	case models.EngineMongoDB:
		return map[string]string{"MONGO_INITDB_ROOT_USERNAME": dbName, "MONGO_INITDB_ROOT_PASSWORD": password}
	case models.EngineRedis:
		return map[string]string{}
	default:
		return map[string]string{}
	}
}

func getDefaultPort(engine string) int {
	switch engine {
	case models.EnginePostgres:
		return 5432
	case models.EngineMySQL, models.EngineMariaDB:
		return 3306
	case models.EngineMongoDB:
		return 27017
	case models.EngineRedis:
		return 6379
	default:
		return 0
	}
}

func buildConnectionString(engine, dbName, password string, port int, orgSlug string) string {
	host := fmt.Sprintf("orbita-db-%s", dbName)
	switch engine {
	case models.EnginePostgres:
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", dbName, password, host, port, dbName)
	case models.EngineMySQL, models.EngineMariaDB:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbName, password, host, port, dbName)
	case models.EngineMongoDB:
		return fmt.Sprintf("mongodb://%s:%s@%s:%d", dbName, password, host, port)
	case models.EngineRedis:
		return fmt.Sprintf("redis://%s:%d", host, port)
	default:
		return ""
	}
}
