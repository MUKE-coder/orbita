package orchestrator

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

type NodeManager struct {
	nodeRepo *repository.NodeRepository
}

func NewNodeManager(nodeRepo *repository.NodeRepository) *NodeManager {
	return &NodeManager{nodeRepo: nodeRepo}
}

func (m *NodeManager) AddNode(ctx context.Context, name, ip string, sshPort int, sshPrivateKey string) (*models.Node, error) {
	// TODO: real impl
	// 1. SSH into node, verify Docker is installed
	// 2. Run `docker swarm join --token {workerToken} {primaryIP}:2377`
	// 3. Label node in Swarm: `orbita.node.id={nodeID}`

	node := &models.Node{
		ID:      uuid.New(),
		Name:    name,
		IP:      ip,
		SSHPort: sshPort,
		Role:    models.NodeRoleWorker,
		Status:  models.NodeStatusOnline,
	}

	if err := m.nodeRepo.Create(ctx, node); err != nil {
		return nil, fmt.Errorf("AddNode: %w", err)
	}

	log.Info().Str("name", name).Str("ip", ip).Msg("Node added (stub)")
	return node, nil
}

func (m *NodeManager) DrainNode(ctx context.Context, nodeID uuid.UUID) error {
	node, err := m.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("DrainNode: %w", err)
	}

	// TODO: real impl — docker node update --availability drain {swarmNodeID}
	node.Status = models.NodeStatusDraining
	if err := m.nodeRepo.Update(ctx, node); err != nil {
		return fmt.Errorf("DrainNode: update: %w", err)
	}

	log.Info().Str("node", node.Name).Msg("Node draining (stub)")
	return nil
}

func (m *NodeManager) RemoveNode(ctx context.Context, nodeID uuid.UUID) error {
	node, err := m.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("RemoveNode: %w", err)
	}

	// TODO: real impl — docker node rm {swarmNodeID}
	_ = node

	if err := m.nodeRepo.Delete(ctx, nodeID); err != nil {
		return fmt.Errorf("RemoveNode: delete: %w", err)
	}

	log.Info().Str("node_id", nodeID.String()).Msg("Node removed (stub)")
	return nil
}

func (m *NodeManager) GetNodeMetrics(ctx context.Context, nodeID uuid.UUID) (*models.NodeMetrics, error) {
	// TODO: real impl — SSH exec for system metrics OR docker stats API
	return &models.NodeMetrics{
		CPUPercent:     15.5,
		MemoryUsed:     4294967296,
		MemoryTotal:    8589934592,
		DiskUsed:       21474836480,
		DiskTotal:      107374182400,
		ContainerCount: 8,
		Uptime:         864000,
	}, nil
}

func (m *NodeManager) ListNodes(ctx context.Context) ([]models.Node, error) {
	return m.nodeRepo.List(ctx)
}

func (m *NodeManager) GetNode(ctx context.Context, id uuid.UUID) (*models.Node, error) {
	return m.nodeRepo.FindByID(ctx, id)
}

func (m *NodeManager) UpdateNode(ctx context.Context, node *models.Node) error {
	return m.nodeRepo.Update(ctx, node)
}
