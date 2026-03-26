package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type NodeRepository struct {
	db *gorm.DB
}

func NewNodeRepository(db *gorm.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

func (r *NodeRepository) Create(ctx context.Context, node *models.Node) error {
	if err := r.db.WithContext(ctx).Create(node).Error; err != nil {
		return fmt.Errorf("NodeRepo.Create: %w", err)
	}
	return nil
}

func (r *NodeRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Node, error) {
	var node models.Node
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&node).Error; err != nil {
		return nil, fmt.Errorf("NodeRepo.FindByID: %w", err)
	}
	return &node, nil
}

func (r *NodeRepository) List(ctx context.Context) ([]models.Node, error) {
	var nodes []models.Node
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&nodes).Error; err != nil {
		return nil, fmt.Errorf("NodeRepo.List: %w", err)
	}
	return nodes, nil
}

func (r *NodeRepository) Update(ctx context.Context, node *models.Node) error {
	if err := r.db.WithContext(ctx).Save(node).Error; err != nil {
		return fmt.Errorf("NodeRepo.Update: %w", err)
	}
	return nil
}

func (r *NodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Node{}).Error; err != nil {
		return fmt.Errorf("NodeRepo.Delete: %w", err)
	}
	return nil
}
