package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	filtertypes "github.com/docker/docker/api/types/filters"
	imagetypes "github.com/docker/docker/api/types/image"
	networktypes "github.com/docker/docker/api/types/network"
	registrytypes "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/swarm"
	dockerclient "github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
)

// Client wraps Docker SDK operations: image pulls, swarm services, network
// management, and container stats. Initialized with a socket path pointing to
// the Docker daemon (typically /var/run/docker.sock mounted into the container).
type Client struct {
	cli        *dockerclient.Client
	socketPath string
}

// NewClient initializes a Docker SDK client with API-version negotiation.
// Version negotiation ensures the client adapts to the daemon's API version,
// avoiding the "client version too old" error on Docker Engine 28+.
func NewClient(socketPath string) *Client {
	host := socketPath
	if !strings.Contains(host, "://") {
		host = "unix://" + host
	}

	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.WithHost(host),
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Error().Err(err).Str("socket", socketPath).Msg("Failed to initialize Docker client; falling back to stub")
		return &Client{socketPath: socketPath}
	}

	return &Client{cli: cli, socketPath: socketPath}
}

// Close releases the underlying HTTP client.
func (c *Client) Close() error {
	if c.cli == nil {
		return nil
	}
	return c.cli.Close()
}

// ping returns an error if the Docker daemon isn't reachable.
func (c *Client) ping(ctx context.Context) error {
	if c.cli == nil {
		return fmt.Errorf("docker client not initialized")
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := c.cli.Ping(pingCtx)
	return err
}

// ---------- public types (same contract as before) ----------

type ServiceSpec struct {
	Name          string
	Image         string
	Replicas      int
	Port          int
	EnvVars       map[string]string
	Labels        map[string]string
	NetworkName   string
	CgroupParent  string // informational only; Swarm services use the daemon's cgroup hierarchy
	CPULimit      int64  // nanoCPUs (e.g. 1e9 = 1 core). 0 = unlimited.
	MemoryLimit   int64  // bytes. 0 = unlimited.
	RestartPolicy string // "any" (default) | "on-failure" | "none"
}

type ServiceInfo struct {
	ID       string
	Name     string
	Image    string
	Replicas int
	Status   string
}

// ---------- image operations ----------

// PullImage pulls an image from a registry. For public images pass empty
// registryAuth. For private images pass a base64-encoded JSON:
//   {"username":"x","password":"y","serveraddress":"https://index.docker.io/v1/"}
// Returns a reader of the daemon's JSON progress stream — caller must
// Close() it even if ignoring content.
func (c *Client) PullImage(ctx context.Context, imageRef, registryAuth string) (io.ReadCloser, error) {
	if c.cli == nil {
		return nil, fmt.Errorf("PullImage: client not initialized")
	}
	log.Info().Str("image", imageRef).Msg("Pulling image")

	reader, err := c.cli.ImagePull(ctx, imageRef, imagetypes.PullOptions{
		RegistryAuth: registryAuth,
	})
	if err != nil {
		return nil, fmt.Errorf("PullImage: %w", err)
	}
	return reader, nil
}

// BuildImage builds an image from a remote git URL. Docker clones the repo
// internally. For private repos embed a PAT in the URL:
//   https://<token>@github.com/user/repo.git#branch
// dockerfile is relative to the repo root (e.g., "Dockerfile" or "backend/Dockerfile").
// Returns a reader of the daemon's JSON build stream — caller must Close() it.
func (c *Client) BuildImage(ctx context.Context, remoteURL, tag, dockerfile string, buildArgs map[string]string, registryAuth map[string]registrytypes.AuthConfig) (io.ReadCloser, error) {
	if c.cli == nil {
		return nil, fmt.Errorf("BuildImage: client not initialized")
	}
	log.Info().Str("remote", sanitizeGitURL(remoteURL)).Str("tag", tag).Msg("Building image")

	// Convert map[string]string to map[string]*string (Docker's build args format)
	args := make(map[string]*string, len(buildArgs))
	for k, v := range buildArgs {
		vv := v
		args[k] = &vv
	}

	opts := dockertypes.ImageBuildOptions{
		RemoteContext: remoteURL,
		Dockerfile:    dockerfile,
		Tags:          []string{tag},
		BuildArgs:     args,
		Remove:        true,
		ForceRemove:   true,
		PullParent:    true,
		AuthConfigs:   registryAuth,
	}

	resp, err := c.cli.ImageBuild(ctx, nil, opts)
	if err != nil {
		return nil, fmt.Errorf("BuildImage: %w", err)
	}
	return resp.Body, nil
}

// sanitizeGitURL strips the token from a git URL for logging.
func sanitizeGitURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}
	u.User = url.User("***")
	return u.String()
}

