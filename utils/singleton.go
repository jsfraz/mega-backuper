package utils

import (
	"jsfraz/mega-backuper/models"
	"sync"

	"github.com/t3rm1n4l/go-mega"
)

type Singleton struct {
	Settings models.BackupSettings
	Mega     *mega.Mega
	// MegaMu serializes access to the Mega API, go-mega is only partially
	// thread-safe and jobs can run in parallel.
	MegaMu sync.Mutex
}

var instance *Singleton

// Returns singleton instance
func GetSingleton() *Singleton {
	if instance == nil {
		instance = new(Singleton)
	}
	return instance
}
