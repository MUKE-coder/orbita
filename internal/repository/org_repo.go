package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type OrgRepository struct {
	db *gorm.DB
}

func NewOrgRepository(db *gorm.DB) *OrgRepository {
	return &OrgRepository{db: db}
}

// Organizations

func (r *OrgRepository) CreateOrg(ctx context.Context, org *models.Organization) error {
	if err := r.db.WithContext(ctx).Create(org).Error; err != nil {
		return fmt.Errorf("CreateOrg: %w", err)
	}
	return nil
}

func (r *OrgRepository) FindOrgBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	var org models.Organization
	if err := r.db.WithContext(ctx).Preload("Plan").Where("slug = ?", slug).First(&org).Error; err != nil {
		return nil, fmt.Errorf("FindOrgBySlug: %w", err)
	}
	return &org, nil
}

func (r *OrgRepository) FindOrgByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	if err := r.db.WithContext(ctx).Preload("Plan").Where("id = ?", id).First(&org).Error; err != nil {
		return nil, fmt.Errorf("FindOrgByID: %w", err)
	}
	return &org, nil
}

func (r *OrgRepository) UpdateOrg(ctx context.Context, org *models.Organization) error {
	if err := r.db.WithContext(ctx).Save(org).Error; err != nil {
		return fmt.Errorf("UpdateOrg: %w", err)
	}
	return nil
}

func (r *OrgRepository) DeleteOrg(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Organization{}).Error; err != nil {
		return fmt.Errorf("DeleteOrg: %w", err)
	}
	return nil
}

func (r *OrgRepository) ListOrgsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Organization, error) {
	var orgs []models.Organization
	if err := r.db.WithContext(ctx).
		Preload("Plan").
		Joins("JOIN org_members ON org_members.org_id = organizations.id").
		Where("org_members.user_id = ?", userID).
		Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("ListOrgsByUserID: %w", err)
	}
	return orgs, nil
}

func (r *OrgRepository) ListAllOrgs(ctx context.Context) ([]models.Organization, error) {
	var orgs []models.Organization
	if err := r.db.WithContext(ctx).Preload("Plan").Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("ListAllOrgs: %w", err)
	}
	return orgs, nil
}

// Members

func (r *OrgRepository) AddMember(ctx context.Context, member *models.OrgMember) error {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("AddMember: %w", err)
	}
	return nil
}

func (r *OrgRepository) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*models.OrgMember, error) {
	var member models.OrgMember
	if err := r.db.WithContext(ctx).Where("org_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("GetMember: %w", err)
	}
	return &member, nil
}

func (r *OrgRepository) ListMembers(ctx context.Context, orgID uuid.UUID) ([]models.OrgMember, error) {
	var members []models.OrgMember
	if err := r.db.WithContext(ctx).Preload("User").Where("org_id = ?", orgID).
		Order("joined_at ASC").Find(&members).Error; err != nil {
		return nil, fmt.Errorf("ListMembers: %w", err)
	}
	return members, nil
}

func (r *OrgRepository) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	if err := r.db.WithContext(ctx).Model(&models.OrgMember{}).
		Where("org_id = ? AND user_id = ?", orgID, userID).
		Update("role", role).Error; err != nil {
		return fmt.Errorf("UpdateMemberRole: %w", err)
	}
	return nil
}

func (r *OrgRepository) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("org_id = ? AND user_id = ?", orgID, userID).
		Delete(&models.OrgMember{}).Error; err != nil {
		return fmt.Errorf("RemoveMember: %w", err)
	}
	return nil
}

func (r *OrgRepository) CountMembers(ctx context.Context, orgID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.OrgMember{}).Where("org_id = ?", orgID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("CountMembers: %w", err)
	}
	return count, nil
}

// Invites

func (r *OrgRepository) CreateInvite(ctx context.Context, invite *models.OrgInvite) error {
	if err := r.db.WithContext(ctx).Create(invite).Error; err != nil {
		return fmt.Errorf("CreateInvite: %w", err)
	}
	return nil
}

func (r *OrgRepository) FindInviteByTokenHash(ctx context.Context, hash string) (*models.OrgInvite, error) {
	var invite models.OrgInvite
	if err := r.db.WithContext(ctx).
		Preload("Inviter").
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hash, time.Now()).
		First(&invite).Error; err != nil {
		return nil, fmt.Errorf("FindInviteByTokenHash: %w", err)
	}
	return &invite, nil
}

func (r *OrgRepository) FindInviteByID(ctx context.Context, id uuid.UUID) (*models.OrgInvite, error) {
	var invite models.OrgInvite
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&invite).Error; err != nil {
		return nil, fmt.Errorf("FindInviteByID: %w", err)
	}
	return &invite, nil
}

func (r *OrgRepository) ListPendingInvites(ctx context.Context, orgID uuid.UUID) ([]models.OrgInvite, error) {
	var invites []models.OrgInvite
	if err := r.db.WithContext(ctx).
		Preload("Inviter").
		Where("org_id = ? AND used_at IS NULL AND expires_at > ?", orgID, time.Now()).
		Order("created_at DESC").Find(&invites).Error; err != nil {
		return nil, fmt.Errorf("ListPendingInvites: %w", err)
	}
	return invites, nil
}

func (r *OrgRepository) MarkInviteUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&models.OrgInvite{}).Where("id = ?", id).
		Update("used_at", &now).Error; err != nil {
		return fmt.Errorf("MarkInviteUsed: %w", err)
	}
	return nil
}

func (r *OrgRepository) DeleteInvite(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.OrgInvite{}).Error; err != nil {
		return fmt.Errorf("DeleteInvite: %w", err)
	}
	return nil
}

// Resource Plans

func (r *OrgRepository) ListPlans(ctx context.Context) ([]models.ResourcePlan, error) {
	var plans []models.ResourcePlan
	if err := r.db.WithContext(ctx).Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("ListPlans: %w", err)
	}
	return plans, nil
}

func (r *OrgRepository) FindPlanByID(ctx context.Context, id uuid.UUID) (*models.ResourcePlan, error) {
	var plan models.ResourcePlan
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&plan).Error; err != nil {
		return nil, fmt.Errorf("FindPlanByID: %w", err)
	}
	return &plan, nil
}

func (r *OrgRepository) CreatePlan(ctx context.Context, plan *models.ResourcePlan) error {
	if err := r.db.WithContext(ctx).Create(plan).Error; err != nil {
		return fmt.Errorf("CreatePlan: %w", err)
	}
	return nil
}

func (r *OrgRepository) UpdatePlan(ctx context.Context, plan *models.ResourcePlan) error {
	if err := r.db.WithContext(ctx).Save(plan).Error; err != nil {
		return fmt.Errorf("UpdatePlan: %w", err)
	}
	return nil
}

func (r *OrgRepository) DeletePlan(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.ResourcePlan{}).Error; err != nil {
		return fmt.Errorf("DeletePlan: %w", err)
	}
	return nil
}

func (r *OrgRepository) AssignPlanToOrg(ctx context.Context, orgID, planID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Organization{}).Where("id = ?", orgID).
		Update("plan_id", planID).Error; err != nil {
		return fmt.Errorf("AssignPlanToOrg: %w", err)
	}
	return nil
}
