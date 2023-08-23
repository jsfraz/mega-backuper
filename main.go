package main

import (
	"jsfraz/mega-backuper/utils"
	"log"

	"github.com/t3rm1n4l/go-mega"
)

func main() {
	// singleton
	s := utils.GetSingleton()
	// log settings
	log.SetPrefix("mega-backuper: ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("Started.")

	// load settings or exit
	settings := utils.LoadSettings()
	// validate struct or exit
	utils.ValidateSettings(*settings)
	s.Settings = settings

	// mega instance
	m := mega.New()
	s.Mega = m

	// login or exit
	utils.Login()

	// check volumes
	utils.CheckVolumes()
}
