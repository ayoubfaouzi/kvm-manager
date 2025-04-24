package vm

import (
	"context"

	"github.com/ayoubfaouzi/kvm-manager/pkg/log"
	"github.com/google/uuid"

	"github.com/ayoubfaouzi/kvm-manager/internal/entity"
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
	newVM := entity.VM{
		State: entity.VMStatusShutOff,
	}
	err := s.repo.Create(ctx, id, newVM)
	if err != nil {
		s.logger.With(ctx).Error(err)
		return VM{}, err
	}

	return VM{newVM}, nil

}
