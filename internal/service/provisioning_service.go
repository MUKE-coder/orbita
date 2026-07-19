package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

var (
	ErrUserEmailTaken   = errors.New("a user with that email already exists")
	ErrInvalidOrgRole   = errors.New("role must be one of: viewer, developer, admin, owner")
	ErrWeakPassword     = errors.New("password must be at least 8 characters")
	ErrCannotSelfManage = errors.New("you cannot change your own super-admin status")
)

// ProvisioningService lets a super admin onboard a tenant directly: create the
// organisation, size it, and create the account its admin signs in with.
//
// This is the path for instances with no email provider configured, where an
// invite link would never arrive. The generated password is returned exactly
// once, for the operator to hand over, and the account is flagged so the user
// must replace it at first sign-in.
type ProvisioningService struct {
	userRepo *repository.UserRepository
	orgRepo  *repository.OrgRepository
	orgSvc   *OrgService
}

func NewProvisioningService(userRepo *repository.UserRepository, orgRepo *repository.OrgRepository, orgSvc *OrgService) *ProvisioningService {
	return &ProvisioningService{userRepo: userRepo, orgRepo: orgRepo, orgSvc: orgSvc}
}

// ProvisionOrgInput describes a new tenant and the account that will run it.
type ProvisionOrgInput struct {
	// Organisation
	Name        string
	Slug        string
	Description *string
	PlanID      *uuid.UUID

	// Resource overrides (nil = inherit from the plan)
	CustomCPUCores     *int
	CustomRAMMB        *int
	CustomDiskGB       *int
	CustomMaxApps      *int
	CustomMaxDatabases *int

	// Billing
	BillingType       string
	PriceMonthlyCents *int
	Currency          string
	BillingCycle      string

	// The org's admin account
	AdminEmail string
	AdminName  string
	// Optional: an operator-chosen password. Empty means generate one.
	AdminPassword string
	// Role within the org. Defaults to owner — the tenant runs their own org.
	AdminRole string
}

// ProvisionResult carries the credentials to hand over. The password is
// plaintext and is never persisted or logged; it exists only in this response.
type ProvisionResult struct {
	Organization *models.Organization `json:"organization"`
	User         *models.User         `json:"user"`
	Password     string               `json:"password"`
	Generated    bool                 `json:"generated"`
}

// ProvisionOrgWithAdmin creates the organisation and its admin account together.
func (s *ProvisioningService) ProvisionOrgWithAdmin(ctx context.Context, in ProvisionOrgInput, actor uuid.UUID) (*ProvisionResult, error) {
	email := strings.ToLower(strings.TrimSpace(in.AdminEmail))
	if email == "" {
		return nil, fmt.Errorf("ProvisionOrgWithAdmin: admin email is required")
	}

	role := in.AdminRole
	if role == "" {
		role = models.RoleOwner
	}
	if !isValidOrgRole(role) {
		return nil, ErrInvalidOrgRole
	}

	// Fail before creating the org, so a taken email doesn't leave an orphan.
	if existing, err := s.userRepo.FindUserByEmail(ctx, email); err == nil && existing != nil {
		return nil, ErrUserEmailTaken
	}

	password, generated, err := resolvePassword(in.AdminPassword)
	if err != nil {
		return nil, err
	}

	user, err := s.createManagedUser(ctx, email, in.AdminName, password, actor)
	if err != nil {
		return nil, err
	}

	// The new account owns the org, so the tenant administers their own space.
	org, err := s.orgSvc.CreateOrganization(ctx, CreateOrgInput{
		OwnerID:            user.ID,
		Name:               in.Name,
		Slug:               in.Slug,
		Description:        in.Description,
		CustomCPUCores:     in.CustomCPUCores,
		CustomRAMMB:        in.CustomRAMMB,
		CustomDiskGB:       in.CustomDiskGB,
		CustomMaxApps:      in.CustomMaxApps,
		CustomMaxDatabases: in.CustomMaxDatabases,
		BillingType:        in.BillingType,
		PriceMonthlyCents:  in.PriceMonthlyCents,
		Currency:           in.Currency,
		BillingCycle:       in.BillingCycle,
	})
	if err != nil {
		// Roll back the account: leaving a user with no org would block the
		// operator from retrying with the same email.
		if delErr := s.userRepo.HardDeleteUser(ctx, user.ID); delErr != nil {
			log.Error().Err(delErr).Str("email", email).
				Msg("provisioning: org creation failed and the orphaned user could not be removed")
		}
		return nil, err
	}

	// CreateOrganization always seats the owner as RoleOwner; downgrade if the
	// operator asked for something narrower.
	if role != models.RoleOwner {
		if err := s.orgRepo.UpdateMemberRole(ctx, org.ID, user.ID, role); err != nil {
			log.Warn().Err(err).Msg("provisioning: could not apply the requested role")
		}
	}

	// Apply an explicit plan if one was chosen (CreateOrganization defaults to Free).
	if in.PlanID != nil {
		if err := s.orgRepo.AssignPlanToOrg(ctx, org.ID, *in.PlanID); err != nil {
			log.Warn().Err(err).Msg("provisioning: could not assign the requested plan")
		} else if reloaded, err := s.orgRepo.FindOrgBySlug(ctx, org.Slug); err == nil {
			org = reloaded
		}
	}

	log.Info().
		Str("org", org.Slug).
		Str("admin", email).
		Str("role", role).
		Msg("Provisioned organisation with an admin account")

	return &ProvisionResult{Organization: org, User: user, Password: password, Generated: generated}, nil
}

