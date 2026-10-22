package utils

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// workDirPrefix is the prefix of per-run temporary directories.
	workDirPrefix = "mega-backuper-"
	// staleTempMaxAge is how old a leftover file must be before it is swept.
	// It is longer than dumpTimeout+shutdownGrace so a concurrently running
	// instance is never touched.
	staleTempMaxAge = 3 * time.Hour
)

var staleArchiveRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*-\d+\.(?:tar\.gz|backup)$`)

// SweepStaleTempFiles removes archives and working directories left behind by
// an interrupted previous run.
func SweepStaleTempFiles() {
	sweepStaleTempFiles(os.TempDir(), time.Now().Add(-staleTempMaxAge))
}

// sweepStaleTempFiles removes known mega-backuper leftovers older than cutoff.
func sweepStaleTempFiles(dir string, cutoff time.Time) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to read temporary directory %s: %v", dir, err)
		}
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		isArchive := !entry.IsDir() && staleArchiveRe.MatchString(name)
		isWorkDir := entry.IsDir() && strings.HasPrefix(name, workDirPrefix)
		if !isArchive && !isWorkDir {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			log.Printf("Failed to inspect temporary file %s: %v", filepath.Join(dir, name), err)
			continue
		}
		if !info.ModTime().Before(cutoff) {
			continue
		}

		path := filepath.Join(dir, name)
		if err := os.RemoveAll(path); err != nil {
			log.Printf("Failed to remove stale temporary file %s: %v", path, err)
			continue
		}
		log.Printf("Removed stale temporary file %s", path)
	}
}
