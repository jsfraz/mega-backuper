package utils

import (
	"log"

	"github.com/t3rm1n4l/go-mega"
)

// Login to Mega or exit  program on fail.
func MegaLogin() {
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

// Check if Mega directory exists. If not, create it. Returns directory's node.
//
//	@param megaDir
//	@return *mega.Node
//	@return error
func MegaCheckDir(megaDir string) (*mega.Node, error) {
	m := GetSingleton().Mega
	root := m.FS.GetRoot()
	paths := extractPaths(megaDir)
	// check if /path/by/path/ exist
	for _, path := range paths {
		// get nodes by nth path element
		nodes, err := m.FS.PathLookup(root, []string{path})
		// check for error
		if err != nil {
			if err == mega.ENOENT {
				// create if doesn't exist
				node, err := m.CreateDir(path, root)
				if err != nil {
					// error
					return nil, err
				} else {
					// ok, set this node as root
					root = node
				}
			} else {
				// error
				return nil, err
			}
		} else {
			// ok, set this node as root
			root = nodes[len(nodes)-1]
		}
	}
	return root, nil
}

// Uploads file to Mega.
//
//	@param localFilePath Path to local file.
//	@param node Node to upload to.
//	@param fileName Name of uploaded file.
//	@return error
func MegaUpload(localFilePath string, node *mega.Node, fileName string) error {
	_, err := GetSingleton().Mega.UploadFile(localFilePath, node, fileName, nil)
	return err
}
