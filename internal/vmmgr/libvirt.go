package vmmgr

import (
	"context"
	"errors"
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