// EncodeRegistryAuth produces the base64-encoded auth string PullImage expects.
func EncodeRegistryAuth(username, password, serverAddress string) (string, error) {
	cfg := registrytypes.AuthConfig{
		Username:      username,
		Password:      password,
		ServerAddress: serverAddress,
	}
	buf, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}

// ---------- swarm service operations ----------

// CreateService creates a Swarm service. Requires Swarm mode initialized on
// the daemon (`docker swarm init`). Resolves NetworkName → ID internally.
func (c *Client) CreateService(ctx context.Context, spec ServiceSpec) (string, error) {
	if c.cli == nil {
		return "", fmt.Errorf("CreateService: client not initialized")
	}

	swarmSpec, err := c.buildSwarmServiceSpec(ctx, spec)
	if err != nil {
		return "", fmt.Errorf("CreateService: %w", err)
	}

	resp, err := c.cli.ServiceCreate(ctx, swarmSpec, dockertypes.ServiceCreateOptions{})
	if err != nil {
		return "", fmt.Errorf("CreateService: %w", err)
	}

	log.Info().
		Str("service_id", resp.ID).
		Str("name", spec.Name).
		Str("image", spec.Image).
		Msg("Created swarm service")
	return resp.ID, nil
}

// UpdateService updates an existing service by re-applying the spec. Uses the
// service's current index as the optimistic-concurrency version.
func (c *Client) UpdateService(ctx context.Context, serviceID string, spec ServiceSpec) error {
	if c.cli == nil {
		return fmt.Errorf("UpdateService: client not initialized")
	}

	current, _, err := c.cli.ServiceInspectWithRaw(ctx, serviceID, dockertypes.ServiceInspectOptions{})
	if err != nil {
		return fmt.Errorf("UpdateService: inspect: %w", err)
	}

	swarmSpec, err := c.buildSwarmServiceSpec(ctx, spec)
	if err != nil {
		return fmt.Errorf("UpdateService: spec: %w", err)
	}

	_, err = c.cli.ServiceUpdate(ctx, serviceID, current.Version, swarmSpec, dockertypes.ServiceUpdateOptions{})
	if err != nil {
		return fmt.Errorf("UpdateService: %w", err)
	}
	log.Info().Str("service_id", serviceID).Str("image", spec.Image).Msg("Updated swarm service")
	return nil
}

// RemoveService removes a service. Swarm takes care of stopping tasks.
func (c *Client) RemoveService(ctx context.Context, serviceID string) error {
	if c.cli == nil {
		return fmt.Errorf("RemoveService: client not initialized")
	}
	if err := c.cli.ServiceRemove(ctx, serviceID); err != nil {
		return fmt.Errorf("RemoveService: %w", err)
	}
	log.Info().Str("service_id", serviceID).Msg("Removed swarm service")
	return nil
}

// ScaleService updates the desired replica count without touching the rest of the spec.
func (c *Client) ScaleService(ctx context.Context, serviceID string, replicas int) error {
	if c.cli == nil {
		return fmt.Errorf("ScaleService: client not initialized")
	}

	current, _, err := c.cli.ServiceInspectWithRaw(ctx, serviceID, dockertypes.ServiceInspectOptions{})
	if err != nil {
		return fmt.Errorf("ScaleService: inspect: %w", err)
	}

	count := uint64(replicas)
	if current.Spec.Mode.Replicated == nil {
		current.Spec.Mode.Replicated = &swarm.ReplicatedService{}
	}
	current.Spec.Mode.Replicated.Replicas = &count

	_, err = c.cli.ServiceUpdate(ctx, serviceID, current.Version, current.Spec, dockertypes.ServiceUpdateOptions{})
	if err != nil {
		return fmt.Errorf("ScaleService: %w", err)
	}
	log.Info().Str("service_id", serviceID).Int("replicas", replicas).Msg("Scaled service")
	return nil
}

