package docker

import (
	"bufio"
	"os"
	"regexp"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"

	"gopkg.in/yaml.v3"
)

const BaseImage config.CheckType = "docker:base_image"

type BaseImageCheck struct {
	config.CheckBase `yaml:",inline"`
	Allowed          []string `yaml:"allowed"`
	Exclude          []string `yaml:"exclude"`
	Deprecated       []string `yaml:"deprecated"`
	Pattern          []string `yaml:"pattern"`
	Paths            []string `yaml:"paths"`
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

// Merge implementation for DbModuleCheck check.
func (c *BaseImageCheck) Merge(mergeCheck config.Check) error {
	baseImageMergeCheck := mergeCheck.(*BaseImageCheck)
	if err := c.CheckBase.Merge(&baseImageMergeCheck.CheckBase); err != nil {
		return err
	}

	utils.MergeStringSlice(&c.Allowed, baseImageMergeCheck.Allowed)
	utils.MergeStringSlice(&c.Exclude, baseImageMergeCheck.Exclude)
	utils.MergeStringSlice(&c.Deprecated, baseImageMergeCheck.Deprecated)
	utils.MergeStringSlice(&c.Pattern, baseImageMergeCheck.Pattern)
	utils.MergeStringSlice(&c.Paths, baseImageMergeCheck.Paths)
	return nil
}

func (c *BaseImageCheck) RequiresData() bool { return false }

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
					from_regex := regexp.MustCompile("^FROM (.[^:@]*)?[:@]?([^ latest$]*)")
					match := from_regex.FindStringSubmatch(scanner.Text())

					if len(match) < 1 {
						continue
					}

					if len(c.Allowed) > 0 && !utils.PackageCheckString(c.Allowed, match[1], match[2]) {
						c.AddBreach(&breach.KeyValueBreach{
							KeyLabel:   "service",
							Key:        name,
							ValueLabel: "invalid base image",
							Value:      match[1],
						})
					} else if len(c.Deprecated) > 0 && utils.StringSliceMatch(c.Deprecated, match[1]) {
						c.AddWarning(name + " is using deprecated image " + match[1])
					} else {
						c.AddPass(name + " is using valid base images")
					}
				}
			} else {
				// Extract image package name and optional version from definition.
				image_regex := regexp.MustCompile("^(.[^:@]*)?[:@]?([^ latest$]*)")
				match := image_regex.FindStringSubmatch(def.Image)

				if len(match) < 1 {
					continue
				}

				if !utils.PackageCheckString(c.Allowed, match[1], match[2]) {
					c.AddBreach(&breach.KeyValueBreach{
						KeyLabel:   "service",
						Key:        name,
						ValueLabel: "invalid base image",
						Value:      def.Image,
					})
				} else if utils.StringSliceMatch(c.Deprecated, match[1]) {
					c.AddWarning(name + " is using deprecated image " + match[1])
				} else {
					c.AddPass(name + " is using valid base images")
				}
			}
		}

		if len(c.Result.Breaches) == 0 {
			c.Result.Status = result.Pass
		}
	}

}
