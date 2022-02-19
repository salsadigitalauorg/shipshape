package core

import "gopkg.in/yaml.v3"

type CheckType string

type Check interface {
	Init(pd string, ct CheckType)
	GetName() string
	FetchData()
	RunCheck()
	GetResult() Result
}

type CheckBase struct {
	Name       string `yaml:"name"`
	ProjectDir string `yaml:""`
	DataMap    map[string][]byte
	Result     Result
}

type Result struct {
	Name      string      `json:"name"`
	CheckType CheckType   `json:"check-type"`
	Status    CheckStatus `json:"status"`
	Passes    []string    `json:"passes"`
	Failures  []string    `json:"failures"`
}

type ResultList struct {
	Results []Result `json:"results"`
}

type CheckStatus string

const (
	Pass CheckStatus = "Pass"
	Fail CheckStatus = "Fail"
)

type KeyValue struct {
	Key        string   `yaml:"key"`
	Value      string   `yaml:"value"`
	IsList     bool     `yaml:"is-list"`
	Disallowed []string `yaml:"disallowed"`
}

type KeyValueResult int8

const (
	KeyValueError           KeyValueResult = -2
	KeyValueNotFound        KeyValueResult = -1
	KeyValueNotEqual        KeyValueResult = 0
	KeyValueEqual           KeyValueResult = 1
	KeyValueDisallowedFound KeyValueResult = 2
)

type FileCheck struct {
	CheckBase         `yaml:",inline"`
	Path              string `yaml:"path"`
	DisallowedPattern string `yaml:"disallowed-pattern"`
}

const (
	File CheckType = "file"
)

type YamlCheck struct {
	Values  []KeyValue `yaml:"values"`
	Node    yaml.Node
	NodeMap map[string]yaml.Node
}
