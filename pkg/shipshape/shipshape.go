// Package shipshape provides the basic types and base functions for parsing
// config, running checks as well as the file & yaml checks.
package shipshape

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

//go:generate go run ../../cmd/gen.go registry --checkpackage=shipshape

var ProjectDir string
var RunConfig config.Config
var RunResultList ResultList

func init() {
	config.ChecksRegistry[File] = func() config.Check { return &FileCheck{} }
	config.ChecksRegistry[Yaml] = func() config.Check { return &YamlCheck{} }
	config.ChecksRegistry[YamlLint] = func() config.Check { return &YamlLintCheck{} }
	config.ChecksRegistry[Crawler] = func() config.Check { return &CrawlerCheck{} }
}

var OutputFormats = []string{"json", "junit", "simple", "table"}

func Init(projectDir string, configFiles []string, checkTypesToRun []string, excludeDb bool, remediate bool, logLevel string) error {
	if logLevel == "" {
		logLevel = "warn"
	}
	if logrusLevel, err := log.ParseLevel(logLevel); err != nil {
		panic(err)
	} else {
		log.SetLevel(logrusLevel)
	}

	log.Print("initialising shipshape")
	err := ReadAndParseConfig(projectDir, configFiles)
	if err != nil {
		return err
	}

	ProjectDir = RunConfig.ProjectDir
	RunResultList = NewResultList(remediate)

	// Remediate is a command-level flag, so we set the value outside of
	// config parsing.
	RunConfig.Remediate = remediate

	log.WithFields(log.Fields{
		"ProjectDir":    RunConfig.ProjectDir,
		"FailSeverity":  RunConfig.FailSeverity,
		"Remediate":     RunConfig.Remediate,
		"RunResultList": fmt.Sprintf("%+v", RunResultList),
	}).Debug("basic config")

	log.Print("initialising checks")
	var checksCount int
	for ct, checks := range RunConfig.Checks {
		for _, c := range checks {
			c.Init(ct)
			c.SetPerformRemediation(remediate)
			checksCount++
		}
	}

	log.Print("filtering checks")
	RunConfig.FilterChecksToRun(checkTypesToRun, excludeDb)
	log.WithField("checksCount", checksCount).Print("checks filtered")
	jsonChecks, _ := json.Marshal(RunConfig.Checks)
	log.WithFields(log.Fields{
		"Checks": string(jsonChecks),
	}).Debug("checks initialised and filtered")

	return nil
}

func ReadAndParseConfig(projectDir string, files []string) error {
	configData, err := FetchConfigData(files)
	if err != nil {
		return err
	}
	err = ParseConfigData(configData)
	if err != nil {
		return err
	}

	if RunConfig.ProjectDir == "" && projectDir != "" {
		RunConfig.ProjectDir = projectDir
	} else {
		// Default project directory is current directory.
		projectDir, _ = os.Getwd()
		RunConfig.ProjectDir = projectDir
	}

	if RunConfig.FailSeverity == "" {
		RunConfig.FailSeverity = config.HighSeverity
	}

	return nil
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

func ParseConfigData(configData [][]byte) error {
	finalCfg := config.Config{}
	for i, data := range configData {
		log.Print("parsing config")
		cfg := config.Config{}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.WithError(err).Error("could not parse config")
			return err
		}

		if i == 0 {
			finalCfg = cfg
			if len(configData) == 1 {
				RunConfig = finalCfg
				return nil
			}
			continue
		}

		log.Print("merging into final config")
		if err := finalCfg.Merge(cfg); err != nil {
			log.WithError(err).Error("could not merge config")
			panic(err)
		}
	}
	RunConfig = finalCfg
	return nil
}

func RunChecks() ResultList {
	log.Print("preparing concurrent check runs")
	var wg sync.WaitGroup
	for ct, checks := range RunConfig.Checks {
		checks := checks
		RunResultList.IncrChecks(ct, len(checks))
		for i := range checks {
			wg.Add(1)
			check := checks[i]
			go func() {
				defer wg.Done()
				ProcessCheck(&RunResultList, check)
			}()
		}
	}
	wg.Wait()
	RunResultList.Sort()
	return RunResultList
}

func ProcessCheck(rl *ResultList, c config.Check) {
	contextLogger := log.WithFields(log.Fields{
		"check-type": c.GetType(),
		"check-name": c.GetName(),
	})
	contextLogger.Print("processing check")
	if c.RequiresData() {
		contextLogger.Print("fetching data")
		c.FetchData()
		c.HasData(true)
		if len(c.GetResult().Failures) == 0 {
			c.UnmarshalDataMap()
		}
	}
	if len(c.GetResult().Failures) == 0 && len(c.GetResult().Passes) == 0 {
		contextLogger.Print("running check")
		c.RunCheck()
		c.GetResult().Sort()
	}
	rl.AddResult(*c.GetResult())
}
