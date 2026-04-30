package nexus

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s *Store) GetSyncStatus() (SyncStatus, error) {
	status := SyncStatus{
		DatabasePath: s.path,
		UpdatedAt:    nowUTC(),
	}
	if err := s.db.QueryRow(`
		SELECT
			COUNT(CASE WHEN status = 'pending' THEN 1 END),
			COUNT(CASE WHEN status = 'synced' THEN 1 END),
			COUNT(CASE WHEN status = 'failed' THEN 1 END)
		FROM sync_log
	`).Scan(&status.PendingCount, &status.SyncedCount, &status.FailedCount); err != nil {
		return status, err
	}
	var oldestPending, lastSynced, lastError sql.NullString
	if err := s.db.QueryRow("SELECT MIN(created_at) FROM sync_log WHERE status = 'pending'").Scan(&oldestPending); err != nil {
		return status, err
	}
	if err := s.db.QueryRow("SELECT MAX(synced_at) FROM sync_log WHERE status = 'synced'").Scan(&lastSynced); err != nil {
		return status, err
	}
	err := s.db.QueryRow("SELECT last_error FROM sync_log WHERE status = 'failed' AND last_error != '' ORDER BY created_at DESC LIMIT 1").Scan(&lastError)
	if err != nil && err != sql.ErrNoRows {
		return status, err
	}
	if oldestPending.Valid {
		status.OldestPendingAt = oldestPending.String
	}
	if lastSynced.Valid {
		status.LastSyncedAt = lastSynced.String
	}
	if lastError.Valid {
		status.LastError = lastError.String
	}
	if info, err := os.Stat(s.path); err == nil {
		status.DatabaseBytes = info.Size()
	}
	if info, err := os.Stat(s.path + "-wal"); err == nil {
		status.WALBytes = info.Size()
	}
	return status, nil
}

func (s *Store) ExportBackup(input BackupInput) (ExportResult, error) {
	dir := strings.TrimSpace(input.Destination)
	if dir == "" {
		_ = s.db.QueryRow("SELECT backup_path FROM restaurant_settings WHERE id = 'primary'").Scan(&dir)
	}
	if strings.TrimSpace(dir) == "" {
		base, err := exportDir()
		if err != nil {
			return ExportResult{}, err
		}
		dir = filepath.Join(base, "backups")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ExportResult{}, err
	}
	if _, err := s.db.Exec("PRAGMA wal_checkpoint(FULL)"); err != nil {
		return ExportResult{}, err
	}

	fileName := safeFileName("nexus_backup_"+time.Now().UTC().Format("20060102_150405")) + ".db"
	path := filepath.Join(dir, fileName)
	if err := copyFile(s.path, path); err != nil {
		return ExportResult{}, err
	}
	result, err := exportResult("sqlite_backup", path, "application/x-sqlite3")
	if err != nil {
		return ExportResult{}, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return ExportResult{}, err
	}
	defer rollback(tx)
	now := nowUTC()
	if err := insertAuditLogTx(tx, "backup_created", "database", filepath.Base(path), "Local SQLite backup exported", now); err != nil {
		return ExportResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return ExportResult{}, err
	}
	return result, nil
}

func copyFile(source, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
