package breach

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type Remediator interface {
	PluginName() string
	Remediate() RemediationResult
	GetRemediationMessage() string
}

type RemediationStatus string

const (
	RemediationStatusNoSupport RemediationStatus = "no-support"
	RemediationStatusSuccess   RemediationStatus = "success"
	RemediationStatusFailed    RemediationStatus = "failed"
	RemediationStatusPartial   RemediationStatus = "partial"
)

type RemediationResult struct {
	Status   RemediationStatus `json:",omitempty"`
	Messages []string          `json:",omitempty"`
}

func RemediatorFromInterface(remediation interface{}) Remediator {
	if remediation == nil {
		return nil
	}

	// Determine the plugin name, if set. Default will be command.
	// Marshal into json first, so we can later unmarshal as required.
	type firstPass struct {
		Plugin string `json:"plugin"`
	}

	jsm, err := json.Marshal(remediation)
	if err != nil {
		log.Fatal(err)
	}
	log.WithField("remediation-json", string(jsm)).Debug()

	var p firstPass
	if err := json.Unmarshal(jsm, &firstPass{}); err != nil {
		log.Fatal(err)
	}

	pluginName := ""
	if p.Plugin == "" {
		pluginName = "command"
	}

	var finalR Remediator
	switch pluginName {
	case "command":
		var r CommandRemediator
		if err := json.Unmarshal(jsm, &r); err != nil {
			log.Fatal(err)
		}
		finalR = &r
	default:
		log.WithField("plugin", pluginName).Fatal("unknown remediation plugin")
	}

	return finalR
}
