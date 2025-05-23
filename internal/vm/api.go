package vm

import (
	"net/http"

	"github.com/ayoubfaouzi/kvm-manager/pkg/log"
	"github.com/ayoubfaouzi/kvm-manager/pkg/pagination"

	"github.com/ayoubfaouzi/kvm-manager/internal/entity"
	"github.com/labstack/echo/v4"
)

type resource struct {
	service Service
	logger  log.Logger
}

func RegisterHandlers(g *echo.Group, service Service, logger log.Logger, verifyID echo.MiddlewareFunc) {

	res := resource{service, logger}

	g.PUT("/vms/", res.create)
	g.GET("/vms/", res.list)
	g.GET("/vms/:id/", res.get, verifyID)
	g.DELETE("/vms/:id/", res.delete, verifyID)
	g.POST("/vms/:id/start/", res.start, verifyID)
	g.POST("/vms/:id/stop/", res.stop, verifyID)
	g.POST("/vms/:id/restart/", res.restart, verifyID)
	g.GET("/vms/:id/stats/", res.stats, verifyID)
}

func (r resource) create(c echo.Context) error {

	ctx := c.Request().Context()

	var input CreateVMRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Error(err)
		return err
	}

	vm, err := r.service.Create(ctx, input)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct {
		Status  string    `json:"status"`
		Message string    `json:"message"`
		VM      entity.VM `json:"item"`
	}{"ok", "vm created successfully", vm.VM})
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
	err := r.service.Delete(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{"ok", "vm deleted successfully"})

}

func (r resource) start(c echo.Context) error {

	ctx := c.Request().Context()
	id := c.Param("id")
	err := r.service.Start(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{"ok", "vm started successfully"})

}

func (r resource) stop(c echo.Context) error {

	ctx := c.Request().Context()
	id := c.Param("id")
	err := r.service.Stop(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{"ok", "vm stopped successfully"})

}

func (r resource) restart(c echo.Context) error {

	ctx := c.Request().Context()
	id := c.Param("id")
	err := r.service.Restart(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{"ok", "vm restarted successfully"})

}

func (r resource) stats(c echo.Context) error {

	ctx := c.Request().Context()
	id := c.Param("id")
	stats, err := r.service.Stats(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct {
		Status  string      `json:"status"`
		Message string      `json:"message"`
		Stats   interface{} `json:"stats"`
	}{"ok", "stats retrieved successfully", stats})
}
