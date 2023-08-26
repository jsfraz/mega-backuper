package utils

import (
	"jsfraz/mega-backuper/models"

	"github.com/t3rm1n4l/go-mega"
)

type Singleton struct {
	Settings models.BackupSettings
	Mega     *mega.Mega
}

var instance *Singleton

// Returns singleton instance
func GetSingleton() *Singleton {
	if instance == nil {
		instance = new(Singleton)
	}
	return instance
}
