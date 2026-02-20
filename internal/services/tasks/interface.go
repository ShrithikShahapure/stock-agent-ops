package tasks

// ManagerInterface defines the contract for managing background training tasks.
// Using an interface allows handlers to be tested with mock implementations.
type ManagerInterface interface {
	GetStatus(taskID string) *TaskStatus
	IsRunning(taskID string) bool
	StartTrainParent() (bool, error)
	StartTrainChild(ticker string, chainFn func()) (bool, error)
}
