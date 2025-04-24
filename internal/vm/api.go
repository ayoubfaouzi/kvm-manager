package vm

import (
	"net/http"

	"github.com/ayoubfaouzi/kvm-manager/pkg/log"
	"github.com/ayoubfaouzi/kvm-manager/pkg/pagination"

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
	g.GET("/vms/", res.list)
	g.GET("/vms/:id/", res.get)
	g.DELETE("/vms/:id/", res.delete)
	g.POST("/vms/:id/start/", res.start)
	g.POST("/vms/:id/stop/", res.stop)
	g.POST("/vms/:id/restart/", res.restart)
	g.GET("/vms/:id/stats/", res.stats)
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

func (r resource) list(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.Count(ctx)
	if err != nil {
		return err
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	vms, err := r.service.List(ctx, pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = vms
	return c.JSON(http.StatusOK, pages)
}

func (r resource) get(c echo.Context) error {

	ctx := c.Request().Context()
	id := c.Param("id")
	vm, err := r.service.Get(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, vm)
}

func (r resource) delete(c echo.Context) error {

	ctx := c.Request().Context()
	id := c.Param("id")
	vm, err := r.service.Delete(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, vm)
}

func (r resource) start(c echo.Context) error {

	ctx := c.Request().Context()
	id := c.Param("id")
	vm, err := r.service.Start(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, vm)
}

func (r resource) stop(c echo.Context) error {

	ctx := c.Request().Context()
	id := c.Param("id")
	vm, err := r.service.Stop(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, vm)
}

func (r resource) restart(c echo.Context) error {

	ctx := c.Request().Context()
	id := c.Param("id")
	vm, err := r.service.Restart(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, vm)
}

func (r resource) stats(c echo.Context) error {

	ctx := c.Request().Context()
	id := c.Param("id")
	vm, err := r.service.Stats(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, vm)
}
