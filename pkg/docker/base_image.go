package docker

import (
	"bufio"
	"os"
	"regexp"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
	"gopkg.in/yaml.v3"
)

const BaseImage shipshape.CheckType = "docker:base_image"

type BaseImageCheck struct {
	shipshape.CheckBase `yaml:",inline"`
	Allowed             []string `yaml:"allowed"`
	Exclude             []string `yaml:"exclude"`
	Deprecated          []string `yaml:"deprecated"`
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

		for name, def := range compose.Services {
			if utils.StringSliceContains(c.Exclude, name) {
				continue
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
					match := from_regex.FindStringSubmatch(scanner.Text())

					if len(match) < 1 {
						continue
					}

					if len(c.Allowed) > 0 && !utils.StringSliceContains(c.Allowed, match[1]) {
						c.AddFail(name + " is using invalid base image " + match[1])
					} else if len(c.Deprecated) > 0 && utils.StringSliceMatch(c.Deprecated, match[1]) {
						c.AddWarning(name + " is using deprecated image " + match[1])
					} else {
						c.AddPass(name + " is using valid base images")
					}
				}
			} else {
				if !utils.StringSliceMatch(c.Allowed, def.Image) {
					c.AddFail(name + " is using invalid base image " + def.Image)
				} else if utils.StringSliceMatch(c.Deprecated, def.Image) {
					c.AddWarning(name + " is using deprecated image " + def.Image)
				} else {
					c.AddPass(name + " is using valid base images")
				}
			}
		}

		if len(c.Result.Failures) == 0 {
			c.Result.Status = shipshape.Pass
		}
	}

}
