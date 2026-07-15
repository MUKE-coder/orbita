package service

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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
	backupDir     string
}

func NewDBService(dbRepo *repository.DBRepository, orch *orchestrator.Orchestrator, encryptionKey []byte) *DBService {
	return &DBService{
		dbRepo:        dbRepo,
		orchestrator:  orch,
		encryptionKey: encryptionKey,
		backupDir:     "backups",
	}
}

// SetBackupDir sets where backup archives are written (default "backups").
func (s *DBService) SetBackupDir(dir string) {
	if dir != "" {
		s.backupDir = dir
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

// EnvDatabaseURLs returns `<NAME>_URL -> connection string` for every managed
// database in an environment. Injected into app deploys so apps get their
// database credentials without copy-pasting connection strings.
func (s *DBService) EnvDatabaseURLs(ctx context.Context, environmentID, orgID uuid.UUID) (map[string]string, error) {
	dbs, err := s.dbRepo.ListByEnvironment(ctx, environmentID, orgID)
	if err != nil {
		return nil, fmt.Errorf("EnvDatabaseURLs: %w", err)
	}

	orgKey, err := auth.DeriveOrgKey(s.encryptionKey, orgID)
	if err != nil {
		return nil, fmt.Errorf("EnvDatabaseURLs: derive key: %w", err)
	}

	urls := make(map[string]string, len(dbs))
	for _, mdb := range dbs {
		if mdb.ConnectionConfig == nil || *mdb.ConnectionConfig == "" {
			continue
		}
		conn, err := auth.Decrypt(*mdb.ConnectionConfig, orgKey)
		if err != nil {
			// Legacy rows may hold plaintext connection strings
			conn = *mdb.ConnectionConfig
		}
		urls[envVarNameForDB(mdb.Name)] = conn
	}
	return urls, nil
}

// envVarNameForDB converts a database name into its injected env var name:
// "rental-db" -> "RENTAL_DB_URL".
func envVarNameForDB(name string) string {
	upper := strings.ToUpper(name)
	var b strings.Builder
	for _, r := range upper {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String() + "_URL"
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

	fail := func(err error) (*models.Backup, error) {
		backup.Status = models.BackupStatusFailed
		_ = s.dbRepo.UpdateBackup(ctx, backup)
		return backup, fmt.Errorf("CreateBackup: %w", err)
	}

	// Dump via the engine's tool inside the DB container
	connStr, err := s.GetConnectionString(ctx, dbID, orgID)
	if err != nil {
		return fail(err)
	}
	dump, err := s.orchestrator.DumpDatabase(ctx, mdb, connStr)
	if err != nil {
		return fail(err)
	}

	// Gzip to <backupDir>/<orgID>/<dbID>/<backupID>.gz
	dir := filepath.Join(s.backupDir, orgID.String(), mdb.ID.String())
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fail(err)
	}
	storagePath := filepath.Join(dir, backup.ID.String()+".gz")

	f, err := os.Create(storagePath)
	if err != nil {
		return fail(err)
	}
	gz := gzip.NewWriter(f)
	if _, err := gz.Write(dump); err != nil {
		f.Close()
		return fail(err)
	}
	if err := gz.Close(); err != nil {
		f.Close()
		return fail(err)
	}
	if err := f.Close(); err != nil {
		return fail(err)
	}

	info, err := os.Stat(storagePath)
	if err != nil {
		return fail(err)
	}

	backup.Status = models.BackupStatusCompleted
	backup.SizeBytes = info.Size()
	backup.StoragePath = &storagePath
	_ = s.dbRepo.UpdateBackup(ctx, backup)

	return backup, nil
}

func (s *DBService) ListBackups(ctx context.Context, dbID uuid.UUID, limit int) ([]models.Backup, error) {
	return s.dbRepo.ListBackups(ctx, dbID, limit)
}

func (s *DBService) RestoreBackup(ctx context.Context, dbID, backupID, orgID uuid.UUID) error {
	mdb, err := s.dbRepo.FindByID(ctx, dbID, orgID)
	if err != nil {
		return ErrDBNotFound
	}

	backup, err := s.dbRepo.FindBackupByID(ctx, backupID)
	if err != nil {
		return ErrBackupNotFound
	}

	if backup.SourceID != dbID || backup.OrganizationID != orgID {
		return ErrBackupNotFound
	}
	if backup.Status != models.BackupStatusCompleted || backup.StoragePath == nil {
		return fmt.Errorf("RestoreBackup: backup %s is not restorable (status %s)", backupID, backup.Status)
	}

	f, err := os.Open(*backup.StoragePath)
	if err != nil {
		return fmt.Errorf("RestoreBackup: open archive: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("RestoreBackup: gunzip: %w", err)
	}
	dump, err := io.ReadAll(gz)
	if err != nil {
		return fmt.Errorf("RestoreBackup: read archive: %w", err)
	}

	connStr, err := s.GetConnectionString(ctx, dbID, orgID)
	if err != nil {
		return fmt.Errorf("RestoreBackup: %w", err)
	}

	if err := s.orchestrator.RestoreDatabase(ctx, mdb, connStr, dump); err != nil {
		return fmt.Errorf("RestoreBackup: %w", err)
	}
	return nil
}

func (s *DBService) GetBackupSchedule(ctx context.Context, dbID uuid.UUID) (*models.BackupSchedule, error) {
	return s.dbRepo.GetBackupSchedule(ctx, dbID)
}

// StartBackupScheduler runs scheduled backups. It checks for due schedules
// once a minute until ctx is cancelled. Call from main in a goroutine.
func (s *DBService) StartBackupScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	log.Info().Msg("Backup scheduler started")
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Backup scheduler stopped")
			return
		case <-ticker.C:
			s.runDueBackups(ctx)
		}
	}
}

