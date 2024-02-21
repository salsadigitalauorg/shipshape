package utils

import "strings"

func MultilineOutputToSlice(output []byte) []string {
	slc := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		slc = append(slc, string(line))
	}
	return slc
}
