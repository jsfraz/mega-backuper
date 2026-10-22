package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSweepStaleTempFiles(t *testing.T) {
	dir := t.TempDir()
	old := time.Now().Add(-4 * time.Hour)
	cutoff := time.Now().Add(-staleTempMaxAge)

	staleArchive := createTestFile(t, dir, "pg-1700000000.backup", old)
	staleWorkDir := filepath.Join(dir, workDirPrefix+"pg-123")
	if err := os.Mkdir(staleWorkDir, 0700); err != nil {
		t.Fatalf("failed to create working directory: %v", err)
	}
	setModTime(t, staleWorkDir, old)

	freshArchive := createTestFile(t, dir, "pg-1700000001.backup", time.Now())
	unrelated := createTestFile(t, dir, "notes.txt", old)
	unrelatedDir := filepath.Join(dir, "other")
	if err := os.Mkdir(unrelatedDir, 0700); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	setModTime(t, unrelatedDir, old)

	sweepStaleTempFiles(dir, cutoff)

	if _, err := os.Stat(staleArchive); !os.IsNotExist(err) {
		t.Errorf("stale archive %s was not removed", staleArchive)
	}
	if _, err := os.Stat(staleWorkDir); !os.IsNotExist(err) {
		t.Errorf("stale working directory %s was not removed", staleWorkDir)
	}
	if _, err := os.Stat(freshArchive); err != nil {
		t.Errorf("fresh archive %s was removed: %v", freshArchive, err)
	}
	if _, err := os.Stat(unrelated); err != nil {
		t.Errorf("unrelated file %s was removed: %v", unrelated, err)
	}
	if _, err := os.Stat(unrelatedDir); err != nil {
		t.Errorf("unrelated directory %s was removed: %v", unrelatedDir, err)
	}
}

func createTestFile(t *testing.T, dir string, name string, modTime time.Time) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("data"), 0600); err != nil {
		t.Fatalf("failed to create %s: %v", path, err)
	}
	setModTime(t, path, modTime)
	return path
}

func setModTime(t *testing.T, path string, modTime time.Time) {
	t.Helper()

	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("failed to set modification time of %s: %v", path, err)
	}
}
