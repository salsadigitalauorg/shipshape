package utils

import (
	"errors"
	"io/fs"
	"path/filepath"
	"regexp"
)

// FindFiles scans a directory for files matching the provided patterns.
func FindFiles(root, pattern string, excludePattern string) ([]string, error) {
	if root == "" {
		return nil, errors.New("directory not provided")
	}
	if pattern == "" {
		return nil, errors.New("pattern not provided")
	}

	var matches []string
	err := filepath.WalkDir(root, func(fullpath string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if excludePattern != "" {
			if excluded, err := regexp.MatchString(excludePattern, d.Name()); err != nil {
				return err
			} else if excluded {
				return nil
			}
		}
		if matched, err := regexp.MatchString(pattern, d.Name()); err != nil {
			return err
		} else if matched {
			matches = append(matches, fullpath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func StringSliceContains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