// GetServiceLogs streams logs from all tasks of a service. If tail < 0, returns
// all logs; otherwise last N lines. Caller must Close() the reader.
func (c *Client) GetServiceLogs(ctx context.Context, serviceID string, tail int) (io.ReadCloser, error) {
	if c.cli == nil {
		return nil, fmt.Errorf("GetServiceLogs: client not initialized")
	}

	tailStr := "all"
	if tail > 0 {
		tailStr = fmt.Sprintf("%d", tail)
	}

	return c.cli.ServiceLogs(ctx, serviceID, containertypes.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       tailStr,
	})
}

// InspectService returns a simplified view of a service.
func (c *Client) InspectService(ctx context.Context, serviceID string) (*ServiceInfo, error) {
	if c.cli == nil {
		return nil, fmt.Errorf("InspectService: client not initialized")
	}

	svc, _, err := c.cli.ServiceInspectWithRaw(ctx, serviceID, dockertypes.ServiceInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("InspectService: %w", err)
	}

	replicas := 0
	if svc.Spec.Mode.Replicated != nil && svc.Spec.Mode.Replicated.Replicas != nil {
		replicas = int(*svc.Spec.Mode.Replicated.Replicas)
	}

	status := "unknown"
	if svc.ServiceStatus != nil {
		if svc.ServiceStatus.RunningTasks > 0 {
			status = "running"
		} else if svc.ServiceStatus.DesiredTasks == 0 {
			status = "stopped"
		} else {
			status = "pending"
		}
	}

	return &ServiceInfo{
		ID:       svc.ID,
		Name:     svc.Spec.Name,
		Image:    svc.Spec.TaskTemplate.ContainerSpec.Image,
		Replicas: replicas,
		Status:   status,
	}, nil
}

// HostInfo summarizes the resources available on the Docker host.
type HostInfo struct {
	CPUCount   int    `json:"cpu_count"`    // logical CPUs
	MemoryMB   int    `json:"memory_mb"`    // total RAM in MB
	OSName     string `json:"os_name"`      // "linux" etc
	Kernel     string `json:"kernel"`       // kernel version
	ServerVer  string `json:"server_ver"`   // Docker daemon version
	DataRoot   string `json:"data_root"`    // host path of /var/lib/docker
}

// HostInfo queries the Docker daemon for host capacity.
func (c *Client) HostInfo(ctx context.Context) (*HostInfo, error) {
	if c.cli == nil {
		return nil, fmt.Errorf("HostInfo: client not initialized")
	}
	info, err := c.cli.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("HostInfo: %w", err)
	}
	return &HostInfo{
		CPUCount:  info.NCPU,
		MemoryMB:  int(info.MemTotal / (1024 * 1024)),
		OSName:    info.OperatingSystem,
		Kernel:    info.KernelVersion,
		ServerVer: info.ServerVersion,
		DataRoot:  info.DockerRootDir,
	}, nil
}

// GetContainerStats fetches a one-shot stats snapshot for a container.
// Returns cpu_percent (0-100 across all CPUs), memory_usage, memory_limit,
// network_rx, network_tx in bytes.
func (c *Client) GetContainerStats(ctx context.Context, containerID string) (map[string]interface{}, error) {
	if c.cli == nil {
		return nil, fmt.Errorf("GetContainerStats: client not initialized")
	}

	resp, err := c.cli.ContainerStatsOneShot(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("GetContainerStats: %w", err)
	}
	defer resp.Body.Close()

	var stats containertypes.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("GetContainerStats: decode: %w", err)
	}

	cpuPct := calcCPUPercent(&stats)
	memUsage := stats.MemoryStats.Usage
	memLimit := stats.MemoryStats.Limit

	var rx, tx uint64
	for _, n := range stats.Networks {
		rx += n.RxBytes
		tx += n.TxBytes
	}

	return map[string]interface{}{
		"cpu_percent":  cpuPct,
		"memory_usage": memUsage,
		"memory_limit": memLimit,
		"network_rx":   rx,
		"network_tx":   tx,
	}, nil
}

func calcCPUPercent(s *containertypes.StatsResponse) float64 {
	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(s.CPUStats.SystemUsage) - float64(s.PreCPUStats.SystemUsage)
	if sysDelta <= 0 || cpuDelta <= 0 {
		return 0
	}
	cpus := float64(s.CPUStats.OnlineCPUs)
	if cpus == 0 {
		cpus = float64(len(s.CPUStats.CPUUsage.PercpuUsage))
	}
	if cpus == 0 {
		cpus = 1
	}
	return (cpuDelta / sysDelta) * cpus * 100.0
}

