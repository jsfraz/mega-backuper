package main

import (
	"jsfraz/mega-backuper/models"
	"jsfraz/mega-backuper/utils"
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/t3rm1n4l/go-mega"
)

func main() {
	// singleton
	singleton := utils.GetSingleton()
	// log settings
	log.SetPrefix("mega-backuper: ")
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmicroseconds)

	log.Println("Started.")

	// load settings or exit
	settings := utils.LoadSettings()
	// validate struct or exit
	utils.ValidateSettings(settings)
	// add settinsg to singleton
	singleton.Settings = settings

	// add mega instance to singleton
	singleton.Mega = mega.New()

	// login or exit
	utils.MegaLogin()

	// Check volumes
	// TODO volumes
	// utils.CheckVolumes()

	// task scheduler
	scheduler := gocron.NewScheduler(time.UTC)
	// iterate trough backups
	log.Println("Initializing jobs...")
	for _, b := range settings.Backups {
		backup := b // creating a new variable to capture the current value
		var backupFunc func()
		// TODO volumes
		/*
			// volume backup
			if backup.Type == models.Volume {
				backupFunc = func() {
					bck := backup
					handleBackup(bck, utils.BackupVolume)
				}
			}
		*/
		// Postgres dump backup
		if backup.Type == models.Postgres {
			backupFunc = func() {
				bck := backup
				handleBackup(bck, utils.BackupPostgres)
			}
		}
		// mysql dump backup
		// TODO mysql
		/*
			if backup.Type == models.Mysql {
				backupFunc = func() {
					// TODO mysql dump backup
					bck := backup
					handleBackup(bck, utils.BackupMysql)
				}
			}
		*/
		scheduler.Cron(backup.Cron).Do(backupFunc)
		log.Printf("Added [%s] backup job '%s'", backup.Type, backup.Name)
	}
	// check if job list is empty or not
	if len(scheduler.Jobs()) != 0 {
		// start blocking
		log.Printf("Started scheduler. Total jobs: %d", len(scheduler.Jobs()))
		scheduler.StartBlocking()
	} else {
		// exit
		log.Fatalln("No backup job was set.")
	}
}

// Wrapper function for actual backup function. Logs backing up/succes/fail.
//
//	@param backup
//	@param backupFunc
func handleBackup(backup models.Backup, backupFunc func(backup models.Backup) error) {
	log.Printf("Backing up [%s] backup job '%s'...", backup.Type, backup.Name)
	err := backupFunc(backup)
	if err != nil {
		log.Printf("Failed to backup [%s] backup job '%s': %s", backup.Type, backup.Name, err)
	} else {
		log.Printf("Successfully backed up [%s] backup job '%s'", backup.Type, backup.Name)
	}
}
