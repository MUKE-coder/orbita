package orchestrator

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/docker"
)

// drainAndClose reads a stream to EOF and closes it — used to complete image
// pulls before the pulled image is run.
func drainAndClose(r io.ReadCloser) {
	_, _ = io.Copy(io.Discard, r)
	_ = r.Close()
}

// MinioResult carries the connection details of a provisioned MinIO addon.
type MinioResult struct {
	ServiceID   string
	ServiceName string
	Endpoint    string // http://<service>:9000 (in-network)
	AccessKey   string
	SecretKey   string
}

// ProvisionMinio creates a MinIO object-storage service in the org's network,
// reachable only in-network (no host binding). Idempotent by service name is the
// caller's responsibility; this always creates. Used for the Grit `minio` addon.
func (o *Orchestrator) ProvisionMinio(ctx context.Context, orgSlug, name string) (*MinioResult, error) {
	serviceName := fmt.Sprintf("orbita-minio-%s", name)
	volumeName := fmt.Sprintf("%s_%s_minio", orgSlug, name)

	// Idempotent: if the service already exists, read its generated credentials
	// back instead of provisioning a second one.
	if env, ok, err := o.dockerClient.FindServiceEnvByName(ctx, serviceName); err == nil && ok {
		return &MinioResult{
			ServiceName: serviceName,
			Endpoint:    fmt.Sprintf("http://%s:9000", serviceName),
			AccessKey:   env["MINIO_ROOT_USER"],
			SecretKey:   env["MINIO_ROOT_PASSWORD"],
		}, nil
	}

	accessKey := generatePassword(20)
	secretKey := generatePassword(40)

	spec := docker.ServiceSpec{
		Name:     serviceName,
		Image:    "minio/minio",
		Replicas: 1,
		// Not published: reachable only by containers on the org network.
		EnvVars: map[string]string{
			"MINIO_ROOT_USER":     accessKey,
			"MINIO_ROOT_PASSWORD": secretKey,
		},
		Command:     []string{"server", "/data", "--console-address", ":9001"},
		NetworkName: docker.GetOrgNetworkName(orgSlug),
		Labels: map[string]string{
			"orbita.minio":   name,
			"orbita.org":     orgSlug,
			"orbita.managed": "true",
		},
		Mounts: []docker.VolumeMount{{Source: volumeName, Target: "/data"}},
	}

	if reader, err := o.dockerClient.PullImage(ctx, "minio/minio", ""); err == nil {
		drainAndClose(reader)
	}

	serviceID, err := o.dockerClient.CreateService(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("ProvisionMinio: %w", err)
	}

	log.Info().Str("service", serviceName).Msg("MinIO addon provisioned")
	return &MinioResult{
		ServiceID:   serviceID,
		ServiceName: serviceName,
		Endpoint:    fmt.Sprintf("http://%s:9000", serviceName),
		AccessKey:   accessKey,
		SecretKey:   secretKey,
	}, nil
}
