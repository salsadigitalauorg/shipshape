package env

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/drone/envsubst"
	"github.com/joho/godotenv"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

type EnvResolver interface {
	ShouldResolveEnv() bool
	GetEnvFile() string
	GetEnvMap() (map[string]string, error)
}

func ReadEnvFile(er EnvResolver) (map[string]string, error) {
	if !er.ShouldResolveEnv() {
		return nil, nil
	}

	envFile := er.GetEnvFile()
	if envFile == "" {
		envFile = filepath.Join(config.ProjectDir, ".env")
	}

	if _, err := os.Stat(envFile); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	envMap, err := godotenv.Read(envFile)
	if err != nil {
		return nil, err
	}

	return envMap, nil
}

func ResolveValue(envMap map[string]string, rawVal string) (string, error) {
	if envMap == nil {
		return rawVal, nil
	}

	evaled, err := envsubst.Eval(rawVal, func(s string) string {
		if val, ok := envMap[s]; ok {
			return val
		}
		return ""
	})
	if err != nil {
		return "", nil
	}

	// Deal with the case `$varname` which is not supported by envsubst.
	if len(evaled) > 1 && evaled[0] == '$' && evaled[1] != '{' {
		if val, ok := envMap[evaled[1:]]; ok {
			return val, nil
		}
	}

	return evaled, err
}
