package licensedb

import (
	"errors"
	paths "path"

	"gopkg.in/src-d/go-license-detector.v2/licensedb/filer"
	"gopkg.in/src-d/go-license-detector.v2/licensedb/internal"
)

var (
	// ErrNoLicenseFound is raised if no license files were found.
	ErrNoLicenseFound = errors.New("no license file was found")
)

// Detect returns the most probable reference licenses matched for the given
// file tree. Each match has the confidence assigned, from 0 to 1, 1 means 100% confident.
func Detect(fs filer.Filer) (map[string]float32, error) {
	files, err := fs.ReadDir("")
	if err != nil {
		return nil, err
	}
	fileNames := []string{}
	for _, file := range files {
		if !file.IsDir {
			fileNames = append(fileNames, file.Name)
		} else if internal.IsLicenseDirectory(file.Name) {
			// "license" directory, let's look inside
			subfiles, err := fs.ReadDir(file.Name)
			if err == nil {
				for _, subfile := range subfiles {
					if !subfile.IsDir {
						fileNames = append(fileNames, paths.Join(file.Name, subfile.Name))
					}
				}
			}
		}
	}
	candidates := internal.ExtractLicenseFiles(fileNames, fs)
	licenses := internal.InvestigateLicenseTexts(candidates)
	if len(licenses) > 0 {
		return licenses, nil
	}
	// Plan B: take the README, find the section about the license and apply NER
	candidates = internal.ExtractReadmeFiles(fileNames, fs)
	if len(candidates) > 0 {
		licenses = internal.InvestigateReadmeTexts(candidates, fs)
		if len(licenses) > 0 {
			return licenses, nil
		}
	}
	// Plan C: look for licence texts in source code files with comments at header
	candidates = internal.ExtractSourceFiles(fileNames, fs)
	if len(candidates) > 0 {
		licenses = internal.InvestigateHeaderComments(candidates, fs)
	}
	if len(licenses) == 0 {
		return nil, ErrNoLicenseFound
	}
	return licenses, nil
}
