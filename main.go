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
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

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
	utils.Login()

	// check volumes
	utils.CheckVolumes()

	// task scheduler
	scheduler := gocron.NewScheduler(time.Local)
	// iterate trough backups
	log.Println("Initializing jobs...")
	for _, b := range settings.Backups {
		backup := b // creating a new variable to capture the current value
		var backupFunc func()
		// volume backup
		if backup.Type == models.Volume {
			backupFunc = func() {
				utils.BackupVolume(backup)
			}
		}
		// mysql dump backup
		if backup.Type == models.Mysql {
			backupFunc = func() {
				// TODO mysql dump backup
				log.Println("TODO mysql dump backup")
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
