package utils

import (
	"encoding/json"
	"jsfraz/mega-backuper/models"
	"jsfraz/mega-backuper/validators"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
)

// Reads JSON from file. Exits wit hstatus 1 on error.
//
//	@return *models.BackupSettings
func LoadSettings() models.BackupSettings {
	log.Println("Loading settings..")
	// read json from file
	data, err := os.ReadFile("backuper.json")
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
	return settings
}

// Validates struct. Exits wit hstatus 1 on error.
//
//	@param settings
//	@return error
func ValidateSettings(settings models.BackupSettings) error {
	log.Println("Validating settings..")
	validator := validator.New()
	validators.RegisterBackupSettingsValidators(validator)
	// validate BackupSettings struct
	err := validator.Struct(settings)
	if err != nil {
		return err
	}
	// validate Backup struct
	for _, element := range settings.Backups {
		err = validator.Struct(element)
		if err != nil {
			return err
		}
	}
	log.Println("Settings OK.")
	return nil
}
