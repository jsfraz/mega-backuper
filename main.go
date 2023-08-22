package main

import (
	"jsfraz/backuper/utils"
)

func main() {
	// read json or exit
	settings := utils.LoadJson()
	// validate struct or exit
	utils.ValidateSettings(*settings)

	// mega login
	_ = utils.MegaLogin(*settings)
}
