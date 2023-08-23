package utils

import (
	"jsfraz/mega-backuper/models"

	"github.com/t3rm1n4l/go-mega"
)

// https://blog.devgenius.io/singleton-pattern-in-go-4faea607ad0f
type Singleton struct {
	Mega     *mega.Mega
	Settings *models.BackupSettings
}

var instance *Singleton

// Returns instance
func GetSingleton() *Singleton {
	if instance == nil {
		instance = new(Singleton)
	}
	return instance
}
