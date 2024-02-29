package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

// FindFiles scans a directory for files matching the provided pattern.
// excludePattern can be used to ignore files using regex, and a list of
// directories can be skipped using skipDir.
func FindFiles(root, pattern string, excludePattern string, skipDir []string) ([]string, error) {
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
		if d.IsDir() {
			return nil
		}
		if excludePattern != "" {
			if excluded, err := regexp.MatchString(excludePattern, fullpath); err != nil {
				return err
			} else if excluded {
				return nil
			}
		}
		if len(skipDir) > 0 && IsFileInDirs(root, fullpath, skipDir) {
			return nil
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

// IsFileInDirs determines whether a file is in a list of directories.
// root is assumed to be a reference directory, file is the file path with
// the root prefixed and dirs is a list of relative paths from the root.
func IsFileInDirs(root string, file string, dirs []string) bool {
	for _, dir := range dirs {
		dirWithRoot := filepath.Join(root, dir)
		// Find the route to the file relative to the skipped directory.
		// If a route is found and it does not contain the '..' prefix,
		// the file is within the skipped dir, and therefore should be
		// skipped.
		rel, err := filepath.Rel(dirWithRoot, file)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(rel, "..") {
			return true
		}
	}
	return false
}

// LookupYamlPath attempts to query Yaml data using a JSONPath query and returns
// the found Node.
// It uses the implemention by https://github.com/vmware-labs/yaml-jsonpath.
func LookupYamlPath(n *yaml.Node, path string) ([]*yaml.Node, error) {
	log.WithField("path", path).Debug("looking up yaml path")
	p, err := yamlpath.NewPath(path)
	if err != nil {
		log.WithError(err).Debug("failed to lookup yaml path")
		return nil, err
	}
	q, _ := p.Find(n)
	return q, nil
}

// MergeBoolPtrs compares two bool pointers and replaces
// boolA with boolB if the latter is non-nil.
func MergeBoolPtrs(boolA *bool, boolB *bool) {
	if boolB != nil && boolB != boolA {
		*boolA = *boolB
	}
}

// MergeString compares two strings and replaces
// strA with strB if the latter is not empty.
func MergeString(strA *string, strB string) {
	if strB != "" && *strA != strB {
		*strA = strB
	}
}

// MergeStringSlice replaces the values of a string slice with those of another.
func MergeStringSlice(slcA *[]string, slcB []string) {
	if len(slcB) == 0 {
		return
	}
	// Create new slice with unique values.
	newSlc := []string{}
	for _, valB := range slcB {
		if !StringSliceContains(newSlc, valB) {
			newSlc = append(newSlc, valB)
		}
	}
	*slcA = newSlc
}

// MergeIntSlice replaces the values of a int slice with those of another.
func MergeIntSlice(slcA *[]int, slcB []int) {
	if len(slcB) == 0 {
		return
	}
	// Create new slice with unique values.
	newSlc := []int{}
	for _, valB := range slcB {
		if !IntSliceContains(newSlc, valB) {
			newSlc = append(newSlc, valB)
		}
	}
	*slcA = newSlc
}

// SliceContains determines whether an item exists in a slice of any type.
func SliceContains(slice any, item any) bool {
	var s []any
	var ok bool
	if s, ok = slice.([]any); !ok {
		return false
	}

	for _, i := range s {
		if i == nil && item == nil {
			return true
		} else if reflect.TypeOf(i) != reflect.TypeOf(item) {
			continue
		} else if fmt.Sprint(i) == fmt.Sprint(item) {
			return true
		}
	}
	return false
}

// StringSliceContains determines whether an item exists in a slice of string.
func StringSliceContains(slice []string, item string) bool {
	if len(slice) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}

// IntSliceContains determines whether an item exists in a slice of int.
func IntSliceContains(slice []int, item int) bool {
	if len(slice) == 0 {
		return false
	}
	set := make(map[int]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}

// StringSlicesIntersect finds the intersection between two slices of string.
func StringSlicesIntersect(slc1 []string, slc2 []string) []string {
	intersect := []string(nil)

	// Create a map of the first slice for easy lookup from a loop over second
	// slice.
	mappedSlc := map[string]string{}
	for _, s1 := range slc1 {
		mappedSlc[s1] = s1
	}

	for _, s2 := range slc2 {
		if _, ok := mappedSlc[s2]; !ok {
			continue
		}
		intersect = append(intersect, s2)
	}

	return intersect
}

// StringSlicesIntersectUnique returns the unique strings in slc2 that also present in slc1.
func StringSlicesIntersectUnique(slc1, slc2 []string) []string {
	// Convert slc1 to a map for faster lookup.
	map1 := make(map[string]struct{})
	for _, x := range slc1 {
		map1[x] = struct{}{}
	}

	intersect := make(map[string]struct{})
	for _, x := range slc2 {
		if _, found := map1[x]; found {
			intersect[x] = struct{}{}
		}
	}

	keys := []string(nil)
	for k := range intersect {
		keys = append(keys, k)
	}

	return keys
}

// StringSlicesInterdiff returns the strings in slc2 that do not present in slc1.
func StringSlicesInterdiff(slc1 []string, slc2 []string) []string {
	// Convert slc1 to a map for faster lookup.
	map1 := make(map[string]struct{}, len(slc1))
	for _, x := range slc1 {
		map1[x] = struct{}{}
	}

	interdiff := []string(nil)
	for _, x := range slc2 {
		if _, found := map1[x]; !found {
			interdiff = append(interdiff, x)
		}
	}

	return interdiff
}

// StringSlicesInterdiffUnique returns the unique strings in slc2 that do not present in slc1.
func StringSlicesInterdiffUnique(slc1, slc2 []string) []string {
	// Convert slc1 to a map for faster lookup.
	map1 := make(map[string]struct{})
	for _, x := range slc1 {
		map1[x] = struct{}{}
	}

	interdiff := make(map[string]struct{})
	for _, x := range slc2 {
		if _, found := map1[x]; !found {
			interdiff[x] = struct{}{}
		}
	}

	keys := []string(nil)
	for k := range interdiff {
		keys = append(keys, k)
	}

	return keys
}

// StringIsUrl determines whether a string is a url by trying to parse it.
func StringIsUrl(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// FetchContentFromUrl fetches the content from a url and returns its bytes.
func FetchContentFromUrl(u string) ([]byte, error) {
	rsp, err := http.Get(u)
	if err != nil {
		return []byte(nil), err
	}

	defer rsp.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(rsp.Body); err != nil {
		return []byte(nil), err
	}
	return buf.Bytes(), nil
}

func StringSliceMatch(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(item, s) {
			return true
		}
	}
	return false
}

// Sift through a slice to determine if it contains eligible package
// with optional version constrains.
func PackageCheckString(slice []string, item string, item_version string) bool {
	for _, s := range slice {
		// Parse slice with regex to:
		// 1 - package name (e.g. "bitnami/kubectl")
		// 2 - version (e.g. "8.0")
		service_regex := regexp.MustCompile("^(.[^:@]*)?[:@]?([^ latest$]*)")
		service_match := service_regex.FindStringSubmatch(s)
		// Only proceed if package names were parsed successfully.
		if len(service_match[1]) > 0 && len(item) > 0 {
			// Check if package name matches.
			if service_match[1] == item {
				// Package name matched.
				// If service does not dictate version than assume any version is allowed.
				if len(service_match[2]) < 1 {
					return true
				} else if len(item_version) > 0 {
					// Ensure that item version is not less than slice version.
					allowedVersion, err := version.NewVersion(service_match[2])
					imageVersion, err := version.NewVersion(item_version)
					// Run version comparison.
					if err == nil && allowedVersion.LessThanOrEqual(imageVersion) {
						return true
					}
				}
			}
		}
	}
	return false
}

func Glob(dir string, match string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if strings.Contains(f.Name(), match) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), err
}

func FileContains(loc string, match string) (bool, error) {
	if _, err := os.Stat(loc); err != nil {
		return false, fmt.Errorf("File not found at %s", loc)
	}
	file, err := os.Open(loc)
	defer file.Close()

	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if scanner.Text() == match {
			return true, nil
		}
	}

	return false, nil
}

func HasComposerDependency(loc string, deps []string) (bool, error) {
	var composer map[string]interface{}
	composerFile := loc + string(os.PathSeparator) + "composer.json"

	if _, err := os.Stat(composerFile); err != nil {
		return false, fmt.Errorf("File not found at %s", loc)
	}

	cf, err := os.Open(composerFile)

	if err != nil {
		return false, err
	}

	defer cf.Close()
	byteValue, _ := io.ReadAll(cf)

	json.Unmarshal([]byte(byteValue), &composer)

	if req, ok := composer["require"]; ok {
		if r, ok := req.(map[string]interface{}); ok {
			for k := range r {
				for _, dep := range deps {
					if k == dep {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}
