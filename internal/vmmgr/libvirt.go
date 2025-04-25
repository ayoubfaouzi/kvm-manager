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

type VMState struct {
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryPercent float64 `json:"mem_percent"`
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
	if _, err := os.Stat(baseImgName); os.IsNotExist(err) {
		return entity.VM{}, 400, errors.New(baseImgName + " image not found")
	}

	destImgName := filepath.Join(vmm.imgDir, vm.Name+".qcow2")
	vmm.logger.Info("Copying image", baseImgName, "to", destImgName)
	err := copyFile(baseImgName, destImgName)
	if err != nil {
		return entity.VM{}, 500, err
	}

	vmm.logger.Info("Resizing image", destImgName, "to", vm.Disk, "GB")
	err = vmm.ResizeImage(destImgName, int(vm.Disk))
	if err != nil {
		return entity.VM{}, 500, err
	}

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
					IOTune: &libvirtxml.DomainDiskIOTune{
						ReadBytesSec:  vm.ReadBytesSec * 1024 * 1024,
						WriteBytesSec: vm.WriteBytesSec * 1024 * 1024,
						ReadIopsSec:   vm.ReadIopsSec,
						WriteIopsSec:  vm.WriteIopsSec,
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

	info, err := domain.GetInfo()
	if err != nil {
		return entity.VM{}, err
	}
	vm.Memory = uint(info.Memory) / 1024
	vm.CPU = info.NrVirtCpu

	xmlDesc, err := domain.GetXMLDesc(0)
	if err != nil {
		return entity.VM{}, err
	}

	type Domain struct {
		Devices struct {
			Disks []struct {
				Target struct {
					Dev string `xml:"dev,attr"`
				} `xml:"target"`
			} `xml:"disk"`
		} `xml:"devices"`
	}

	var dom Domain
	if err := xml.Unmarshal([]byte(xmlDesc), &dom); err != nil {
		return entity.VM{}, err
	}

	for _, disk := range dom.Devices.Disks {
		info, err := domain.GetBlockInfo(disk.Target.Dev, 0)
		if err != nil {
			vmm.logger.Debugf("Could not get info for %s: %v", disk.Target.Dev, err)
			continue
		}
		vm.Disk = info.Capacity / (1 << 30)

		stats, err := domain.GetBlockIoTune(disk.Target.Dev, 0)
		if err != nil {
			vmm.logger.Debugf("Failed to get block stats: %v", err)
		}
		vm.ReadBytesSec = stats.ReadBytesSec / 1024 / 1024
		vm.WriteBytesSec = stats.WriteBytesSec / 1024 / 1024
		vm.TotalBytesSec = stats.TotalBytesSec / 1024 / 1024
		vm.ReadIopsSec = stats.ReadIopsSec
		vm.WriteIopsSec = stats.WriteIopsSec
		vm.TotalIopsSec = stats.TotalIopsSec
		break
	}

	vm.State = ParseState(state)
	vm.ID = id
	return vm, nil
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
		return fmt.Errorf("failed to destroy domain: %w", err)
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
		return []entity.VM{}, fmt.Errorf("failed to list domains: %w", err)
	}
	for _, domain := range domains {
		defer domain.Free()
	}
	vms := make([]entity.VM, 0, len(domains))
	for _, domain := range domains {
		id, err := domain.GetUUIDString()
		if err != nil {
			return []entity.VM{}, fmt.Errorf("failed to get domain id: %w", err)
		}
		vm, err := vmm.GetVM(id)
		if err != nil {
			return []entity.VM{}, err
		}
		vms = append(vms, vm)
	}

	return vms, nil
}

// GetStats returns CPU and memory stats.
func (vmm VMManager) GetStats(id string) (VMState, error) {

	domain, err := vmm.conn.LookupDomainByUUIDString(id)
	if err != nil {
		return VMState{}, err
	}

	// Get CPU usage
	cpuStart, err := domain.GetCPUStats(-1, 1, 0)
	if err != nil {
		return VMState{}, err
	}

	startTime := cpuStart[0].CpuTime
	time.Sleep(1 * time.Second)
	cpuEnd, err := domain.GetCPUStats(-1, 1, 0)
	if err != nil {
		return VMState{}, err
	}
	endTime := cpuEnd[0].CpuTime

	info, err := domain.GetInfo()
	if err != nil {
		return VMState{}, err
	}

	// CPU usage percentage over the interval
	cpuDelta := endTime - startTime // in nanoseconds
	cpuUsage := float64(cpuDelta) / (1e9 * float64(info.NrVirtCpu)) * 100

	// Memory Usage
	memStats, err := domain.MemoryStats(11, 0)
	if err != nil {
		return VMState{}, err
	}

	var available, unused uint64
	for _, stat := range memStats {
		switch stat.Tag {
		case int32(libvirt.DOMAIN_MEMORY_STAT_AVAILABLE):
			available = stat.Val
		case int32(libvirt.DOMAIN_MEMORY_STAT_UNUSED):
			unused = stat.Val
		}
	}

	used := available - unused
	memPercent := float64(used) / float64(available) * 100
	return VMState{
		CPUUsage:      cpuUsage,
		MemoryPercent: memPercent,
	}, nil

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
