package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/orchestrator"
	"github.com/orbita-sh/orbita/internal/response"
)

type NodeHandler struct {
	nodeManager *orchestrator.NodeManager
}

func NewNodeHandler(nodeManager *orchestrator.NodeManager) *NodeHandler {
	return &NodeHandler{nodeManager: nodeManager}
}

type AddNodeRequest struct {
	Name          string `json:"name" binding:"required"`
	IP            string `json:"ip" binding:"required"`
	SSHPort       int    `json:"ssh_port"`
	SSHPrivateKey string `json:"ssh_private_key" binding:"required"`
}

func (h *NodeHandler) ListNodes(c *gin.Context) {
	nodes, err := h.nodeManager.ListNodes(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to list nodes")
		return
	}
	response.Success(c, http.StatusOK, nodes)
}

func (h *NodeHandler) AddNode(c *gin.Context) {
	var req AddNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	sshPort := req.SSHPort
	if sshPort == 0 {
		sshPort = 22
	}

	node, err := h.nodeManager.AddNode(c.Request.Context(), req.Name, req.IP, sshPort, req.SSHPrivateKey)
	if err != nil {
		response.InternalError(c, "Failed to add node")
		return
	}
	response.Success(c, http.StatusCreated, node)
}

func (h *NodeHandler) GetNode(c *gin.Context) {
	nodeID, err := uuid.Parse(c.Param("nodeId"))
	if err != nil {
		response.BadRequest(c, "Invalid node ID")
		return
	}

	node, err := h.nodeManager.GetNode(c.Request.Context(), nodeID)
	if err != nil {
		response.NotFound(c, "Node not found")
		return
	}
	response.Success(c, http.StatusOK, node)
}

func (h *NodeHandler) GetNodeMetrics(c *gin.Context) {
	nodeID, err := uuid.Parse(c.Param("nodeId"))
	if err != nil {
		response.BadRequest(c, "Invalid node ID")
		return
	}

	metrics, err := h.nodeManager.GetNodeMetrics(c.Request.Context(), nodeID)
	if err != nil {
		response.InternalError(c, "Failed to get node metrics")
		return
	}
	response.Success(c, http.StatusOK, metrics)
}

func (h *NodeHandler) DrainNode(c *gin.Context) {
	nodeID, _ := uuid.Parse(c.Param("nodeId"))
	if err := h.nodeManager.DrainNode(c.Request.Context(), nodeID); err != nil {
		response.InternalError(c, "Failed to drain node")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Node draining"})
}

func (h *NodeHandler) RemoveNode(c *gin.Context) {
	nodeID, _ := uuid.Parse(c.Param("nodeId"))
	if err := h.nodeManager.RemoveNode(c.Request.Context(), nodeID); err != nil {
		response.InternalError(c, "Failed to remove node")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Node removed"})
}

func (h *NodeHandler) GetPlatformMetrics(c *gin.Context) {
	nodes, err := h.nodeManager.ListNodes(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to get platform metrics")
		return
	}

	totalNodes := len(nodes)
	onlineNodes := 0
	for _, n := range nodes {
		if n.Status == "online" {
			onlineNodes++
		}
	}

	response.Success(c, http.StatusOK, gin.H{
		"total_nodes":  totalNodes,
		"online_nodes": onlineNodes,
	})
}
