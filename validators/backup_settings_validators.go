package validators

import (
	"jsfraz/mega-backuper/models"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// safeSegmentRe matches values that can be used as a single path segment
// (backup names and database names are used in /tmp paths and file names).
var safeSegmentRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// validateSafeSegment rejects path separators, traversal and leading dashes.
// Empty values are left to the required/required_if validators.
func validateSafeSegment(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	return safeSegmentRe.MatchString(value)
}

// Registers custom validators for BackupSettings struct.
//
//	@param v
//	@return error
func RegisterBackupSettingsValidators(v *validator.Validate) error {
	if err := v.RegisterValidation("safesegment", validateSafeSegment); err != nil {
		return err
	}

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

	return nil
}
