package utils

import (
	"encoding/json"
	"jsfraz/mega-backuper/models"
	"jsfraz/mega-backuper/validators"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
)

// Reads JSON from file. Exits with status 1 on error.
//
//	@return models.BackupSettings
func LoadSettings() models.BackupSettings {
	log.Println("Loading settings..")
	// read json from file
	data, err := os.ReadFile("backuper.json")
	if err != nil {
		log.Fatalf("Failed to read backuper.json (is the config mounted?): %v", err)
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

// Validates struct.
//
//	@param settings
//	@return error
func ValidateSettings(settings models.BackupSettings) error {
	log.Println("Validating settings..")
	v := validator.New()
	if err := validators.RegisterBackupSettingsValidators(v); err != nil {
		return err
	}
	// validate BackupSettings struct
	if err := v.Struct(settings); err != nil {
		return err
	}
	// validate Backup struct
	for _, element := range settings.Backups {
		if err := v.Struct(element); err != nil {
			return err
		}
	}
	log.Println("Settings OK.")
	return nil
}
