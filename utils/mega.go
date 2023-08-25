package utils

import (
	"log"
)

// Login to Mega or exit  program on fail.
func Login() {
	singleton := GetSingleton()
	log.Println("Logging in...")
	// login
	err := singleton.Mega.Login(singleton.Settings.Email, singleton.Settings.Password)
	if err != nil {
		log.Fatalln(err)
	}
	// get logged user
	user, err := singleton.Mega.GetUser()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Logged as " + user.Name + ".")
}
