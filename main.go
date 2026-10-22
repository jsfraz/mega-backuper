package main

import (
	"jsfraz/mega-backuper/models"
	"jsfraz/mega-backuper/utils"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/lnquy/cron"
	"github.com/t3rm1n4l/go-mega"
)

// shutdownGrace is how long a SIGTERM/SIGINT waits for running backups.
const shutdownGrace = 30 * time.Second

// inFlight counts backup jobs currently running.
var inFlight atomic.Int32

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
	if err := utils.ValidateSettings(settings); err != nil {
		log.Fatalln(err)
	}
	// add settings to singleton
	singleton.Settings = settings

	// Check config
	if err := utils.CheckConfig(); err != nil {
		log.Fatalln(err)
	}

	// add mega instance to singleton
	singleton.Mega = mega.New()

	// login or exit
	utils.MegaLogin()

	// remove leftovers from an interrupted previous run
	utils.SweepStaleTempFiles()

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
	if len(scheduler.Jobs()) == 0 {
		// exit
		log.Fatalln("No backup job was set.")
	}

	log.Printf("Started scheduler. Total jobs: %d", len(scheduler.Jobs()))
	scheduler.StartAsync()

	// graceful shutdown on container stop
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	received := <-stop
	log.Printf("Received %s, stopping scheduler (waiting up to %s for running jobs)...", received, shutdownGrace)
	scheduler.Stop()

	deadline := time.Now().Add(shutdownGrace)
	for time.Now().Before(deadline) {
		// inFlight covers our own counter, IsRunning closes the window between
		// gocron dispatching a job and handleBackup incrementing the counter
		if inFlight.Load() == 0 && !anyJobRunning(scheduler.Jobs()) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if remaining := inFlight.Load(); remaining > 0 || anyJobRunning(scheduler.Jobs()) {
		log.Printf("Shutdown grace period expired with %d backup job(s) still running.", remaining)
		os.Exit(1)
	}
	log.Println("Shutdown complete.")
}

// anyJobRunning reports whether any scheduled job is currently executing.
//
//	@param jobs
//	@return bool
func anyJobRunning(jobs []*gocron.Job) bool {
	for _, job := range jobs {
		if job.IsRunning() {
			return true
		}
	}
	return false
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
	inFlight.Add(1)
	defer inFlight.Add(-1)

	err := backupFunc(backup)
	if err != nil {
		log.Printf("Failed to run [%s] backup job '%s': %s", backup.Type, backup.Name, err)
	} else {
		log.Printf("Successfully ran [%s] backup job '%s'", backup.Type, backup.Name)
	}
}
