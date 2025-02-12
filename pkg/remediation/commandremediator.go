package remediation

import "github.com/salsadigitalauorg/shipshape/pkg/command"

type CommandRemediator struct {
	// Common fields.
	Message string `json:"msg"`

	// Plugin fields.
	Command   string   `json:"cmd"`
	Arguments []string `json:"args"`
}

//go:generate go run ../../cmd/gen.go remediator --plugin=CommandRemediator --name=command

func init() {
	Registry["command"] = func() Remediator { return &CommandRemediator{} }
}

func (p *CommandRemediator) Remediate() RemediationResult {
	_, err := command.ShellCommander(p.Command, p.Arguments...).Output()
	if err != nil {
		return RemediationResult{
			Status:   RemediationStatusFailed,
			Messages: []string{command.GetMsgFromCommandError(err)},
		}
	}

	return RemediationResult{
		Status:   RemediationStatusSuccess,
		Messages: []string{p.GetRemediationMessage()},
	}
}
