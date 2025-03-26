package utils

// TODOvolumes
/*
// Check if volume directories exists or exit program on fail.
func CheckVolumes() {
	// singleton
	s := GetSingleton()

	// cycle trough backups
	for _, backup := range s.Settings.Backups {
		if backup.Type == models.Volume {
			tmp := "/tmp/"
			log.Printf("Checking for directory %s%s...", tmp, backup.Name)
			// check if dir exists
			if _, err := os.Stat(fmt.Sprintf("%s%s", tmp, backup.Name)); os.IsNotExist(err) {
				log.Fatalf("Directory %s%s does not exists.", tmp, backup.Name)
			}
		}
	}
}
*/
