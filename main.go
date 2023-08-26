package main

import (
	"jsfraz/mega-backuper/models"
	"jsfraz/mega-backuper/utils"
	"log"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/t3rm1n4l/go-mega"
)

func main() {
	// singleton
	singleton := utils.GetSingleton()
	// log settings
	log.SetPrefix("mega-backuper: ")
	log.SetFlags(log.LstdFlags | log.LUTC)

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

	// check volumes
	utils.CheckVolumes()

	// task scheduler
	scheduler := gocron.NewScheduler(time.UTC)
	// iterate trough backups
	log.Println("Initializing jobs...")
	for _, b := range settings.Backups {
		backup := b // creating a new variable to capture the current value
		var backupFunc func()
		// volume backup
		if backup.Type == models.Volume {
			backupFunc = func() {
				handleBackup(backup, utils.BackupVolume)
			}
		}
		// mysql dump backup
		if backup.Type == models.Mysql {
			backupFunc = func() {
				// TODO mysql dump backup
				handleBackup(backup, func(backup models.Backup) error {
					log.Println("TODO mysql dump backup")
					return nil
				})
			}
		}
		scheduler.Cron(backup.Cron).Do(backupFunc)
		log.Println("Added [" + string(backup.Type) + "] backup job '" + backup.Name + "'.")
	}
	// check if job list is empty or not
	if len(scheduler.Jobs()) != 0 {
		// start blocking
		log.Println("Started scheduler. Total jobs: " + strconv.Itoa(len(scheduler.Jobs())))
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
	log.Println("Backing up [" + string(backup.Type) + "] backup job '" + backup.Name + "'...")
	err := backupFunc(backup)
	if err != nil {
		log.Println("Failed to backup ["+string(backup.Type)+"] backup job '"+backup.Name+"': ", err)
	} else {
		log.Println("Successfully backed up [" + string(backup.Type) + "] backup job '" + backup.Name + "'.")
	}
}
