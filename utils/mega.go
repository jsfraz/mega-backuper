package utils

import (
	"errors"
	"jsfraz/mega-backuper/models"
	"log"
	"regexp"
	"sort"
	"strings"

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

// Get /each/path/element/ from path string.
//
//	@param path
//	@return []string
func extractPaths(path string) []string {
	substrings := strings.Split(path, "/")
	// remove empty strings from the resulting slice
	var result []string
	for _, s := range substrings {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
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

// Uploads file to Mega and keeps last n versions. Others are deleted.
//
//	@param backup
//	@param node Node to upload/delete files to/from.
//	@return error
func MegaDeleteFilesByLastCopyCount(backup models.Backup, node *mega.Node) error {
	// get node children
	m := GetSingleton().Mega
	children, err := m.FS.GetChildren(node)
	if err != nil {
		return err
	}
	// filter all files matching name
	var fileNodes []*mega.Node
	// NAME_RFC3339.tar.gz
	pattern := `^(` + backup.Name + `_)((?:(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}(?:\.\d+)?))(Z|[\+-]\d{2}:\d{2})?)(.tar.gz)$`
	for _, child := range children {
		match, _ := regexp.MatchString(pattern, child.GetName())
		// check if child is file (type == 0) and if name matches pattern
		if child.GetType() == 0 && match {
			fileNodes = append(fileNodes, child)
		}
	}
	var deleteErrs []error
	// delete oldest file(s)
	if len(fileNodes) > backup.LastCopies {
		// sort by newest
		sort.Slice(fileNodes, func(i, j int) bool {
			return fileNodes[i].GetTimeStamp().After(fileNodes[j].GetTimeStamp())
		})
		// delete
		for _, file := range fileNodes[backup.LastCopies:] {
			deleteErr := m.Delete(file, backup.DestroyOldCopies)
			if deleteErr != nil {
				deleteErrs = append(deleteErrs, deleteErr)
			}
		}
	}
	// return error
	if len(deleteErrs) != 0 {
		return errors.Join(deleteErrs...)
	} else {
		return nil
	}
}
