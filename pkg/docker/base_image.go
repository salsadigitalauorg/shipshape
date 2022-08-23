package docker

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"gopkg.in/yaml.v3"
)

const BaseImage shipshape.CheckType = "docker:base_image"

type BaseImageCheck struct {
	shipshape.CheckBase `yaml:",inline"`
	Allowed             []string `yaml:"allowed"`
	Exclude             []string `yaml:"exclude"`
	Pattern             []string `yaml:"pattern"`
	Paths               []string `yaml:"paths"`
}

type Compose struct {
	Version  string                    `yaml:"version"`
	Services map[string]ComposeService `yaml:"services"`
}

type ComposeService struct {
	Image string `yaml:"image"`
	Build struct {
		Dockerfile string `yaml:"dockerfile"`
	} `yaml:"build"`
}

func (c *BaseImageCheck) RunCheck() {
	for _, path := range c.Paths {
		composeFile := path + string(os.PathSeparator) + "docker-compose.yml"
		bytes, err := os.ReadFile(composeFile)
		if err != nil {
			c.AddWarning("Unable to find " + composeFile)
			continue
		}
		compose := Compose{}
		err = yaml.Unmarshal(bytes, &compose)
		if err != nil {
			c.AddWarning("Invalid docker-compose.yml file " + composeFile)
			continue
		}

	SERVICES:
		for name, def := range compose.Services {
			for _, exclude := range c.Exclude {
				if name == exclude {
					continue SERVICES
				}
			}

			if def.Build.Dockerfile != "" {
				df, err := os.Open(path + string(os.PathSeparator) + def.Build.Dockerfile)
				if err != nil {
					c.AddWarning("Unable to find " + def.Build.Dockerfile)
					continue
				}
				defer df.Close()
				scanner := bufio.NewScanner(df)
				for scanner.Scan() {
					from_regex := regexp.MustCompile("^FROM (.*)")
					match := from_regex.FindString(scanner.Text())
					uses_allowed := false
					if match == "" {
						continue
					}
					for _, i := range c.Allowed {
						if strings.Contains(match, i) {
							uses_allowed = true
						}
					}
					if !uses_allowed {
						c.AddFail(name + " is using invalid base image " + match)
					}
				}
			} else {
				uses_allowed := false
				for _, i := range c.Allowed {
					if strings.Contains(def.Image, i) {
						uses_allowed = true
					}
				}
				if !uses_allowed {
					c.AddFail(name + " is using an invalid base image " + def.Image)
				}
			}
		}

		if len(c.Result.Failures) == 0 {
			c.AddPass("Dockerfiles adhere to the policy")
			c.Result.Status = shipshape.Pass
		}
	}

}
