package utils

import (
	"encoding/json"
	"jsfraz/mega-backuper/models"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
)

// Reads JSON from file. Exits wit hstatus 1 on error.
//
//	@return *models.BackupSettings
func LoadSettings() *models.BackupSettings {
	log.Println("Loading settings..")
	// read json from file
	// TODO change to backuper.json
	data, err := os.ReadFile("backuper_test.json")
	if err != nil {
		log.Fatal(err)
	}
	// unmarshall json
	var settings models.BackupSettings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Settings loaded.")
	return &settings
}

// TODO validate unique names and volume subdirs
// Validates struct. Exits wit hstatus 1 on error.
//
//	@param settings
func ValidateSettings(settings models.BackupSettings) {
	log.Println("Validating settings..")
	validator := validator.New()
	// validate BackupSettings struct
	err := validator.Struct(settings)
	if err != nil {
		log.Fatal(err)
	}
	// validate Backup struct
	for _, element := range settings.Backups {
		err = validator.Struct(element)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Settings OK.")
}
