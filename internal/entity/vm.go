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
	Disk          uint64      `json:"disk"`
	DiskPath      string      `json:"-"`
	ReadIopsSec   uint64      `json:"read_iops_sec,omitempty"`
	WriteIopsSec  uint64      `json:"write_iops_sec,omitempty"`
	TotalIopsSec  uint64      `json:"total_iops_sec,omitempty"`
	ReadBytesSec  uint64      `json:"read_bytes_sec,omitempty"`
	WriteBytesSec uint64      `json:"write_bytes_sec,omitempty"`
	TotalBytesSec uint64      `json:"total_bytes_sec,omitempty"`
}