func (s *DBService) runDueBackups(ctx context.Context) {
	due, err := s.dbRepo.ListDueBackupSchedules(ctx, time.Now())
	if err != nil {
		log.Error().Err(err).Msg("Backup scheduler: list due schedules")
		return
	}

	for i := range due {
		bs := &due[i]

		if _, err := s.CreateBackup(ctx, bs.SourceID, bs.OrganizationID); err != nil {
			log.Error().Err(err).Str("db_id", bs.SourceID.String()).Msg("Scheduled backup failed")
			// fall through: still advance next_run_at so we don't hot-loop a broken DB
		}

		now := time.Now()
		next := nextBackupTime(bs.Frequency, now)
		bs.LastRunAt = &now
		bs.NextRunAt = &next
		if err := s.dbRepo.UpdateBackupSchedule(ctx, bs); err != nil {
			log.Error().Err(err).Str("db_id", bs.SourceID.String()).Msg("Backup scheduler: update schedule")
		}

		s.pruneBackups(ctx, bs.SourceID, bs.RetentionCount)
	}
}

// pruneBackups keeps only the newest `retain` completed backups for a source,
// deleting older archives from disk and their rows.
func (s *DBService) pruneBackups(ctx context.Context, sourceID uuid.UUID, retain int) {
	if retain <= 0 {
		return
	}
	backups, err := s.dbRepo.ListBackups(ctx, sourceID, 1000)
	if err != nil {
		log.Error().Err(err).Msg("Backup prune: list")
		return
	}
	kept := 0
	for _, b := range backups {
		if b.Status != models.BackupStatusCompleted {
			continue
		}
		kept++
		if kept <= retain {
			continue
		}
		if b.StoragePath != nil {
			if err := os.Remove(*b.StoragePath); err != nil && !os.IsNotExist(err) {
				log.Warn().Err(err).Str("path", *b.StoragePath).Msg("Backup prune: remove file")
			}
		}
		if err := s.dbRepo.DeleteBackup(ctx, b.ID); err != nil {
			log.Warn().Err(err).Str("backup_id", b.ID.String()).Msg("Backup prune: delete row")
		}
	}
}

func nextBackupTime(frequency string, from time.Time) time.Time {
	switch frequency {
	case "hourly":
		return from.Add(time.Hour)
	case "weekly":
		return from.Add(7 * 24 * time.Hour)
	default: // daily
		return from.Add(24 * time.Hour)
	}
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
