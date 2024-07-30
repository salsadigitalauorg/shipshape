package config

import (
	"fmt"
	"os"

	"github.com/salsadigitalauorg/shipshape/pkg/utils"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Flags & arg
var ProjectDir string
var Files []string
var CheckTypesToRun []string
var ExcludeDb bool

// Config
var ChecksRegistry = map[CheckType]func() Check{}

func ConfigFilesExist() bool {
	inexistentFilesFound := false
	for _, f := range Files {
		if utils.StringIsUrl(f) {
			continue
		}
		if _, err := os.Stat(f); os.IsNotExist(err) {
			log.WithField("file", f).Error("config file not found")
			inexistentFilesFound = true
		}
	}
	return !inexistentFilesFound
}

func ReadAndParseConfig() (bool, Config, ConfigV2, error) {
	configData, err := FetchConfigData(Files)
	if err != nil {
		return false, Config{}, ConfigV2{}, err
	}

	return ParseConfigData(configData)
}

func FetchConfigData(files []string) ([][]byte, error) {
	var err error
	configData := [][]byte{}
	for _, f := range files {
		var data []byte
		log.WithField("source", f).Info("fetching config")
		if utils.StringIsUrl(f) {
			data, err = utils.FetchContentFromUrl(f)
			if err != nil {
				log.WithField("url", f).WithError(
					err).Error("could not fetch config from url")
				return nil, err
			}
		} else {
			data, err = os.ReadFile(f)
			if err != nil {
				log.WithField("file", f).WithError(
					err).Error("could not fetch config from file")
				return nil, err
			}
		}
		configData = append(configData, data)
	}
	return configData, nil
}

func ParseConfigData(configData [][]byte) (bool, Config, ConfigV2, error) {
	cfgV2 := ConfigV2{}
	data := configData[0]
	if err := yaml.Unmarshal(data, &cfgV2); err != nil {
		log.WithError(err).Debug("config not v2-compatible")
	}

	if len(cfgV2.Collect) > 0 {
		log.WithField("fact plugins", len(cfgV2.Collect)).
			WithField("analyse plugins", len(cfgV2.Analyse)).
			Debug("v2-config parsed")
		return true, Config{}, cfgV2, nil
	}

	finalCfg := Config{}
	for i, data := range configData {
		log.Print("parsing config")
		cfg := Config{}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.WithError(err).Error("could not parse config")
			return false, Config{}, ConfigV2{}, err
		}

		if i == 0 {
			finalCfg = cfg
			if len(configData) == 1 {
				return false, finalCfg, ConfigV2{}, nil
			}
			continue
		}

		log.Print("merging into final config")
		if err := finalCfg.Merge(cfg); err != nil {
			log.WithError(err).Error("could not merge config")
			panic(err)
		}
	}
	return false, finalCfg, ConfigV2{}, nil
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
			return fmt.Errorf(
				"list required under check type '%s', got %s instead",
				ct, check_values[0].ShortTag())
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

// Merge allows multiple checks configurations to be consolidated.
func (cfg *Config) Merge(mrgCfg Config) error {
	if mrgCfg.Checks == nil {
		return nil
	} else if cfg.Checks == nil && mrgCfg.Checks != nil {
		cfg.Checks = mrgCfg.Checks
		return nil
	}

	findCheck := func(checkMap CheckMap, cType CheckType, lookupCheck Check) Check {
		checks, ok := checkMap[cType]
		if !ok {
			return nil
		}
		for _, check := range checks {
			if lookupCheck.GetName() == check.GetName() {
				return check
			}
		}
		return nil
	}

	for cType, checks := range mrgCfg.Checks {
		checksOfSameType := cfg.Checks[cType]
		for _, mrgCheck := range checks {
			if mrgCheck.GetName() == "" {
				for _, existingCheck := range checksOfSameType {
					if err := existingCheck.Merge(mrgCheck); err != nil {
						panic(err)
					}
				}
				continue
			}
			existingCheck := findCheck(cfg.Checks, cType, mrgCheck)
			if existingCheck == nil {
				cfg.Checks[cType] = append(cfg.Checks[cType], mrgCheck)
			} else {
				if err := existingCheck.Merge(mrgCheck); err != nil {
					panic(err)
				}
			}
		}
	}
	return nil
}

// FilterChecksToRun iterates over all the checks and filters them based on
// a provided list of check types to run or whether to exclude database checks.
func (cfg *Config) FilterChecksToRun() {
	newCm := CheckMap{}
	for ct, checks := range cfg.Checks {
		newChecks := []Check{}
		for _, c := range checks {
			if len(CheckTypesToRun) > 0 && !utils.StringSliceContains(CheckTypesToRun, string(ct)) {
				continue
			}
			if ExcludeDb && c.RequiresDatabase() {
				continue
			}
			newChecks = append(newChecks, c)
		}
		if len(newChecks) > 0 {
			newCm[ct] = newChecks
		}
	}
	cfg.Checks = newCm
}
