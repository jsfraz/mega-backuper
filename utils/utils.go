package utils

import (
	"encoding/json"
	"fmt"
	"jsfraz/backuper/models"
	"os"
	"os/exec"

	"github.com/go-playground/validator/v10"
)

// Reads JSON from file. Exits wit hstatus 1 on error.
//
//	@return *models.BackupSettings
func LoadJson() *models.BackupSettings {
	// read json from file
	data, err := os.ReadFile("backuper_example.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// unmarshall json
	var settings models.BackupSettings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return &settings
}

// Validates struct. Exits wit hstatus 1 on error.
//
//	@param settings
func ValidateSettings(settings models.BackupSettings) {
	validator := validator.New()
	// validate BackupSettings struct
	err := validator.Struct(settings)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// validate Backup struct
	for _, element := range settings.Backups {
		err = validator.Struct(element)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

// Executes system command.
//
//	@param name
//	@param arg
//	@return error
func Exec(name string, arg ...string) error {
	out, err := exec.Command(name, arg...).Output()
	if err != nil {
		fmt.Println(string(out))
	}
	return err
}
