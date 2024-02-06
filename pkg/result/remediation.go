package result

type RemediationStatus string

const (
	RemediationStatusNoSupport RemediationStatus = "no-support"
	RemediationStatusSuccess   RemediationStatus = "success"
	RemediationStatusFailed    RemediationStatus = "failed"
	RemediationStatusPartial   RemediationStatus = "partial"
)

type Remediation struct {
	Status   RemediationStatus
	Messages []string
}
