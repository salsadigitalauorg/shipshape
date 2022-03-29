package internal

import (
	"net/http"

	"github.com/minio/selfupdate"
)

func SelfUpdate(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	return err
}
