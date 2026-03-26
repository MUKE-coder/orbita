package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/orchestrator"
	"github.com/orbita-sh/orbita/internal/repository"
)

var (
	ErrDBNotFound     = errors.New("database not found")
	ErrBackupNotFound = errors.New("backup not found")
)

type DBService struct {
	dbRepo        *repository.DBRepository
	orchestrator  *orchestrator.Orchestrator
	encryptionKey []byte
}

func NewDBService(dbRepo *repository.DBRepository, orch *orchestrator.Orchestrator, encryptionKey []byte) *DBService {
	return &DBService{
		dbRepo:        dbRepo,
		orchestrator:  orch,
		encryptionKey: encryptionKey,
	}
}

type CreateDBInput struct {
	Name          string    `json:"name"`
	Engine        string    `json:"engine"`
	Version       string    `json:"version"`
	EnvironmentID uuid.UUID `json:"environment_id"`
	CPULimit      int       `json:"cpu_limit"`
	MemoryLimit   int       `json:"memory_limit"`
}

func (s *DBService) CreateDatabase(ctx context.Context, orgID uuid.UUID, orgSlug string, input CreateDBInput) (*models.ManagedDatabase, error) {
	mdb := &models.ManagedDatabase{
		ID:             uuid.New(),
		EnvironmentID:  input.EnvironmentID,
		OrganizationID: orgID,
		Name:           input.Name,
		Engine:         input.Engine,
		Version:        input.Version,
		Status:         models.DBStatusCreating,
		CPULimit:       input.CPULimit,
		MemoryLimit:    input.MemoryLimit,
	}

	if err := s.dbRepo.Create(ctx, mdb); err != nil {
		return nil, fmt.Errorf("CreateDatabase: %w", err)
	}

	// Provision
	if err := s.orchestrator.ProvisionDatabase(ctx, mdb, orgSlug); err != nil {
		mdb.Status = models.DBStatusFailed
		_ = s.dbRepo.Update(ctx, mdb)
		return mdb, fmt.Errorf("CreateDatabase: provision: %w", err)
	}

	// Encrypt connection config
	if mdb.ConnectionConfig != nil {
		orgKey, err := auth.DeriveOrgKey(s.encryptionKey, orgID)
		if err == nil {
			encrypted, err := auth.Encrypt(*mdb.ConnectionConfig, orgKey)
			if err == nil {
				mdb.ConnectionConfig = &encrypted
			}
		}
	}

	_ = s.dbRepo.Update(ctx, mdb)
	return mdb, nil
}

func (s *DBService) GetDatabase(ctx context.Context, id, orgID uuid.UUID) (*models.ManagedDatabase, error) {
	mdb, err := s.dbRepo.FindByID(ctx, id, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDBNotFound
		}
		return nil, fmt.Errorf("GetDatabase: %w", err)
	}
	return mdb, nil
}

func (s *DBService) GetConnectionString(ctx context.Context, id, orgID uuid.UUID) (string, error) {
	mdb, err := s.dbRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return "", ErrDBNotFound
	}
	if mdb.ConnectionConfig == nil {
		return "", nil
	}

	orgKey, err := auth.DeriveOrgKey(s.encryptionKey, orgID)
	if err != nil {
		return "", fmt.Errorf("GetConnectionString: derive key: %w", err)
	}

	decrypted, err := auth.Decrypt(*mdb.ConnectionConfig, orgKey)
	if err != nil {
		// May not be encrypted (legacy), return as-is
		return *mdb.ConnectionConfig, nil
	}
	return decrypted, nil
}

func (s *DBService) ListDatabases(ctx context.Context, orgID uuid.UUID) ([]models.ManagedDatabase, error) {
	return s.dbRepo.ListByOrgID(ctx, orgID)
}

func (s *DBService) DeleteDatabase(ctx context.Context, id, orgID uuid.UUID) error {
	mdb, err := s.dbRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return ErrDBNotFound
	}
	if err := s.orchestrator.RemoveDatabase(ctx, mdb); err != nil {
		return fmt.Errorf("DeleteDatabase: %w", err)
	}
	return s.dbRepo.Delete(ctx, id, orgID)
}