// ---------- spec builder ----------

// buildSwarmServiceSpec translates our ServiceSpec into Docker Swarm's ServiceSpec.
// Resolves the network name into an ID. If the network doesn't exist it's treated
// as a soft warning (the service spec omits the network reference).
func (c *Client) buildSwarmServiceSpec(ctx context.Context, spec ServiceSpec) (swarm.ServiceSpec, error) {
	replicas := uint64(spec.Replicas)
	if spec.Replicas < 1 {
		replicas = 1
	}

	env := make([]string, 0, len(spec.EnvVars))
	for k, v := range spec.EnvVars {
		env = append(env, k+"="+v)
	}

	labels := spec.Labels
	if labels == nil {
		labels = map[string]string{}
	}

	container := swarm.ContainerSpec{
		Image:  spec.Image,
		Env:    env,
		Labels: labels,
	}

	// Resources
	resources := &swarm.ResourceRequirements{
		Limits: &swarm.Limit{},
	}
	if spec.CPULimit > 0 {
		resources.Limits.NanoCPUs = spec.CPULimit
	}
	if spec.MemoryLimit > 0 {
		resources.Limits.MemoryBytes = spec.MemoryLimit
	}

	// Restart policy
	restartCondition := swarm.RestartPolicyConditionAny
	switch spec.RestartPolicy {
	case "on-failure":
		restartCondition = swarm.RestartPolicyConditionOnFailure
	case "none":
		restartCondition = swarm.RestartPolicyConditionNone
	}

	taskSpec := swarm.TaskSpec{
		ContainerSpec: &container,
		Resources:     resources,
		RestartPolicy: &swarm.RestartPolicy{Condition: restartCondition},
	}

	// Network attachment (resolve name → ID)
	if spec.NetworkName != "" {
		netID, err := c.resolveNetworkID(ctx, spec.NetworkName)
		if err != nil {
			log.Warn().Err(err).Str("network", spec.NetworkName).Msg("Network not found; service will be attached on ingress only")
		} else {
			taskSpec.Networks = []swarm.NetworkAttachmentConfig{
				{Target: netID},
			}
		}
	}

	// Endpoint / port
	var endpoint *swarm.EndpointSpec
	if spec.Port > 0 {
		endpoint = &swarm.EndpointSpec{
			Ports: []swarm.PortConfig{
				{
					Protocol:    swarm.PortConfigProtocolTCP,
					TargetPort:  uint32(spec.Port),
					PublishMode: swarm.PortConfigPublishModeIngress,
				},
			},
		}
	}

	return swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name:   spec.Name,
			Labels: labels,
		},
		TaskTemplate: taskSpec,
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{Replicas: &replicas},
		},
		EndpointSpec: endpoint,
	}, nil
}

// resolveNetworkID finds a Docker network by name and returns its ID.
func (c *Client) resolveNetworkID(ctx context.Context, name string) (string, error) {
	nets, err := c.cli.NetworkList(ctx, networktypes.ListOptions{
		Filters: filtertypes.NewArgs(filtertypes.Arg("name", name)),
	})
	if err != nil {
		return "", err
	}
	for _, n := range nets {
		if n.Name == name {
			return n.ID, nil
		}
	}
	return "", fmt.Errorf("network %q not found", name)
}

// ---------- network operations (used by network.go helpers) ----------

// ensureNetwork creates a network if it doesn't exist. Idempotent.
func (c *Client) ensureNetwork(ctx context.Context, name, driver string, labels map[string]string) (string, error) {
	if id, err := c.resolveNetworkID(ctx, name); err == nil {
		return id, nil
	}
	resp, err := c.cli.NetworkCreate(ctx, name, networktypes.CreateOptions{
		Driver:     driver,
		Attachable: true,
		Labels:     labels,
	})
	if err != nil {
		return "", err
	}
	log.Info().Str("network", name).Str("driver", driver).Msg("Created docker network")
	return resp.ID, nil
}

// removeNetwork deletes a network by name. No-op if not found.
func (c *Client) removeNetwork(ctx context.Context, name string) error {
	id, err := c.resolveNetworkID(ctx, name)
	if err != nil {
		return nil
	}
	return c.cli.NetworkRemove(ctx, id)
}
