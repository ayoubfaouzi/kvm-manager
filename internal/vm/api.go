package vm

import (
	"net/http"

	"github.com/ayoubfaouzi/kvm-manager/pkg/log"

	"github.com/ayoubfaouzi/kvm-manager/internal/errors"
	"github.com/labstack/echo/v4"
)

type resource struct {
	service Service
	logger  log.Logger
}

func RegisterHandlers(g *echo.Group, service Service, logger log.Logger) {

	res := resource{service, logger}

	g.PUT("/vms/", res.create)
}

func (r resource) create(c echo.Context) error {

	ctx := c.Request().Context()

	var input CreateVMRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return errors.BadRequest("")
	}

	vm, err := r.service.Create(ctx, input)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, vm)
}
