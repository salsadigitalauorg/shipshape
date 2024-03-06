// Package shipshape provides the basic types and base functions for parsing
// config, running checks as well as the file & yaml checks.
package shipshape

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/analyse"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

// Flags & arg
var ErrorCodeOnFailure bool
var Remediate bool
var FailSeverity string

// Config
var IsV2 bool
var RunConfig config.Config
var RunConfigV2 config.ConfigV2

// Results
var RunResultList result.ResultList

func Init() error {
	log.Print("initialising shipshape")
	config.Files = []string{"testdata/shipshape.yml"}
	isV2, cfg, cfgV2, err := config.ReadAndParseConfig()
	if err != nil {
		return err
	}
	IsV2 = isV2

	if IsV2 {
		RunConfigV2 = cfgV2
		return nil
	}

	RunConfig = cfg

	RunResultList = result.NewResultList(Remediate)

	log.WithFields(log.Fields{
		"ProjectDir":    config.ProjectDir,
		"FailSeverity":  FailSeverity,
		"Remediate":     Remediate,
		"RunResultList": fmt.Sprintf("%+v", RunResultList),
	}).Debug("basic config")

	log.Print("initialising checks")
	var checksCount int
	for ct, checks := range RunConfig.Checks {
		for _, c := range checks {
			c.Init(ct)
			c.SetPerformRemediation(Remediate)
			checksCount++
		}
	}

	log.Print("filtering checks")
	RunConfig.FilterChecksToRun()
	log.WithField("checksCount", checksCount).Print("checks filtered")
	jsonChecks, _ := json.Marshal(RunConfig.Checks)
	log.WithFields(log.Fields{
		"Checks": string(jsonChecks),
	}).Debug("checks initialised and filtered")

	return nil
}

func Run() {
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
	fact.ParseConfig(RunConfigV2.Collect)
	log.Print("parsing analysers config")
	analyse.ParseConfig(RunConfigV2.Analyse)

	log.Print("validating connection connections - TODO")
	log.Print("validating fact connections - TODO")
	log.Print("validating fact inputs - TODO")

	log.Print("validating analyser inputs")
	analyse.ValidateInputs()
	if len(analyse.Errors) > 0 {
		log.WithField("errors", analyse.Errors).
			Fatal("failed to validate analyser inputs")
	}

	log.Print("collecting facts")
	fact.CollectAllFacts()

	log.Print("analysing facts")
	results := analyse.AnalyseAll()

	RunResultList = result.NewResultList(false)
	for _, r := range results {
		r.DetermineResultStatus(false)
		RunResultList.AddResult(r)
	}
}

func Exit(code int) {
	if code > 0 && ErrorCodeOnFailure {
		os.Exit(1)
	}
	os.Exit(0)
}
