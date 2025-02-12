package testdata

import "github.com/salsadigitalauorg/shipshape/pkg/remediation"

type TestRemediator struct {
	// Common fields.
	Message string `json:"msg"`

	// Plugin fields.
	ExpectedRemediationResult remediation.RemediationResult `json:"expected-remediation-result"`
}

func (p *TestRemediator) PluginName() string {
	return "test"
}

func (p *TestRemediator) GetRemediationMessage() string {
	return p.Message
}

func (p *TestRemediator) Remediate() remediation.RemediationResult {
	return p.ExpectedRemediationResult
}
