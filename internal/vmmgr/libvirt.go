package vmmgr

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"libvirt.org/go/libvirtxml"

	"github.com/ayoubfaouzi/kvm-manager/internal/entity"
	"github.com/ayoubfaouzi/kvm-manager/pkg/log"
	"libvirt.org/go/libvirt"
)

type VMManager struct {
	logger log.Logger
	conn   *libvirt.Connect
	imgDir string
}

const (
	// Base images names.
	linuxAlpineBaseImage = "alpinelinux3.21.qcow2"
)

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

	return VMManager{logger, conn, node.LibVirtImageDir}, nil
}

// CreateVM creates the vm.
func (vmm VMManager) CreateVM(vm entity.VM) (entity.VM, int, error) {

	baseImgName := filepath.Join(vmm.imgDir, linuxAlpineBaseImage)
	// if _, err := os.Stat(baseImgName); os.IsNotExist(err) {
	// 	return entity.VM{}, 400, errors.New(baseImgName + " image not found")
	// }

	destImgName := filepath.Join(vmm.imgDir, vm.Name+".qcow2")
	vmm.logger.Info("Copying image", baseImgName, "to", destImgName)
	// err := copyFile(baseImgName, destImgName)
	// if err != nil {
	// 	return entity.VM{}, 500, err
	// }

	vmm.logger.Info("Resizing image", destImgName, "to", vm.Disk, "GB")
	// err = vmm.ResizeImage(destImgName, int(vm.Disk))
	// if err != nil {
	// 	return entity.VM{}, 500, err
	// }

	domainXML := libvirtxml.Domain{
		Type: "kvm",
		Name: vm.Name,
		Memory: &libvirtxml.DomainMemory{
			Value: uint(vm.Memory),
			Unit:  "MB",
		},
		VCPU: &libvirtxml.DomainVCPU{
			Value: vm.CPU,
		},
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Arch:    "x86_64",
				Machine: "pc",
				Type:    "hvm",
			},
			BootDevices: []libvirtxml.DomainBootDevice{
				{
					Dev: "hd",
				},
			},
		},
		Devices: &libvirtxml.DomainDeviceList{
			Disks: []libvirtxml.DomainDisk{
				{
					Device: "disk",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "qcow2",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: destImgName,
						},
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "sda",
						Bus: "virtio",
					},
				},
			},
			Interfaces: []libvirtxml.DomainInterface{
				{
					Source: &libvirtxml.DomainInterfaceSource{
						Network: &libvirtxml.DomainInterfaceSourceNetwork{
							Network: "default",
						},
					},
					Model: &libvirtxml.DomainInterfaceModel{
						Type: "virtio",
					},
				},
			},
		},
	}
	vmxml, err := domainXML.Marshal()
	if err != nil {
		return entity.VM{}, 500, err
	}
	domain, err := vmm.conn.DomainDefineXML(vmxml)
	if err != nil {
		return entity.VM{}, 500, err
	}
	defer func() {
		err := domain.Free()
		if err != nil {
			return
		}
	}()
	err = domain.Create()
	if err != nil {
		return entity.VM{}, 500, err
	}
	id, err := domain.GetUUIDString()
	if err != nil {
		return entity.VM{}, 500, err
	}
	vmDesc, err := domain.GetXMLDesc(libvirt.DomainXMLFlags(0))
	if err != nil {
		return entity.VM{}, 500, err
	}
	var vmXML libvirtxml.Domain
	err = xml.Unmarshal([]byte(vmDesc), &vmXML)
	if err != nil {
		return entity.VM{}, 500, err
	}

	vm.ID = id
	vm.State = entity.VMStateRunning
	return vm, 200, nil
}

// GetVM gets the vm.
func (vmm VMManager) GetVM(id string) (entity.VM, error) {
	domain, err := vmm.conn.LookupDomainByUUIDString(id)
	if err != nil {
		return entity.VM{}, err
	}

	var vm entity.VM
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
func (vmm VMManager) StartVM(id string) error {
	domain, err := vmm.conn.LookupDomainByUUIDString(id)
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
func (vmm VMManager) DeleteVM(id string) error {
	domain, err := vmm.conn.LookupDomainByUUIDString(id)
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
func (vmm VMManager) StopVM(id string) error {
	domain, err := vmm.conn.LookupDomainByUUIDString(id)
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

// RebootVM stops the vm.
func (vmm VMManager) RebootVM(id string) error {
	domain, err := vmm.conn.LookupDomainByUUIDString(id)
	if err != nil {
		return err
	}
	defer func() {
		err := domain.Free()
		if err != nil {
			return
		}
	}()
	err = domain.Reboot(libvirt.DOMAIN_REBOOT_DEFAULT)
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

// ResizeImage resizes the image.
func (vmm VMManager) ResizeImage(image string, newSize int) error {
	imgInfo, err := exec.Command("qemu-img", "info", image).CombinedOutput()
	if err != nil {
		return err
	}
	lines := strings.Split(string(imgInfo), "\n")
	re := regexp.MustCompile(`^virtual size: (\d+)`)
	var virtualSize string
	for _, line := range lines {
		if re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			virtualSize = matches[1]
			break
		}
	}
	size, err := strconv.Atoi(virtualSize)
	if err != nil {
		return err
	}
	var cmdStrings []string
	if size < newSize {
		cmdStrings = []string{"resize", image, strconv.Itoa(newSize) + "G"}
	} else {
		cmdStrings = []string{"resize", "--shrink", image, strconv.Itoa(newSize) + "G"}
	}
	cmd := exec.Command("qemu-img", cmdStrings...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		vmm.logger.Info(string(output))
		return err
	}
	return nil
}

// copyFile copies the file.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}
