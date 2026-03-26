package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/config"
	"github.com/orbita-sh/orbita/internal/docker"
	"github.com/orbita-sh/orbita/internal/mailer"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

var (
	ErrOrgNotFound      = errors.New("organization not found")
	ErrOrgSlugTaken     = errors.New("organization slug already taken")
	ErrMemberNotFound   = errors.New("member not found")
	ErrAlreadyMember    = errors.New("user is already a member")
	ErrCannotLeave      = errors.New("owner cannot leave the organization")
	ErrInsufficientRole = errors.New("insufficient role")
	ErrInviteNotFound   = errors.New("invite not found")
	ErrInviteExpired    = errors.New("invite expired or already used")
	ErrSelfRoleChange   = errors.New("cannot change your own role")
	ErrPlanNotFound     = errors.New("plan not found")
)

var slugRegex = regexp.MustCompile(`[^a-z0-9-]`)

type OrgService struct {
	orgRepo  *repository.OrgRepository
	userRepo *repository.UserRepository
	mailer   *mailer.Mailer
	cfg      *config.Config
}

func NewOrgService(orgRepo *repository.OrgRepository, userRepo *repository.UserRepository, mailer *mailer.Mailer, cfg *config.Config) *OrgService {
	return &OrgService{
		orgRepo:  orgRepo,
		userRepo: userRepo,
		mailer:   mailer,
		cfg:      cfg,
	}
}

func (s *OrgService) CreateOrganization(ctx context.Context, ownerID uuid.UUID, name, slug string) (*models.Organization, error) {
	slug = sanitizeSlug(slug)

	// Check slug uniqueness
	existing, err := s.orgRepo.FindOrgBySlug(ctx, slug)
	if err == nil && existing != nil {
		return nil, ErrOrgSlugTaken
	}

	// Assign default plan (Free)
	plans, _ := s.orgRepo.ListPlans(ctx)
	var planID *uuid.UUID
	for _, p := range plans {
		if p.Name == "Free" {
			id := p.ID
			planID = &id
			break
		}
	}

	org := &models.Organization{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		PlanID:    planID,
		CreatedBy: ownerID,
	}

	if err := s.orgRepo.CreateOrg(ctx, org); err != nil {
		return nil, fmt.Errorf("CreateOrganization: %w", err)
	}

	// Add owner as member
	member := &models.OrgMember{
		OrgID:    org.ID,
		UserID:   ownerID,
		Role:     models.RoleOwner,
		JoinedAt: time.Now(),
	}
	if err := s.orgRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("CreateOrganization: add owner: %w", err)
	}

	// Create Docker network
	if err := docker.CreateOrgNetwork(slug); err != nil {
		// Non-fatal, log and continue
		fmt.Printf("Warning: failed to create Docker network for org %s: %v\n", slug, err)
	}

	// Reload with plan
	org, _ = s.orgRepo.FindOrgBySlug(ctx, slug)

	return org, nil
}

func (s *OrgService) GetOrganization(ctx context.Context, slug string) (*models.Organization, error) {
	org, err := s.orgRepo.FindOrgBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, fmt.Errorf("GetOrganization: %w", err)
	}
	return org, nil
}

func (s *OrgService) UpdateOrganization(ctx context.Context, slug string, name, description *string) (*models.Organization, error) {
	org, err := s.orgRepo.FindOrgBySlug(ctx, slug)
	if err != nil {
		return nil, ErrOrgNotFound
	}

	if name != nil {
		org.Name = *name
	}
	if description != nil {
		org.Description = description
	}

	if err := s.orgRepo.UpdateOrg(ctx, org); err != nil {
		return nil, fmt.Errorf("UpdateOrganization: %w", err)
	}
	return org, nil
}

func (s *OrgService) DeleteOrganization(ctx context.Context, slug string) error {
	org, err := s.orgRepo.FindOrgBySlug(ctx, slug)
	if err != nil {
		return ErrOrgNotFound
	}

	if err := s.orgRepo.DeleteOrg(ctx, org.ID); err != nil {
		return fmt.Errorf("DeleteOrganization: %w", err)
	}

	// Clean up Docker network
	_ = docker.DeleteOrgNetwork(slug)

	return nil
}

func (s *OrgService) ListUserOrgs(ctx context.Context, userID uuid.UUID) ([]models.Organization, error) {
	return s.orgRepo.ListOrgsByUserID(ctx, userID)
}

func (s *OrgService) ListAllOrgs(ctx context.Context) ([]models.Organization, error) {
	return s.orgRepo.ListAllOrgs(ctx)
}

// Members

func (s *OrgService) ListMembers(ctx context.Context, orgID uuid.UUID) ([]models.OrgMember, error) {
	return s.orgRepo.ListMembers(ctx, orgID)
}

func (s *OrgService) GetMemberRole(ctx context.Context, orgID, userID uuid.UUID) (string, error) {
	member, err := s.orgRepo.GetMember(ctx, orgID, userID)
	if err != nil {
		return "", ErrMemberNotFound
	}
	return member.Role, nil
}

func (s *OrgService) UpdateMemberRole(ctx context.Context, orgID, targetUserID, requesterID uuid.UUID, newRole string) error {
	if targetUserID == requesterID {
		return ErrSelfRoleChange
	}

	// Verify target is a member
	target, err := s.orgRepo.GetMember(ctx, orgID, targetUserID)
	if err != nil {
		return ErrMemberNotFound
	}

	// Cannot change the owner's role
	if target.Role == models.RoleOwner {
		return ErrInsufficientRole
	}

	return s.orgRepo.UpdateMemberRole(ctx, orgID, targetUserID, newRole)
}

