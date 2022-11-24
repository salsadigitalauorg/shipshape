package utils

import (
	"bytes"
	"errors"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
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

// LookupYamlPath attempts to query Yaml data using a JSONPath query and returns
// the found Node.
// It uses the implemention by https://github.com/vmware-labs/yaml-jsonpath.
func LookupYamlPath(n *yaml.Node, path string) ([]*yaml.Node, error) {
	p, err := yamlpath.NewPath(path)
	if err != nil {
		return nil, err
	}
	q, err := p.Find(n)
	if err != nil {
		return nil, err
	}
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

// MergeStringSlice appends the values of a string slice to another
// if those values do not already exist.
func MergeStringSlice(strSlcA *[]string, strSlcB []string) {
	if len(strSlcB) == 0 {
		return
	}
	for _, strB := range strSlcB {
		if !StringSliceContains(*strSlcA, strB) {
			*strSlcA = append(*strSlcA, strB)
		}
	}
}

// MergeIntSlice appends the values of an int slice to another
// if those values do not already exist.
func MergeIntSlice(slcA *[]int, slcB []int) {
	if len(slcB) == 0 {
		return
	}
	for _, strB := range slcB {
		if !IntSliceContains(*slcA, strB) {
			*slcA = append(*slcA, strB)
		}
	}
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
