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
	Create(ctx context.Context, id string, vm entity.VM) error
}

// NewRepository creates a new vm repository.
func NewRepository(logger log.Logger, vmMgr vmmgr.VMManager) Repository {
	return repository{logger, vmMgr}
}

// Create saves a new VM in QEMU/KVM server.
// It returns the ID of the newly inserted VM record.
func (r repository) Create(ctx context.Context, key string,
	file entity.VM) error {
	return nil
}
