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
	ErrInvalidBilling   = errors.New("paid billing requires price_monthly_cents > 0")
)

// CreateOrgInput carries optional per-org resource overrides and billing config
// supplied at creation. Any nil resource field falls back to the Free plan default.
type CreateOrgInput struct {
	OwnerID     uuid.UUID
	Name        string
	Slug        string
	Description *string

	// Resource overrides (nil = inherit from plan)
	CustomCPUCores     *int
	CustomRAMMB        *int
	CustomDiskGB       *int
	CustomMaxApps      *int
	CustomMaxDatabases *int

	// Billing (empty = free/USD/monthly defaults)
	BillingType       string // "free" | "paid"
	PriceMonthlyCents *int
	Currency          string // ISO 4217
	BillingCycle      string // "monthly" | "yearly" | "one_time"
}

// UpdateOrgResourcesInput carries fields a super admin can change after creation.
type UpdateOrgResourcesInput struct {
	CustomCPUCores     *int
	CustomRAMMB        *int
	CustomDiskGB       *int
	CustomMaxApps      *int
	CustomMaxDatabases *int
	BillingType        *string
	PriceMonthlyCents  *int
	Currency           *string
	BillingCycle       *string
}

var slugRegex = regexp.MustCompile(`[^a-z0-9-]`)

// CgroupEnforcer is the subset of orchestrator.CgroupManager that OrgService
// needs. Injected as an interface to avoid a service→orchestrator dependency.
type CgroupEnforcer interface {
	EnsureOrgSlice(orgSlug string, cpuCores, ramMB int) error
	UpdateOrgSlice(orgSlug string, cpuCores, ramMB int) error
	RemoveOrgSlice(orgSlug string) error
}

type OrgService struct {
	orgRepo  *repository.OrgRepository
	userRepo *repository.UserRepository
	mailer   *mailer.Mailer
	cfg      *config.Config
	cgroup   CgroupEnforcer
}

func NewOrgService(orgRepo *repository.OrgRepository, userRepo *repository.UserRepository, mailer *mailer.Mailer, cfg *config.Config) *OrgService {
	return &OrgService{
		orgRepo:  orgRepo,
		userRepo: userRepo,
		mailer:   mailer,
		cfg:      cfg,
	}
}

// SetCgroupEnforcer wires a cgroup manager for per-org resource enforcement.
// Optional — if never called, cgroup operations are skipped silently.
func (s *OrgService) SetCgroupEnforcer(c CgroupEnforcer) {
	s.cgroup = c
}

func (s *OrgService) CreateOrganization(ctx context.Context, in CreateOrgInput) (*models.Organization, error) {
	slug := sanitizeSlug(in.Slug)

	// Validate billing config
	billingType := in.BillingType
	if billingType == "" {
		billingType = "free"
	}
	if billingType != "free" && billingType != "paid" {
		return nil, fmt.Errorf("invalid billing_type %q", billingType)
	}
	if billingType == "paid" && (in.PriceMonthlyCents == nil || *in.PriceMonthlyCents <= 0) {
		return nil, ErrInvalidBilling
	}
	currency := in.Currency
	if currency == "" {
		currency = "USD"
	}
	cycle := in.BillingCycle
	if cycle == "" {
		cycle = "monthly"
	}

	// Slug uniqueness
	existing, err := s.orgRepo.FindOrgBySlug(ctx, slug)
	if err == nil && existing != nil {
		return nil, ErrOrgSlugTaken
	}

	// Assign default Free plan as a baseline template (custom_* fields override per-field)
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
		ID:                 uuid.New(),
		Name:               in.Name,
		Slug:               slug,
		Description:        in.Description,
		PlanID:             planID,
		CreatedBy:          in.OwnerID,
		CustomCPUCores:     in.CustomCPUCores,
		CustomRAMMB:        in.CustomRAMMB,
		CustomDiskGB:       in.CustomDiskGB,
		CustomMaxApps:      in.CustomMaxApps,
		CustomMaxDatabases: in.CustomMaxDatabases,
		BillingType:        billingType,
		PriceMonthlyCents:  in.PriceMonthlyCents,
		Currency:           currency,
		BillingCycle:       cycle,
	}

	if err := s.orgRepo.CreateOrg(ctx, org); err != nil {
		return nil, fmt.Errorf("CreateOrganization: %w", err)
	}

	// Owner membership
	member := &models.OrgMember{
		OrgID:    org.ID,
		UserID:   in.OwnerID,
		Role:     models.RoleOwner,
		JoinedAt: time.Now(),
	}
	if err := s.orgRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("CreateOrganization: add owner: %w", err)
	}

	// Docker network — non-fatal
	if err := docker.CreateOrgNetwork(slug); err != nil {
		fmt.Printf("Warning: failed to create Docker network for org %s: %v\n", slug, err)
	}

	// cgroup slice — non-fatal (skipped when enforcement disabled)
	if s.cgroup != nil {
		if err := s.cgroup.EnsureOrgSlice(slug, org.EffectiveCPUCores(), org.EffectiveRAMMB()); err != nil {
			fmt.Printf("Warning: cgroup slice setup failed for org %s: %v\n", slug, err)
		}
	}

	// Reload with plan
	org, _ = s.orgRepo.FindOrgBySlug(ctx, slug)

	return org, nil
}

// UpdateOrgResources lets a super admin edit resource overrides + billing config
// for an existing organization.
func (s *OrgService) UpdateOrgResources(ctx context.Context, slug string, in UpdateOrgResourcesInput) (*models.Organization, error) {
	org, err := s.orgRepo.FindOrgBySlug(ctx, slug)
	if err != nil {
		return nil, ErrOrgNotFound
	}

	if in.CustomCPUCores != nil {
		org.CustomCPUCores = in.CustomCPUCores
	}
	if in.CustomRAMMB != nil {
		org.CustomRAMMB = in.CustomRAMMB
	}
	if in.CustomDiskGB != nil {
		org.CustomDiskGB = in.CustomDiskGB
	}
	if in.CustomMaxApps != nil {
		org.CustomMaxApps = in.CustomMaxApps
	}
	if in.CustomMaxDatabases != nil {
		org.CustomMaxDatabases = in.CustomMaxDatabases
	}
	if in.BillingType != nil {
		if *in.BillingType != "free" && *in.BillingType != "paid" {
			return nil, fmt.Errorf("invalid billing_type %q", *in.BillingType)
		}
		org.BillingType = *in.BillingType
	}
	if in.PriceMonthlyCents != nil {
		org.PriceMonthlyCents = in.PriceMonthlyCents
	}
	if in.Currency != nil {
		org.Currency = *in.Currency
	}
	if in.BillingCycle != nil {
		org.BillingCycle = *in.BillingCycle
	}

	// Final cross-check: paid must have price
	if org.BillingType == "paid" && (org.PriceMonthlyCents == nil || *org.PriceMonthlyCents <= 0) {
		return nil, ErrInvalidBilling
	}

	if err := s.orgRepo.UpdateOrg(ctx, org); err != nil {
		return nil, fmt.Errorf("UpdateOrgResources: %w", err)
	}

	// Push new limits to the kernel — non-fatal
	if s.cgroup != nil {
		if err := s.cgroup.UpdateOrgSlice(slug, org.EffectiveCPUCores(), org.EffectiveRAMMB()); err != nil {
			fmt.Printf("Warning: cgroup slice update failed for org %s: %v\n", slug, err)
		}
	}

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

	// Remove cgroup slice
	if s.cgroup != nil {
		_ = s.cgroup.RemoveOrgSlice(slug)
	}

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
