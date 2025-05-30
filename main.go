package main

import (
	"jsfraz/mega-backuper/models"
	"jsfraz/mega-backuper/utils"
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/lnquy/cron"
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
	err := utils.ValidateSettings(settings)
	if err != nil {
		log.Fatalln(err)
	}
	// add settinsg to singleton
	singleton.Settings = settings

	// Check config
	utils.CheckConfig()

	// add mega instance to singleton
	singleton.Mega = mega.New()

	// login or exit
	utils.MegaLogin()

	// task scheduler
	scheduler := gocron.NewScheduler(time.UTC)
	// iterate trough backups
	log.Println("Initializing jobs...")
	for _, b := range settings.Backups {
		backup := b // creating a new variable to capture the current value
		var backupFunc func()
		switch backup.Type {

		// Volume backup
		case models.Volume:
			backupFunc = func() {
				bck := backup
				handleBackup(bck, utils.BackupVolume)
			}
			break

		// Postgres dump backup
		case models.Postgres:
			backupFunc = func() {
				bck := backup
				handleBackup(bck, utils.BackupPostgres)
			}
			break

			// TODO MySQL dump backup
			/*
				case models.Mysql:
					backupFunc = func() {
						// TODO mysql dump backup
						bck := backup
						handleBackup(bck, utils.BackupMysql)
					}
					break
			*/
		}
		scheduler.Cron(backup.Cron).Do(backupFunc)
		exprDesc, _ := cron.NewDescriptor()
		desc, _ := exprDesc.ToDescription(backup.Cron, cron.Locale_en)
		log.Printf("Added [%s] backup job '%s': %s", backup.Type, backup.Name, desc)
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
	log.Printf("Running [%s] backup job '%s'...", backup.Type, backup.Name)
	err := backupFunc(backup)
	if err != nil {
		log.Printf("Failed to run [%s] backup job '%s': %s", backup.Type, backup.Name, err)
	} else {
		log.Printf("Successfully ran up [%s] backup job '%s'", backup.Type, backup.Name)
	}
}
