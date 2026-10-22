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
	// add settings to singleton
	singleton.Settings = settings

	// Check config
	utils.CheckConfig()

	// add mega instance to singleton
	singleton.Mega = mega.New()

	// login or exit
	utils.MegaLogin()

	// task scheduler
	scheduler := gocron.NewScheduler(time.UTC)
	exprDescriptor, descriptorErr := cron.NewDescriptor()
	// iterate trough backups
	log.Println("Initializing jobs...")
	for _, b := range settings.Backups {
		backup := b // creating a new variable to capture the current value

		action, known := backupAction(backup.Type)
		if !known {
			log.Fatalf("Unknown backup type: %s", backup.Type)
		}

		// validate the expression exactly as the scheduler will parse it
		withSeconds, err := utils.ParseSchedule(backup.Cron)
		if err != nil {
			log.Fatalf("Job '%s' has an invalid cron expression: %v", backup.Name, err)
		}

		var jobScheduler *gocron.Scheduler
		if withSeconds {
			jobScheduler = scheduler.CronWithSeconds(backup.Cron)
		} else {
			jobScheduler = scheduler.Cron(backup.Cron)
		}
		// never run the same job twice in parallel
		jobScheduler.SingletonMode()
		if backup.RunOnStart {
			jobScheduler = jobScheduler.StartImmediately()
		}

		if _, err := jobScheduler.Do(handleBackup, backup, action); err != nil {
			log.Fatalf("Failed to schedule job '%s': %v", backup.Name, err)
		}

		description := backup.Cron
		if descriptorErr == nil {
			if desc, err := exprDescriptor.ToDescription(backup.Cron, cron.Locale_en); err == nil {
				description = desc
			}
		}
		runOnStartNote := ""
		if backup.RunOnStart {
			runOnStartNote = " (runs on start)"
		}
		log.Printf("Added [%s] backup job '%s': %s%s", backup.Type, backup.Name, description, runOnStartNote)
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

// backupAction maps a backup type to its backup function.
//
//	@param backupType
//	@return func(models.Backup) error
//	@return bool False when the type is unknown.
func backupAction(backupType models.BackupType) (func(models.Backup) error, bool) {
	switch backupType {
	case models.Volume:
		return utils.BackupVolume, true
	case models.Postgres:
		return utils.BackupPostgres, true
	case models.Mysql:
		return utils.BackupMysql, true
	default:
		return nil, false
	}
}

// Wrapper function for actual backup function. Logs backing up/success/fail.
//
//	@param backup
//	@param backupFunc
func handleBackup(backup models.Backup, backupFunc func(backup models.Backup) error) {
	log.Printf("Running [%s] backup job '%s'...", backup.Type, backup.Name)
	err := backupFunc(backup)
	if err != nil {
		log.Printf("Failed to run [%s] backup job '%s': %s", backup.Type, backup.Name, err)
	} else {
		log.Printf("Successfully ran [%s] backup job '%s'", backup.Type, backup.Name)
	}
}
