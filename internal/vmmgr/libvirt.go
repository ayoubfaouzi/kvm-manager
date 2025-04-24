package vmmgr

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ayoubfaouzi/kvm-manager/internal/entity"
	"github.com/ayoubfaouzi/kvm-manager/pkg/log"
	"libvirt.org/go/libvirt"
)

type VMManager struct {
	logger log.Logger
	conn   *libvirt.Connect
}

// New initializes the VM manager service.
func New(logger log.Logger, node entity.NodeInstance) (VMManager, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := libvirt.NewConnect(node.LibVirtURI)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return VMManager{}, errors.New("connection to libvirt daemon timed out")
		}
		return VMManager{}, err
	}

	return VMManager{logger, conn}, nil
}

// GetVM gets the vm.
func (vmm VMManager) GetVM(vm entity.VM) (entity.VM, error) {
	domain, err := vmm.conn.LookupDomainByUUIDString(vm.ID)
	if err != nil {
		return entity.VM{}, err
	}
	vm.Name, err = domain.GetName()
	if err != nil {
		return entity.VM{}, err
	}
	state, _, err := domain.GetState()
	if err != nil {
		return entity.VM{}, err
	}
	vm.State = ParseState(state)
	return entity.VM{}, nil
}

// StartVM starts the vm.
func (vmm VMManager) StartVM(vm entity.VM) error {
	domain, err := vmm.conn.LookupDomainByUUIDString(vm.ID)
	if err != nil {
		return err
	}
	defer func() {
		err := domain.Free()
		if err != nil {
			return
		}
	}()
	err = domain.Create()
	if err != nil {
		return err
	}
	return nil
}

// DeleteVM deletes the vm.
func (vmm VMManager) DeleteVM(vm entity.VM) error {
	domain, err := vmm.conn.LookupDomainByUUIDString(vm.ID)
	if err != nil {
		return fmt.Errorf("failed to lookup domain: %w", err)
	}
	defer func() {
		err := domain.Free()
		if err != nil {
			return
		}
	}()
	active, err := domain.IsActive()
	if err != nil {
		return fmt.Errorf("failed to check domain status: %w", err)
	}

	if active {
		err = domain.Destroy()
		if err != nil {
			return fmt.Errorf("failed to destroy domain: %w", err)
		}
	}

	err = domain.Undefine()
	if err != nil {
		return fmt.Errorf("failed to undefine domain: %w", err)
	}
	return nil
}

// StopVM stops the vm.
func (vmm VMManager) StopVM(vm entity.VM) error {
	domain, err := vmm.conn.LookupDomainByUUIDString(vm.ID)
	if err != nil {
		return err
	}
	defer func() {
		err := domain.Free()
		if err != nil {
			return
		}
	}()
	err = domain.Destroy()
	if err != nil {
		return err
	}
	return nil
}

// ListVMs lists the vms.
func (vmm VMManager) ListVMs(active, inactive bool) ([]entity.VM, error) {
	var flags libvirt.ConnectListAllDomainsFlags
	if active {
		flags |= libvirt.CONNECT_LIST_DOMAINS_ACTIVE
	}
	if inactive {
		flags |= libvirt.CONNECT_LIST_DOMAINS_INACTIVE
	}
	domains, err := vmm.conn.ListAllDomains(flags)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}
	for _, domain := range domains {
		defer domain.Free()
	}
	vms := make([]entity.VM, 0, len(domains))
	for _, domain := range domains {
		name, err := domain.GetName()
		if err != nil {
			return nil, fmt.Errorf("failed to get domain name: %w", err)
		}
		state, _, err := domain.GetState()
		if err != nil {
			return nil, fmt.Errorf("failed to get domain state: %w", err)
		}
		id, err := domain.GetUUIDString()
		if err != nil {
			return nil, fmt.Errorf("failed to get domain id: %w", err)
		}
		vm := entity.VM{
			Name:  name,
			State: ParseState(state),
			ID:    id,
		}
		vms = append(vms, vm)
	}

	return vms, nil
}

// ParseState parses the state of the vm.
func ParseState(state libvirt.DomainState) entity.VMStateType {
	switch state {
	case libvirt.DOMAIN_NOSTATE:
		return entity.VMStateNoState
	case libvirt.DOMAIN_RUNNING:
		return entity.VMStateRunning
	case libvirt.DOMAIN_BLOCKED:
		return entity.VMStateBlocked
	case libvirt.DOMAIN_PAUSED:
		return entity.VMStatePaused
	case libvirt.DOMAIN_SHUTDOWN:
		return entity.VMStateShutdown
	case libvirt.DOMAIN_SHUTOFF:
		return entity.VMStateShutOff
	case libvirt.DOMAIN_CRASHED:
		return entity.VMStateCrashed
	case libvirt.DOMAIN_PMSUSPENDED:
		return entity.VMStatePMSuspended
	}
	return entity.VMStateUnknown
}
