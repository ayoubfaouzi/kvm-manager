package vm

import (
	"context"

	"github.com/ayoubfaouzi/kvm-manager/internal/entity"
	"github.com/ayoubfaouzi/kvm-manager/internal/vmmgr"
	"github.com/ayoubfaouzi/kvm-manager/pkg/log"
)

// repository persists files in database.
type repository struct {
	logger log.Logger
	vmMgr  vmmgr.VMManager
}

// Repository encapsulates the logic to access files from the data source.
type Repository interface {
	// Create saves a new VM in the storage.
	Create(ctx context.Context, vm CreateVMRequest) (entity.VM, error)
	// Get retrieves VM information from the server.
	Get(ctx context.Context, id string) (entity.VM, error)
	// List enumerates all VMs.
	List(ctx context.Context, offset, limit int) ([]entity.VM, error)
	// Deletes a VM given its ID.
	Delete(ctx context.Context, id string) error
	// Starts a VM given its ID.
	Start(ctx context.Context, id string) error
	// Stop a VM given its ID.
	Stop(ctx context.Context, id string) error
	// Restart a VM given its ID.
	Restart(ctx context.Context, id string) error
	// Stats returns VM statistics and metrics.
	Stats(ctx context.Context, id string) (interface{}, error)
}

// NewRepository creates a new vm repository.
func NewRepository(logger log.Logger, vmMgr vmmgr.VMManager) Repository {
	return repository{logger, vmMgr}
}

// Create saves a new VM in QEMU/KVM server.
// It returns the ID of the newly inserted VM record.
func (r repository) Create(ctx context.Context, req CreateVMRequest) (entity.VM, error) {

	newVM, _, err := r.vmMgr.CreateVM(entity.VM{
		Name:          req.Name,
		CPU:           req.CPU,
		Memory:        req.Memory,
		Disk:          req.Disk,
		ReadIopsSec:   req.ReadIopsSec,
		WriteIopsSec:  req.WriteIopsSec,
		ReadBytesSec:  req.ReadBytesSec,
		WriteBytesSec: req.WriteBytesSec,
	})
	return newVM, err
}

// Get retrieves VM information.
func (r repository) Get(ctx context.Context, id string) (entity.VM, error) {
	return r.vmMgr.GetVM(id)
}

// List enumerates all VMs.
func (r repository) List(ctx context.Context, offset, limit int) ([]entity.VM, error) {
	return r.vmMgr.ListVMs(true, true)
}

// Deletes a VM given its ID.
func (r repository) Delete(ctx context.Context, id string) error {
	return r.vmMgr.DeleteVM(id)
}

// Starts a VM given its ID.
func (r repository) Start(ctx context.Context, id string) error {
	return r.vmMgr.StartVM(id)
}

// Stop a VM given its ID.
func (r repository) Stop(ctx context.Context, id string) error {
	return r.vmMgr.StopVM(id)
}

// Restart a VM given its ID.
func (r repository) Restart(ctx context.Context, id string) error {
	return r.vmMgr.RebootVM(id)
}

// Stats returns VM statistics and metrics.
func (r repository) Stats(ctx context.Context, id string) (interface{}, error) {
	return r.vmMgr.GetStats(id)
}
