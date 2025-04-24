package vm

import (
	"context"
	"time"

	"github.com/ayoubfaouzi/kvm-manager/pkg/log"
	"github.com/google/uuid"

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
	CPU    uint `json:"cpu"`
	Memory uint `json:"mem"`
	Disk   uint `json:"disk"`
}

type service struct {
	repo   Repository
	logger log.Logger
}

// Service encapsulates use case logic for vms.
type Service interface {
	Create(ctx context.Context, input CreateVMRequest) (VM, error)
}

// NewService creates a new File service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// Create creates a new VM.
func (s service) Create(ctx context.Context, req CreateVMRequest) (
	VM, error) {

	id := uuid.New().String()
	now := time.Now().UTC()
	name := "lx-" + OSFlavor + now.Format("-01022006")
	newVM := entity.VM{
		ID:     id,
		Name:   name,
		CPU:    req.CPU,
		Disk:   req.Disk,
		Memory: req.Memory,
		State:  entity.VMStateCreating,
	}
	err := s.repo.Create(ctx, id, newVM)
	if err != nil {
		s.logger.With(ctx).Error(err)
		return VM{}, err
	}

	return VM{newVM}, nil

}