func (s *DBService) RestartDatabase(ctx context.Context, id, orgID uuid.UUID) error {
	mdb, err := s.dbRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return ErrDBNotFound
	}
	if err := s.orchestrator.RestartDatabase(ctx, mdb); err != nil {
		return fmt.Errorf("RestartDatabase: %w", err)
	}
	mdb.Status = models.DBStatusRunning
	return s.dbRepo.Update(ctx, mdb)
}

func (s *DBService) StopDatabase(ctx context.Context, id, orgID uuid.UUID) error {
	mdb, err := s.dbRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return ErrDBNotFound
	}
	if err := s.orchestrator.StopDatabase(ctx, mdb); err != nil {
		return fmt.Errorf("StopDatabase: %w", err)
	}
	mdb.Status = models.DBStatusStopped
	return s.dbRepo.Update(ctx, mdb)
}

func (s *DBService) StartDatabase(ctx context.Context, id, orgID uuid.UUID) error {
	mdb, err := s.dbRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return ErrDBNotFound
	}
	if err := s.orchestrator.StartDatabase(ctx, mdb); err != nil {
		return fmt.Errorf("StartDatabase: %w", err)
	}
	mdb.Status = models.DBStatusRunning
	return s.dbRepo.Update(ctx, mdb)
}

// Backups

func (s *DBService) CreateBackup(ctx context.Context, dbID, orgID uuid.UUID) (*models.Backup, error) {
	mdb, err := s.dbRepo.FindByID(ctx, dbID, orgID)
	if err != nil {
		return nil, ErrDBNotFound
	}

	backup := &models.Backup{
		ID:             uuid.New(),
		SourceID:       mdb.ID,
		SourceType:     "database",
		OrganizationID: orgID,
		Status:         models.BackupStatusRunning,
	}

	if err := s.dbRepo.CreateBackup(ctx, backup); err != nil {
		return nil, fmt.Errorf("CreateBackup: %w", err)
	}

	// TODO: real impl — run pg_dump/mysqldump/mongodump in sidecar container
	backup.Status = models.BackupStatusCompleted
	backup.SizeBytes = 1024 * 1024 // stub: 1MB
	storagePath := fmt.Sprintf("/backups/%s/%s/%s.gz", orgID, mdb.ID, backup.ID)
	backup.StoragePath = &storagePath

	_ = s.dbRepo.UpdateBackup(ctx, backup)

	return backup, nil
}

func (s *DBService) ListBackups(ctx context.Context, dbID uuid.UUID, limit int) ([]models.Backup, error) {
	return s.dbRepo.ListBackups(ctx, dbID, limit)
}

func (s *DBService) RestoreBackup(ctx context.Context, dbID, backupID, orgID uuid.UUID) error {
	_, err := s.dbRepo.FindByID(ctx, dbID, orgID)
	if err != nil {
		return ErrDBNotFound
	}

	backup, err := s.dbRepo.FindBackupByID(ctx, backupID)
	if err != nil {
		return ErrBackupNotFound
	}

	if backup.SourceID != dbID {
		return ErrBackupNotFound
	}

	// TODO: real impl — restore from backup file
	_ = backup
	return nil
}

func (s *DBService) GetBackupSchedule(ctx context.Context, dbID uuid.UUID) (*models.BackupSchedule, error) {
	return s.dbRepo.GetBackupSchedule(ctx, dbID)
}

func (s *DBService) SetBackupSchedule(ctx context.Context, dbID, orgID uuid.UUID, frequency string, retentionCount int) (*models.BackupSchedule, error) {
	_, err := s.dbRepo.FindByID(ctx, dbID, orgID)
	if err != nil {
		return nil, ErrDBNotFound
	}

	var nextRun time.Time
	switch frequency {
	case "hourly":
		nextRun = time.Now().Add(1 * time.Hour)
	case "daily":
		nextRun = time.Now().Add(24 * time.Hour)
	case "weekly":
		nextRun = time.Now().Add(7 * 24 * time.Hour)
	default:
		nextRun = time.Now().Add(24 * time.Hour)
	}

	bs := &models.BackupSchedule{
		ID:             uuid.New(),
		SourceID:       dbID,
		SourceType:     "database",
		OrganizationID: orgID,
		Frequency:      frequency,
		RetentionCount: retentionCount,
		Enabled:        true,
		NextRunAt:      &nextRun,
	}

	if err := s.dbRepo.UpsertBackupSchedule(ctx, bs); err != nil {
		return nil, fmt.Errorf("SetBackupSchedule: %w", err)
	}

	return bs, nil
}
