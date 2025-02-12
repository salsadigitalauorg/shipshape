package remediation

import (
	"encoding/json"
	"fmt"

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

var Registry = map[string]func() Remediator{}

func RemediatorFromInterface(remediation interface{}) Remediator {
	if remediation == nil {
		return nil
	}

	// Marshal into json first, so we can later unmarshal as required.
	jsm, err := json.Marshal(remediation)
	if err != nil {
		log.Fatal(err)
	}
	log.WithField("remediation-json", string(jsm)).Trace()

	// Determine the plugin name, if set. Default will be command.
	var firstPass struct {
		Plugin string `json:"plugin"`
	}
	if err := json.Unmarshal(jsm, &firstPass); err != nil {
		log.Fatal(err)
	}
	log.WithField("firstPass", firstPass.Plugin).Debug("parsing remediator")

	pluginName := firstPass.Plugin
	if pluginName == "" {
		pluginName = "command"
	}

	var finalR Remediator
	regEntry, ok := Registry[pluginName]
	if !ok {
		log.WithField("plugin", pluginName).Fatal("unknown remediation plugin")
	}

	finalR = regEntry()
	if err := json.Unmarshal(jsm, &finalR); err != nil {
		log.Fatal(err)
	}

	log.WithField("remediator", fmt.Sprintf("%#v", finalR)).
		Debug("parsed remediator")
	return finalR
}
