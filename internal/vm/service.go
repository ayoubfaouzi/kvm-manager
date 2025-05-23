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
	CPU           uint   `json:"cpu" validate:"required,gte=1,lte=1024" example:"2"`
	Memory        uint   `json:"memory" validate:"required,gte=128,lte=1048576" example:"8192"` // In MiB
	Disk          uint64 `json:"disk" validate:"required,gte=2,lte=2048000" example:"40"`       // In GiB
	ReadIopsSec   uint64 `json:"read_iops_sec" validate:"at_least_one_io_throttle" example:"500"`
	WriteIopsSec  uint64 `json:"write_iops_sec" validate:"at_least_one_io_throttle" example:"1000"`
	ReadBytesSec  uint64 `json:"read_bytes_sec" validate:"at_least_one_io_throttle" example:"10"`  // In MiB
	WriteBytesSec uint64 `json:"write_bytes_sec" validate:"at_least_one_io_throttle" example:"20"` // In MiB
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
	Delete(ctx context.Context, id string) error
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Restart(ctx context.Context, id string) error
	Stats(ctx context.Context, id string) (interface{}, error)
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

func (s service) Start(ctx context.Context, id string) error {
	return s.repo.Start(ctx, id)
}

func (s service) Stop(ctx context.Context, id string) error {

	return s.repo.Stop(ctx, id)
}

func (s service) Restart(ctx context.Context, id string) error {

	return s.repo.Restart(ctx, id)

}

func (s service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s service) Stats(ctx context.Context, id string) (interface{}, error) {

	stats, err := s.repo.Stats(ctx, id)
	if err != nil {
		return VM{}, err
	}
	return stats, nil
}
