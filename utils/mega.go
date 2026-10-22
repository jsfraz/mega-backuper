package utils

import (
	"fmt"
	"jsfraz/mega-backuper/models"
	"log"
	"sort"
	"strings"

	"github.com/t3rm1n4l/go-mega"
)

// Login to Mega or exit program on fail.
func MegaLogin() {
	s := GetSingleton()
	log.Println("Logging in...")

	s.MegaMu.Lock()
	err := s.Mega.Login(s.Settings.Email, s.Settings.Password)
	if err != nil {
		s.MegaMu.Unlock()
		log.Fatalln(err)
	}
	// get logged user
	user, err := s.Mega.GetUser()
	s.MegaMu.Unlock()
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Logged as %s", user.Name)
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
	s := GetSingleton()

	// hold the lock for the whole lookup/create sequence so parallel jobs
	// cannot create the same directory twice
	s.MegaMu.Lock()
	defer s.MegaMu.Unlock()

	root := s.Mega.FS.GetRoot()
	paths := extractPaths(megaDir)
	// check if /path/by/path/ exist
	for _, path := range paths {
		// get nodes by nth path element
		nodes, err := s.Mega.FS.PathLookup(root, []string{path})
		// check for error
		if err != nil {
			if err == mega.ENOENT {
				// create if doesn't exist
				node, err := s.Mega.CreateDir(path, root)
				if err != nil {
					// error
					return nil, err
				}
				// ok, set this node as root
				root = node
			} else {
				// error
				return nil, err
			}
		} else {
			if len(nodes) == 0 {
				return nil, fmt.Errorf("Mega path component '%s' of '%s' could not be resolved", path, megaDir)
			}
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
	s := GetSingleton()
	s.MegaMu.Lock()
	defer s.MegaMu.Unlock()
	_, err := s.Mega.UploadFile(localFilePath, node, fileName, nil)
	return err
}

// backupFilePrefix returns the file name prefix and extension this job produces.
func backupFilePrefix(backup models.Backup) (string, string) {
	if backup.Type == models.Postgres {
		return backup.PgDb, ".backup"
	}
	return backup.Name, ".tar.gz"
}

// isBackupFileFor reports whether fileName is an archive produced by this job.
//
//	@param backup
//	@param fileName
//	@return bool
func isBackupFileFor(backup models.Backup, fileName string) bool {
	prefix, extension := backupFilePrefix(backup)

	rest, found := strings.CutSuffix(fileName, extension)
	if !found {
		return false
	}
	rest, found = strings.CutPrefix(rest, prefix+"-")
	if !found || rest == "" {
		return false
	}
	for _, r := range rest {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Keeps last n versions. Others are moved to the rubbish bin. Only regular
// files produced by this job are considered, directories and files belonging
// to other jobs are never touched.
//
//	@param backup
//	@param node Node to upload/delete files to/from.
//	@return error
func MegaDeleteFilesByLastCopyCount(backup models.Backup, node *mega.Node) error {
	if backup.LastCopies == nil {
		return nil
	}
	if node == nil {
		return fmt.Errorf("no Mega directory to prune for job '%s'", backup.Name)
	}

	s := GetSingleton()

	s.MegaMu.Lock()
	children, err := s.Mega.FS.GetChildren(node)
	s.MegaMu.Unlock()
	if err != nil {
		return err
	}

	// select only this job's archives
	candidates := make([]*mega.Node, 0, len(children))
	for _, child := range children {
		if child.GetType() != mega.FILE {
			continue
		}
		if !isBackupFileFor(backup, child.GetName()) {
			continue
		}
		candidates = append(candidates, child)
	}

	if len(candidates) <= *backup.LastCopies {
		return nil
	}

	// sort by newest
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].GetTimeStamp().After(candidates[j].GetTimeStamp())
	})

	// move the oldest copies to the rubbish bin
	for _, file := range candidates[*backup.LastCopies:] {
		// FIXME https://github.com/t3rm1n4l/go-mega/pull/46
		// s.Mega.Delete(file, backup.DestroyOldCopies)
		s.MegaMu.Lock()
		err := s.Mega.Delete(file, false)
		s.MegaMu.Unlock()
		if err != nil {
			log.Printf("Failed to move old copy '%s' of job '%s' to the rubbish bin: %v", file.GetName(), backup.Name, err)
		}
	}
	return nil
}
