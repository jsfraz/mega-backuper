package utils

import (
	"jsfraz/mega-backuper/models"
	"log"
	"os"
)

// Check if volume directories exists or exit program on fail.
func CheckVolumes() {
	// singleton
	s := GetSingleton()

	// cycle trough backups
	for _, backup := range s.Settings.Backups {
		if backup.Type == models.Volume {
			tmp := "/tmp/"
			log.Println("Checking for directory " + tmp + backup.Name + "...")
			// check if dir exists
			if _, err := os.Stat(tmp + backup.Name); os.IsNotExist(err) {
				log.Fatal("Directory does not exists.")
			} else {
				log.Println("Directory exists.")
			}
		}
	}
}