// CreateUserInOrgInput adds a person to an existing org without email.
type CreateUserInOrgInput struct {
	Email    string
	Name     string
	Password string // empty = generate
	Role     string
}

// CreateUserInOrg provisions an account and seats it in an existing org. This is
// the no-email equivalent of sending an invite.
func (s *ProvisioningService) CreateUserInOrg(ctx context.Context, orgID uuid.UUID, in CreateUserInOrgInput, actor uuid.UUID) (*ProvisionResult, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" {
		return nil, fmt.Errorf("CreateUserInOrg: email is required")
	}
	role := in.Role
	if role == "" {
		role = models.RoleDeveloper
	}
	if !isValidOrgRole(role) {
		return nil, ErrInvalidOrgRole
	}

	if existing, err := s.userRepo.FindUserByEmail(ctx, email); err == nil && existing != nil {
		return nil, ErrUserEmailTaken
	}

	password, generated, err := resolvePassword(in.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.createManagedUser(ctx, email, in.Name, password, actor)
	if err != nil {
		return nil, err
	}

	member := &models.OrgMember{
		OrgID:    orgID,
		UserID:   user.ID,
		Role:     role,
		JoinedAt: time.Now(),
	}
	if err := s.orgRepo.AddMember(ctx, member); err != nil {
		if delErr := s.userRepo.HardDeleteUser(ctx, user.ID); delErr != nil {
			log.Error().Err(delErr).Msg("provisioning: could not remove the orphaned user")
		}
		return nil, fmt.Errorf("CreateUserInOrg: add member: %w", err)
	}

	org, _ := s.orgRepo.FindOrgByID(ctx, orgID)
	return &ProvisionResult{Organization: org, User: user, Password: password, Generated: generated}, nil
}

// createManagedUser writes an admin-provisioned account. Such accounts are
// pre-verified (an operator vouched for the address) but must rotate their
// password at first sign-in.
func (s *ProvisioningService) createManagedUser(ctx context.Context, email, name, password string, actor uuid.UUID) (*models.User, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("createManagedUser: hash: %w", err)
	}
	if name == "" {
		name = email
	}

	user := &models.User{
		ID:                 uuid.New(),
		Email:              email,
		PasswordHash:       hash,
		Name:               name,
		IsEmailVerified:    true,
		MustChangePassword: true,
		CreatedBy:          &actor,
	}
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("createManagedUser: %w", err)
	}
	return user, nil
}

// resolvePassword returns the password to use and whether it was generated.
func resolvePassword(chosen string) (string, bool, error) {
	chosen = strings.TrimSpace(chosen)
	if chosen == "" {
		generated, err := auth.GeneratePassword(16)
		if err != nil {
			return "", false, fmt.Errorf("resolvePassword: %w", err)
		}
		return generated, true, nil
	}
	if len(chosen) < 8 {
		return "", false, ErrWeakPassword
	}
	return chosen, false, nil
}

func isValidOrgRole(role string) bool {
	switch role {
	case models.RoleViewer, models.RoleDeveloper, models.RoleAdmin, models.RoleOwner:
		return true
	}
	return false
}
