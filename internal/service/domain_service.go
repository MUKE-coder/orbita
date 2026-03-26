package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
	"github.com/orbita-sh/orbita/internal/traefik"
)

var (
	ErrDomainNotFound   = errors.New("domain not found")
	ErrDomainTaken      = errors.New("domain already in use")
	ErrDomainNotVerified = errors.New("domain DNS not verified")
)

type DomainService struct {
	domainRepo *repository.DomainRepository
	traefik    *traefik.Manager
}

func NewDomainService(domainRepo *repository.DomainRepository, traefikMgr *traefik.Manager) *DomainService {
	return &DomainService{
		domainRepo: domainRepo,
		traefik:    traefikMgr,
	}
}

func (s *DomainService) AddDomain(ctx context.Context, resourceID uuid.UUID, resourceType, domain string, orgID uuid.UUID, sslEnabled bool, port int) (*models.Domain, error) {
	// Check if domain is already taken
	existing, err := s.domainRepo.FindByDomain(ctx, domain)
	if err == nil && existing != nil {
		if existing.OrganizationID != orgID {
			return nil, ErrDomainTaken
		}
	}

	sslConfig, _ := json.Marshal(map[string]bool{"auto_tls": sslEnabled})

	d := &models.Domain{
		ID:             uuid.New(),
		ResourceID:     resourceID,
		ResourceType:   resourceType,
		OrganizationID: orgID,
		Domain:         domain,
		SSLEnabled:     sslEnabled,
		SSLConfig:      sslConfig,
		Status:         models.DomainStatusPending,
	}

	if err := s.domainRepo.Create(ctx, d); err != nil {
		return nil, fmt.Errorf("AddDomain: %w", err)
	}

	// Update Traefik config
	serviceName := fmt.Sprintf("orbita-%s", resourceID.String()[:8])
	if err := s.traefik.UpsertRoute(traefik.TraefikResource{
		ResourceID:  resourceID,
		Domain:      domain,
		ServiceName: serviceName,
		ServicePort: port,
		SSLEnabled:  sslEnabled,
	}); err != nil {
		// Non-fatal — domain created but Traefik config may be delayed
		d.Status = models.DomainStatusError
		_ = s.domainRepo.Update(ctx, d)
		return d, nil
	}

	d.Status = models.DomainStatusActive
	_ = s.domainRepo.Update(ctx, d)

	return d, nil
}

func (s *DomainService) RemoveDomain(ctx context.Context, domainID, orgID uuid.UUID) error {
	d, err := s.domainRepo.FindByID(ctx, domainID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDomainNotFound
		}
		return fmt.Errorf("RemoveDomain: %w", err)
	}

	if d.OrganizationID != orgID {
		return ErrDomainNotFound
	}

	// Remove Traefik config
	_ = s.traefik.RemoveRoute(d.ResourceID)

	return s.domainRepo.Delete(ctx, domainID)
}

func (s *DomainService) ListDomainsByResource(ctx context.Context, resourceID uuid.UUID, resourceType string) ([]models.Domain, error) {
	return s.domainRepo.ListByResource(ctx, resourceID, resourceType)
}

func (s *DomainService) ListDomainsByOrg(ctx context.Context, orgID uuid.UUID) ([]models.Domain, error) {
	return s.domainRepo.ListByOrgID(ctx, orgID)
}

func (s *DomainService) VerifyDomain(ctx context.Context, domain string) (bool, error) {
	// Check DNS CNAME or A record
	cnames, err := net.LookupCNAME(domain)
	if err == nil && cnames != "" {
		return true, nil
	}

	ips, err := net.LookupIP(domain)
	if err == nil && len(ips) > 0 {
		return true, nil
	}

	return false, nil
}

func (s *DomainService) CheckAndUpdateStatus(ctx context.Context, domainID uuid.UUID) (*models.Domain, error) {
	d, err := s.domainRepo.FindByID(ctx, domainID)
	if err != nil {
		return nil, ErrDomainNotFound
	}

	verified, _ := s.VerifyDomain(ctx, d.Domain)
	d.Verified = verified
	if verified {
		d.Status = models.DomainStatusActive
	}

	_ = s.domainRepo.Update(ctx, d)
	return d, nil
}
