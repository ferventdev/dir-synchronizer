package model

//Status of the sync operation.
type Status string

const (
	Scheduled  = "scheduled"
	InProgress = "in_progress"
	Complete   = "complete"
	Canceled   = "canceled"
)

// Operation - synchronization operation between the dir entry in the source directory and same entry in the copy directory.
type Operation struct {
	status Status
}
