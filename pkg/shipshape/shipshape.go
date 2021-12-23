package shipshape

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

func ParseCheckFile(f string) (Config, error) {
	var err error

	c := Config{}

	data, err := ioutil.ReadFile(f)
	// The file does not yet exist, just exit for now.
	if err != nil {
		log.Fatalf("error reading file: %v", err)
		return c, err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		log.Fatalf("error unmarshaling: %v", err)
		return c, err
	}
	return c, nil
}
