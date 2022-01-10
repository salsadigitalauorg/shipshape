package core

import "gopkg.in/yaml.v3"

type CheckType string

type Check interface {
	Init(pd string, ct CheckType)
	GetName() string
	FetchData() error
	RunCheck() error
	GetResult() Result
}

type CheckBase struct {
	Name       string `yaml:"name"`
	ProjectDir string `yaml:""`
	Data       []byte
	Result     Result
}

type Result struct {
	Name      string      `json:"name"`
	CheckType CheckType   `json:"check-type"`
	Status    CheckStatus `json:"status"`
	Passes    []string    `json:"passes"`
	Failures  []string    `json:"failures"`
	Error     string      `json:"error"`
}

type ResultList struct {
	Results map[string]Result `json:"results"`
	Errors  map[string]error  `json:"errors"`
}

type CheckStatus string

const (
	Pass CheckStatus = "Pass"
	Fail CheckStatus = "Fail"
)

type KeyValue struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type KeyValueResult int8

const (
	KeyValueError    KeyValueResult = -2
	KeyValueNotFound KeyValueResult = -1
	KeyValueNotEqual KeyValueResult = 0
	KeyValueEqual    KeyValueResult = 1
)

type YamlCheck struct {
	Values []KeyValue `yaml:"config-values"`
	Node   yaml.Node
}
