package shipshape

import (
	"sync"

	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

// Merge allows multiple checks configurations to be consolidated.
func (cfg *Config) Merge(mrgCfg Config) error {
	if mrgCfg.ProjectDir != "" {
		cfg.ProjectDir = mrgCfg.ProjectDir
	}
	if mrgCfg.FailSeverity != "" {
		cfg.FailSeverity = mrgCfg.FailSeverity
	}

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
		for _, mrgCheck := range checks {
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

func (cfg *Config) Init() {
	ProjectDir = cfg.ProjectDir
	for ct, checks := range cfg.Checks {
		for _, c := range checks {
			c.Init(ct)
		}
	}
}

// FilterChecksToRun iterates over all the checks and filters them based on
// a provided list of check types to run or whether to exclude database checks.
func (cfg *Config) FilterChecksToRun(checkTypesToRun []string, excludeDb bool) {
	newCm := CheckMap{}
	for ct, checks := range cfg.Checks {
		newChecks := []Check{}
		for _, c := range checks {
			if len(checkTypesToRun) > 0 && !utils.StringSliceContains(checkTypesToRun, string(ct)) {
				continue
			}
			if excludeDb && c.RequiresDatabase() {
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

func (cfg *Config) RunChecks() ResultList {
	rl := ResultList{
		config:  cfg,
		Results: []Result{},
	}
	var wg sync.WaitGroup
	for ct, checks := range cfg.Checks {
		checks := checks
		rl.IncrChecks(ct, len(checks))
		for i := range checks {
			wg.Add(1)
			check := checks[i]
			go func() {
				defer wg.Done()
				cfg.ProcessCheck(&rl, check)
			}()
		}
	}
	wg.Wait()
	rl.Sort()
	return rl
}

func (cfg *Config) ProcessCheck(rl *ResultList, c Check) {
	if c.RequiresData() {
		c.FetchData()
		c.HasData(true)
		if len(c.GetResult().Failures) == 0 {
			c.UnmarshalDataMap()
		}
	}
	if len(c.GetResult().Failures) == 0 && len(c.GetResult().Passes) == 0 {
		c.RunCheck()
		c.GetResult().Sort()
	}
	rl.Results = append(rl.Results, *c.GetResult())
	rl.IncrBreaches(c, len(c.GetResult().Failures))
}
