package shipshape

import (
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