func (s *OrgService) RemoveMember(ctx context.Context, orgID, targetUserID, requesterID uuid.UUID) error {
	target, err := s.orgRepo.GetMember(ctx, orgID, targetUserID)
	if err != nil {
		return ErrMemberNotFound
	}

	if target.Role == models.RoleOwner {
		return ErrCannotLeave
	}

	return s.orgRepo.RemoveMember(ctx, orgID, targetUserID)
}

func (s *OrgService) LeaveOrg(ctx context.Context, orgID, userID uuid.UUID) error {
	member, err := s.orgRepo.GetMember(ctx, orgID, userID)
	if err != nil {
		return ErrMemberNotFound
	}

	if member.Role == models.RoleOwner {
		return ErrCannotLeave
	}

	return s.orgRepo.RemoveMember(ctx, orgID, userID)
}

// Invites

func (s *OrgService) InviteMember(ctx context.Context, orgID uuid.UUID, email, role string, inviterID uuid.UUID) error {
	// Check if already a member
	org, err := s.orgRepo.FindOrgByID(ctx, orgID)
	if err != nil {
		return ErrOrgNotFound
	}

	// Check if user exists and is already a member
	existingUser, err := s.userRepo.FindUserByEmail(ctx, email)
	if err == nil {
		_, memberErr := s.orgRepo.GetMember(ctx, orgID, existingUser.ID)
		if memberErr == nil {
			return ErrAlreadyMember
		}
	}

	token, err := auth.GenerateRandomToken(32)
	if err != nil {
		return fmt.Errorf("InviteMember: generate token: %w", err)
	}

	invite := &models.OrgInvite{
		ID:        uuid.New(),
		OrgID:     orgID,
		Email:     email,
		Role:      role,
		TokenHash: auth.HashToken(token),
		InvitedBy: inviterID,
		ExpiresAt: time.Now().Add(72 * time.Hour),
	}

	if err := s.orgRepo.CreateInvite(ctx, invite); err != nil {
		return fmt.Errorf("InviteMember: %w", err)
	}

	// Send invite email
	inviter, _ := s.userRepo.FindUserByID(ctx, inviterID)
	inviterName := "A team member"
	if inviter != nil {
		inviterName = inviter.Name
	}

	acceptURL := fmt.Sprintf("%s/join?token=%s", s.cfg.AppBaseURL, token)
	go func() {
		_ = s.mailer.SendInvite(context.Background(), email, org.Name, inviterName, acceptURL)
	}()

	return nil
}

func (s *OrgService) AcceptInvite(ctx context.Context, token string, userID uuid.UUID) error {
	tokenHash := auth.HashToken(token)
	invite, err := s.orgRepo.FindInviteByTokenHash(ctx, tokenHash)
	if err != nil {
		return ErrInviteNotFound
	}

	// Check if already a member
	_, memberErr := s.orgRepo.GetMember(ctx, invite.OrgID, userID)
	if memberErr == nil {
		// Already a member, just mark invite as used
		_ = s.orgRepo.MarkInviteUsed(ctx, invite.ID)
		return ErrAlreadyMember
	}

	member := &models.OrgMember{
		OrgID:    invite.OrgID,
		UserID:   userID,
		Role:     invite.Role,
		JoinedAt: time.Now(),
	}

	if err := s.orgRepo.AddMember(ctx, member); err != nil {
		return fmt.Errorf("AcceptInvite: %w", err)
	}

	_ = s.orgRepo.MarkInviteUsed(ctx, invite.ID)

	return nil
}

func (s *OrgService) GetInviteInfo(ctx context.Context, token string) (*models.OrgInvite, *models.Organization, error) {
	tokenHash := auth.HashToken(token)
	invite, err := s.orgRepo.FindInviteByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, ErrInviteNotFound
	}

	org, err := s.orgRepo.FindOrgByID(ctx, invite.OrgID)
	if err != nil {
		return nil, nil, ErrOrgNotFound
	}

	return invite, org, nil
}

func (s *OrgService) ListPendingInvites(ctx context.Context, orgID uuid.UUID) ([]models.OrgInvite, error) {
	return s.orgRepo.ListPendingInvites(ctx, orgID)
}

func (s *OrgService) RevokeInvite(ctx context.Context, inviteID uuid.UUID) error {
	return s.orgRepo.DeleteInvite(ctx, inviteID)
}

// Resource Plans

func (s *OrgService) ListPlans(ctx context.Context) ([]models.ResourcePlan, error) {
	return s.orgRepo.ListPlans(ctx)
}

func (s *OrgService) CreatePlan(ctx context.Context, plan *models.ResourcePlan) error {
	plan.ID = uuid.New()
	return s.orgRepo.CreatePlan(ctx, plan)
}

func (s *OrgService) UpdatePlan(ctx context.Context, plan *models.ResourcePlan) error {
	return s.orgRepo.UpdatePlan(ctx, plan)
}

func (s *OrgService) DeletePlan(ctx context.Context, id uuid.UUID) error {
	return s.orgRepo.DeletePlan(ctx, id)
}

func (s *OrgService) AssignPlanToOrg(ctx context.Context, orgSlug string, planID uuid.UUID) error {
	org, err := s.orgRepo.FindOrgBySlug(ctx, orgSlug)
	if err != nil {
		return ErrOrgNotFound
	}

	_, err = s.orgRepo.FindPlanByID(ctx, planID)
	if err != nil {
		return ErrPlanNotFound
	}

	return s.orgRepo.AssignPlanToOrg(ctx, org.ID, planID)
}

func sanitizeSlug(slug string) string {
	slug = strings.ToLower(strings.TrimSpace(slug))
	slug = slugRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
