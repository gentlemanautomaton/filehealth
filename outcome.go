package filehealth

// Outcome records the outcome of an attempted fix for an issue.
type Outcome interface {
	Issue() Issue
	String() string
	Err() error
}
