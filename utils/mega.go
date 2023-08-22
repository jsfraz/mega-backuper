package utils

import "jsfraz/mega-backuper/models"

// Does Mega login.
//
//	@param email
//	@param password
//	@return error
func MegaLogin(settings models.BackupSettings) error {
	return Exec("mega-login", settings.MegaEmail, settings.MegaPassword)
}
