package entity

// VMStateType represents the VM running state type.
type VMStateType uint8

// VM running state.
const (
	VMStatusRunning  VMStateType = 1
	VMStatusPaused   VMStateType = 2
	VMStatusShutdown VMStateType = 3
	VMStatusShutOff  VMStateType = 4
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
