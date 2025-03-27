package validators

import (
	"jsfraz/mega-backuper/models"

	"github.com/go-playground/validator/v10"
)

// Validator for BackupSettings struct.
//
//	@param v
func RegisterBackupSettingsValidators(v *validator.Validate) {
	v.RegisterStructValidation(func(sl validator.StructLevel) {
		settings, ok := sl.Current().Interface().(models.BackupSettings)
		if !ok {
			return
		}

		names := make(map[string]struct{})
		for _, backup := range settings.Backups {
			if _, exists := names[backup.Name]; exists {
				sl.ReportError(settings.Backups, "Backups", "backups", "unique_names", "")
				return
			}
			names[backup.Name] = struct{}{}
		}
	}, models.BackupSettings{})
}
