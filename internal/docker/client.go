package docker

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
)

// Client wraps Docker SDK operations.
// TODO: real impl with github.com/docker/docker/client
type Client struct {
	socketPath string
}

func NewClient(socketPath string) *Client {
	return &Client{socketPath: socketPath}
}

type ServiceSpec struct {
	Name         string
	Image        string
	Replicas     int
	Port         int
	EnvVars      map[string]string
	Labels       map[string]string
	NetworkName  string
	CgroupParent string
	CPULimit     int64 // nanoCPUs
	MemoryLimit  int64 // bytes
	RestartPolicy string
}

type ServiceInfo struct {
	ID       string
	Name     string
	Image    string
	Replicas int
	Status   string
}

func (c *Client) PullImage(ctx context.Context, imageRef, registryAuth string) (io.ReadCloser, error) {
	// TODO: real impl — docker.ImagePull()
	log.Info().Str("image", imageRef).Msg("Pulling image (stub)")
	return io.NopCloser(strings.NewReader(fmt.Sprintf("Pulling %s...\nDone.\n", imageRef))), nil
}

func (c *Client) CreateService(ctx context.Context, spec ServiceSpec) (string, error) {
	// TODO: real impl — docker.ServiceCreate()
	serviceID := fmt.Sprintf("svc-%s", spec.Name)
	log.Info().Str("service", spec.Name).Str("image", spec.Image).Msg("Created service (stub)")
	return serviceID, nil
}

func (c *Client) UpdateService(ctx context.Context, serviceID string, spec ServiceSpec) error {
	// TODO: real impl — docker.ServiceUpdate()
	log.Info().Str("service_id", serviceID).Str("image", spec.Image).Msg("Updated service (stub)")
	return nil
}

func (c *Client) RemoveService(ctx context.Context, serviceID string) error {
	// TODO: real impl — docker.ServiceRemove()
	log.Info().Str("service_id", serviceID).Msg("Removed service (stub)")
	return nil
}

func (c *Client) ScaleService(ctx context.Context, serviceID string, replicas int) error {
	// TODO: real impl — docker.ServiceUpdate() with replica count
	log.Info().Str("service_id", serviceID).Int("replicas", replicas).Msg("Scaled service (stub)")
	return nil
}

func (c *Client) GetServiceLogs(ctx context.Context, serviceID string, tail int) (io.ReadCloser, error) {
	// TODO: real impl — docker.ServiceLogs()
	mockLogs := fmt.Sprintf("[%s] Application started on port 3000\n[%s] Ready to accept connections\n", serviceID, serviceID)
	return io.NopCloser(strings.NewReader(mockLogs)), nil
}

func (c *Client) InspectService(ctx context.Context, serviceID string) (*ServiceInfo, error) {
	// TODO: real impl — docker.ServiceInspectWithRaw()
	return &ServiceInfo{
		ID:       serviceID,
		Name:     serviceID,
		Replicas: 1,
		Status:   "running",
	}, nil
}

func (c *Client) GetContainerStats(ctx context.Context, containerID string) (map[string]interface{}, error) {
	// TODO: real impl — docker.ContainerStats()
	return map[string]interface{}{
		"cpu_percent":    2.5,
		"memory_usage":   67108864,
		"memory_limit":   134217728,
		"network_rx":     1024000,
		"network_tx":     512000,
	}, nil
}
