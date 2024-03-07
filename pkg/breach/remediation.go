package breach

type RemediationStatus string

const (
	RemediationStatusNoSupport RemediationStatus = "no-support"
	RemediationStatusSuccess   RemediationStatus = "success"
	RemediationStatusFailed    RemediationStatus = "failed"
	RemediationStatusPartial   RemediationStatus = "partial"
)

type Remediation struct {
	Status   RemediationStatus `json:",omitempty"`
	Messages []string          `json:",omitempty"`
}
