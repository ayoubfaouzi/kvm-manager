package entity

// VMStateType represents the VM running state type.
type VMStateType string

// VM running state.
const (
	VMStatusCreating VMStateType = "creating"
	VMStatusRunning  VMStateType = "running"
	VMStatusStarting VMStateType = "starting"
	VMStatusStopping VMStateType = "stopping"
	VMStatusPaused   VMStateType = "paused"
	VMStatusShutdown VMStateType = "shutdown"
	VMStatusShutOff  VMStateType = "shutoff"
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
