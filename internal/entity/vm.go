package entity

// VMStateType represents the VM running state type.
type VMStateType string

// VM running state.
const (
	VMStateNoState     VMStateType = "nostate"
	VMStateRunning     VMStateType = "running"
	VMStateBlocked     VMStateType = "blocked"
	VMStatePaused      VMStateType = "paused"
	VMStateShutdown    VMStateType = "shutdown"
	VMStateShutOff     VMStateType = "shutoff"
	VMStateCrashed     VMStateType = "crashed"
	VMStatePMSuspended VMStateType = "pmsuspended"

	// Other state for tracking progress.
	VMStateUnknown  VMStateType = "unknown"
	VMStateCreating VMStateType = "creating"
	VMStateStarting VMStateType = "starting"
	VMStateStopping VMStateType = "stopping"
)

// VM represents a virtual machine object.
type VM struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	State         VMStateType `json:"state"`
	CPU           uint        `json:"cpu"`
	Memory        uint        `json:"memory"`
	Disk          uint        `json:"disk"`
	ReadIopsSec   uint        `json:"read_iops_sec"`
	WriteIopsSec  uint        `json:"write_iops_sec"`
	TotalIopsSec  uint        `json:"total_iops_sec"`
	ReadBytesSec  uint        `json:"read_bytes_sec"`
	WriteBytesSec uint        `json:"write_bytes_sec"`
	TotalBytesSec uint        `json:"total_bytes_sec"`
}
