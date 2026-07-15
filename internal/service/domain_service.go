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
	ErrDomainNotFound    = errors.New("domain not found")
	ErrDomainTaken       = errors.New("domain already in use")
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
		DomainID:    d.ID,
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
	_ = s.traefik.RemoveRoute(d.ResourceID, d.ID)

	return s.domainRepo.Delete(ctx, domainID)
}

// RefreshAppRoutes re-writes the Traefik dynamic config for every domain of an
// app. Called after each successful deploy so routes track the current port
// and survive a wiped config directory.
func (s *DomainService) RefreshAppRoutes(ctx context.Context, appID uuid.UUID, port int) error {
	domains, err := s.domainRepo.ListByResource(ctx, appID, models.ResourceTypeApp)
	if err != nil {
		return fmt.Errorf("RefreshAppRoutes: %w", err)
	}

	serviceName := fmt.Sprintf("orbita-%s", appID.String()[:8])
	for _, d := range domains {
		if err := s.traefik.UpsertRoute(traefik.TraefikResource{
			ResourceID:  d.ResourceID,
			DomainID:    d.ID,
			Domain:      d.Domain,
			ServiceName: serviceName,
			ServicePort: port,
			SSLEnabled:  d.SSLEnabled,
		}); err != nil {
			return fmt.Errorf("RefreshAppRoutes: %s: %w", d.Domain, err)
		}
	}
	return nil
}

// RemoveAppRoutes deletes every Traefik route belonging to an app. Called when
// the app is deleted.
func (s *DomainService) RemoveAppRoutes(ctx context.Context, appID uuid.UUID) error {
	return s.traefik.RemoveResourceRoutes(appID)
}

func (s *DomainService) ListDomainsByResource(ctx context.Context, resourceID uuid.UUID, resourceType string) ([]models.Domain, error) {
	return s.domainRepo.ListByResource(ctx, resourceID, resourceType)
}

func (s *DomainService) ListDomainsByOrg(ctx context.Context, orgID uuid.UUID) ([]models.Domain, error) {
	return s.domainRepo.ListByOrgID(ctx, orgID)
}

// DomainDiagnosis is the result of a DNS/TLS pre-flight check on a domain,
// with an actionable message for the most common Let's Encrypt failure modes.
type DomainDiagnosis struct {
	Domain            string   `json:"domain"`
	Resolves          bool     `json:"resolves"`
	ResolvedIPs       []string `json:"resolved_ips"`
	CloudflareProxied bool     `json:"cloudflare_proxied"`
	Guidance          string   `json:"guidance"`
}

// Cloudflare's published proxy IPv4 ranges (heuristic detection of the
// "orange cloud", which breaks Let's Encrypt HTTP-01 issuance behind Traefik).
var cloudflareCIDRs = []string{
	"173.245.48.0/20", "103.21.244.0/22", "103.22.200.0/22", "103.31.4.0/22",
	"141.101.64.0/18", "108.162.192.0/18", "190.93.240.0/20", "188.114.96.0/20",
	"197.234.240.0/22", "198.41.128.0/17", "162.158.0.0/15", "104.16.0.0/13",
	"104.24.0.0/14", "172.64.0.0/13", "131.0.72.0/22",
}

// DiagnoseDomain resolves the domain and classifies the most common causes of
// TLS issuance failure, returning copy-ready guidance.
func (s *DomainService) DiagnoseDomain(ctx context.Context, domain string) DomainDiagnosis {
	diag := DomainDiagnosis{Domain: domain}

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", domain)
	if err != nil || len(ips) == 0 {
		diag.Guidance = "DNS does not resolve yet. Create an A record pointing " + domain +
			" at your server's IP, then wait for propagation (check with: dig " + domain + " +short). " +
			"Let's Encrypt cannot issue a certificate until DNS resolves."
		return diag
	}

	diag.Resolves = true
	for _, ip := range ips {
		diag.ResolvedIPs = append(diag.ResolvedIPs, ip.String())
		for _, cidr := range cloudflareCIDRs {
			if _, ipnet, err := net.ParseCIDR(cidr); err == nil && ipnet.Contains(ip) {
				diag.CloudflareProxied = true
			}
		}
	}

	if diag.CloudflareProxied {
		diag.Guidance = "This domain resolves to Cloudflare's proxy (orange cloud). Let's Encrypt HTTP-01 " +
			"challenges will fail with a 404. In the Cloudflare DNS tab, switch the record to 'DNS only' " +
			"(gray cloud) at least until the certificate is issued."
		return diag
	}

	diag.Guidance = "DNS resolves. If the certificate still fails to issue: (1) confirm ports 80/443 are open " +
		"(ufw allow 80/tcp && ufw allow 443/tcp), (2) confirm the resolved IP is this server, " +
		"(3) check Let's Encrypt rate limits (5 certificates/week per domain) — see docker compose logs traefik."
	return diag
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
