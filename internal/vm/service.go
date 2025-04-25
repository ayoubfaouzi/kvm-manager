package vm

import (
	"context"
	"time"

	"github.com/ayoubfaouzi/kvm-manager/pkg/log"

	"github.com/ayoubfaouzi/kvm-manager/internal/entity"
)

const (
	// #TODO: allow user to choose OS image.
	OSFlavor = "linux-alpine"
)

type VM struct {
	entity.VM
}

type CreateVMRequest struct {
	Name          string `json:"name"`
	CPU           uint   `json:"cpu"`
	Memory        uint   `json:"memory"`
	Disk          uint64 `json:"disk"`
	ReadIopsSec   uint64 `json:"read_iops_sec"`
	WriteIopsSec  uint64 `json:"write_iops_sec"`
	ReadBytesSec  uint64 `json:"read_bytes_sec"`
	WriteBytesSec uint64 `json:"write_bytes_sec"`
}

type service struct {
	repo   Repository
	logger log.Logger
}

// Service encapsulates use case logic for vms.
type Service interface {
	Create(ctx context.Context, input CreateVMRequest) (VM, error)
	Get(ctx context.Context, id string) (VM, error)
	List(ctx context.Context, offset, limit int) ([]VM, error)
	Count(ctx context.Context) (int, error)
	Delete(ctx context.Context, id string) (VM, error)
	Start(ctx context.Context, id string) (VM, error)
	Stop(ctx context.Context, id string) (VM, error)
	Restart(ctx context.Context, id string) (VM, error)
	Stats(ctx context.Context, id string) (VM, error)
}

// NewService creates a new File service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// Create creates a new VM.
func (s service) Create(ctx context.Context, req CreateVMRequest) (
	VM, error) {

	now := time.Now().UTC()

	if req.Name == "" {
		req.Name = "lx-" + OSFlavor + now.Format("-01022006")
	}
	newVM, err := s.repo.Create(ctx, req)
	if err != nil {
		s.logger.With(ctx).Error(err)
		return VM{}, err
	}

	return VM{newVM}, nil

}

func (s service) Count(ctx context.Context) (int, error) {

	vms, err := s.List(ctx, 0, 0)
	if err != nil {
		return 0, err
	}
	return len(vms), nil
}

func (s service) Get(ctx context.Context, id string) (
	VM, error) {

	vm, err := s.repo.Get(ctx, id)
	return VM{vm}, err
}

func (s service) List(ctx context.Context, offset, limit int) (
	[]VM, error) {

	vms, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return []VM{}, err
	}

	if limit > 0 {
		vms = vms[offset:limit]
	}

	listVMs := []VM{}
	for _, vm := range vms {
		listVMs = append(listVMs, VM{vm})
	}
	return listVMs, err
}

func (s service) Start(ctx context.Context, id string) (
	VM, error) {

	err := s.repo.Start(ctx, id)
	if err != nil {
		return s.Get(ctx, id)
	}
	return VM{}, err
}

func (s service) Stop(ctx context.Context, id string) (
	VM, error) {

	err := s.repo.Stop(ctx, id)
	if err != nil {
		return s.Get(ctx, id)
	}
	return VM{}, err
}

func (s service) Restart(ctx context.Context, id string) (
	VM, error) {

	err := s.repo.Restart(ctx, id)
	if err != nil {
		return s.Get(ctx, id)
	}
	return VM{}, err
}

func (s service) Delete(ctx context.Context, id string) (
	VM, error) {

	oldVM, err := s.Get(ctx, id)
	if err != nil {
		return VM{}, err
	}
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return VM{}, err
	}
	return oldVM, nil
}

func (s service) Stats(ctx context.Context, id string) (
	VM, error) {

	err := s.repo.Stats(ctx, id)
	if err != nil {
		return s.Get(ctx, id)
	}
	return VM{}, err
}
