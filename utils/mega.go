package utils

import (
	"log"
)

// Mega login
func Login() {
	s := GetSingleton()
	log.Println("Logging in...")
	// login
	err := s.Mega.Login(s.Settings.Email, s.Settings.Password)
	if err != nil {
		log.Fatalln(err)
	}
	// get logged user
	user, err := s.Mega.GetUser()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Logged as " + user.Name + ".")
}
