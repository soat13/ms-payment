package domain

type Status string

const (
	StatusPending    Status = "PENDING"
	StatusProcessing Status = "PROCESSING"
	StatusSucceeded  Status = "SUCCEEDED"
	StatusFailed     Status = "FAILED"
	StatusError      Status = "ERROR"
)

func (s Status) IsAttemptResult() bool {
	return s == StatusSucceeded || s == StatusFailed || s == StatusError
}
