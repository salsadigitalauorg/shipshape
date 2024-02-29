// Package shipshape provides the basic types and base functions for parsing
// config, running checks as well as the file & yaml checks.
package shipshape

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/salsadigitalauorg/shipshape/pkg/analyse"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/lagoon"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var IsV2 bool
var RunConfig config.Config
var RunConfigV2 config.ConfigV2
var RunResultList result.ResultList
var OutputFormats = []string{"json", "junit", "simple", "table"}

func Init(projectDir string, configFiles []string, checkTypesToRun []string, excludeDb bool, remediate bool, logLevel string, lagoonApiBaseUrl string, lagoonApiToken string) error {
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

	if IsV2 {
		return nil
	}

	config.ProjectDir = RunConfig.ProjectDir
	RunResultList = result.NewResultList(remediate)

	// Remediate is a command-level flag, so we set the value outside of
	// config parsing.
	RunConfig.Remediate = remediate

	// Base url can either be provided in the config file or in env var, the
	// latter being final.
	if lagoonApiBaseUrl != "" {
		lagoon.ApiBaseUrl = lagoonApiBaseUrl
	} else {
		lagoon.ApiBaseUrl = RunConfig.LagoonApiBaseUrl
	}
	lagoon.ApiToken = lagoonApiToken

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

	if IsV2 {
		return nil
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
	cfgV2 := config.ConfigV2{}
	data := configData[0]
	if err := yaml.Unmarshal(data, &cfgV2); err != nil {
		log.WithError(err).Debug("config not v2-compatible")
		// return err
	}
	RunConfigV2 = cfgV2
	if len(RunConfigV2.Gather) > 0 {
		IsV2 = true
		return nil
	}

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

func RunChecks() {
	log.Print("preparing concurrent check runs")
	var wg sync.WaitGroup
	for ct, checks := range RunConfig.Checks {
		checks := checks
		RunResultList.IncrChecks(string(ct), len(checks))
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
	RunResultList.RemediationTotalsCount()
}

func ProcessCheck(rl *result.ResultList, c config.Check) {
	contextLogger := log.WithFields(log.Fields{
		"check-type": c.GetType(),
		"check-name": c.GetName(),
	})
	contextLogger.Print("processing check")
	if c.RequiresData() {
		contextLogger.Print("fetching data")
		c.FetchData()
		c.HasData(true)
		if len(c.GetResult().Breaches) == 0 {
			c.UnmarshalDataMap()
		}
	}
	if len(c.GetResult().Breaches) == 0 && len(c.GetResult().Passes) == 0 {
		contextLogger.Print("running check")
		c.RunCheck()
	}
	if len(c.GetResult().Breaches) > 0 && c.ShouldPerformRemediation() {
		contextLogger.Print("performing remediation")
		c.Remediate()
	}
	c.GetResult().DetermineResultStatus(c.ShouldPerformRemediation())
	contextLogger.
		WithFields(log.Fields{"result": c.GetResult()}).
		Print("check processed")
	rl.AddResult(*c.GetResult())
}

func RunV2() {
	log.WithField("config", fmt.Sprintf("%+v", RunConfigV2)).Trace("running v2")

	log.Print("parsing connections config")
	connection.ParseConfig(RunConfigV2.Connections)
	log.Print("parsing facts config")
	fact.ParseConfig(RunConfigV2.Gather)
	log.Print("parsing analysers config")
	analyse.ParseConfig(RunConfigV2.Analyse)

	log.Print("validating connection connections - TODO")
	log.Print("validating fact connections - TODO")
	log.Print("validating fact inputs - TODO")

	log.Print("validating analyser inputs")
	analyse.ValidateInputs()
	if len(analyse.Errors) > 0 {
		log.WithField("errors", analyse.Errors).Fatal("failed to validate analyser inputs")
	}

	log.Print("gathering facts")
	fact.GatherAllFacts()

	log.Print("analysing facts")
	analyse.AnalyseAll()
}
