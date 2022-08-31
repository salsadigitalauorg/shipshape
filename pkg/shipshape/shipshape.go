// Package shipshape provides the basic types and base functions for parsing
// config, running checks as well as the file & yaml checks.
package shipshape

import (
	"errors"
	"os"

	"github.com/salsadigitalauorg/shipshape/pkg/utils"

	"gopkg.in/yaml.v3"
)

func ReadAndParseConfig(projectDir string, files []string) (Config, error) {
	finalCfg := Config{}
	for i, f := range files {
		var data []byte
		var err error
		cfg := Config{}
		if utils.StringIsUrl(f) {
			data, err = utils.FetchContentFromUrl(f)
			if err != nil {
				return cfg, err
			}
		} else {
			data, err = os.ReadFile(f)
			if err != nil {
				return cfg, err
			}
		}

		if err := ParseConfig(data, projectDir, &cfg); err != nil {
			return cfg, err
		}

		if i == 0 {
			finalCfg = cfg
			if len(files) == 1 {
				return finalCfg, nil
			}
			continue
		}

		if err := finalCfg.Merge(cfg); err != nil {
			panic(err)
		}
	}

	return finalCfg, nil
}

func ParseConfig(data []byte, projectDir string, cfg *Config) error {
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}

	if cfg.ProjectDir == "" && projectDir != "" {
		cfg.ProjectDir = projectDir
	} else {
		// Default project directory is current directory.
		projectDir, _ = os.Getwd()
		cfg.ProjectDir = projectDir
	}

	if cfg.FailSeverity == "" {
		cfg.FailSeverity = HighSeverity
	}

	return nil
}

func (cm *CheckMap) UnmarshalYAML(value *yaml.Node) error {
	newcm := make(CheckMap)
	for ct, cFunc := range ChecksRegistry {
		check_values, err := utils.LookupYamlPath(value, string(ct))
		if err != nil {
			return err
		}

		if len(check_values) == 0 {
			continue
		}

		if check_values[0].Kind != yaml.SequenceNode {
			return errors.New("yaml: unmarshal errors")
		}

		for _, cv := range check_values[0].Content {
			c := cFunc()
			err := cv.Decode(c)
			if err != nil {
				return err
			}
			newcm[ct] = append(newcm[ct], c)
		}
	}
	*cm = newcm
	return nil
}
