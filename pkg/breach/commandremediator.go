package breach

import "github.com/salsadigitalauorg/shipshape/pkg/command"

type CommandRemediator struct {
	// Common fields.
	Message string `yaml:"msg"`

	// Plugin fields.
	Command   string `yaml:"cmd"`
	Arguments string `yaml:"args"`
}

//go:generate go run ../../cmd/gen.go remediator --plugin=CommandRemediator --name=command

func (r *CommandRemediator) Remediate() RemediationResult {
	out, err := command.ShellCommander(r.Command, r.Arguments).Output()
	if err != nil {
		return RemediationResult{
			Status:   RemediationStatusFailed,
			Messages: []string{command.GetMsgFromCommandError(err)},
		}
	}

	return RemediationResult{
		Status:   RemediationStatusSuccess,
		Messages: []string{string(out)},
	}
}
