package server

import (
	"net/http"
	"runtime/debug"

	"github.com/ayoubfaouzi/kvm-manager/internal/config"
	"github.com/ayoubfaouzi/kvm-manager/internal/errors"
	"github.com/ayoubfaouzi/kvm-manager/internal/queue"
	"github.com/ayoubfaouzi/kvm-manager/internal/vm"
	"github.com/ayoubfaouzi/kvm-manager/internal/vmmgr"
	"github.com/ayoubfaouzi/kvm-manager/pkg/log"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	// Returned when request body length is null.
	errEmptyBody = "You have sent an empty body."
)

// BuildHandler sets up the HTTP routing and builds an HTTP handler.
func BuildHandler(logger log.Logger, cfg *config.Config, version string,
	trans ut.Translator, p queue.Producer, vmMgr vmmgr.VMManager) http.Handler {

	// Create `echo` instance.
	e := echo.New()

	// Logging middleware.
	e.Use(middleware.LoggerWithConfig(
		middleware.LoggerConfig{
			Format: `{"remote_ip":"${remote_ip}","host":"${host}",` +
				`"method":"${method}","uri":"${uri}","status":${status},` +
				`"latency":${latency},"latency_human":"${latency_human}",` +
				`"bytes_in":${bytes_in},bytes_out":${bytes_out}}` + "\n",
		}))

	// Recover from panic middleware.
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisablePrintStack: true,
	}))

	// Rate limiter middleware.
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	// Add trailing slash for consistent URIs.
	e.Pre(middleware.AddTrailingSlash())

	// Register a custom binder.
	e.Binder = &CustomBinder{b: &echo.DefaultBinder{}}

	// Register a custom fields validator.
	validate := validator.New()
	_ = validate.RegisterValidation("at_least_one_io_throttle", validateVMThrottling)
	e.Validator = &CustomValidator{validator: validate}

	// Setup a custom HTTP error handler.
	e.HTTPErrorHandler = CustomHTTPErrorHandler(trans)

	// Creates a new group for v1.
	g := e.Group("/v1")

	// Create the services and register the handlers.
	vmSvc := vm.NewService(vm.NewRepository(logger, vmMgr), logger)

	// Create the middlewares.
	vmMiddleware := vm.NewMiddleware(vmSvc, logger)

	// Register the handlers.
	vm.RegisterHandlers(g, vmSvc, logger, vmMiddleware.VerifyID)

	return e
}

// CustomValidator holds custom validator.
type CustomValidator struct {
	validator *validator.Validate
}

// Validate performs field validation.
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

// validateVMThrottling checks if at least of the throttling parameters is provided.
func validateVMThrottling(fl validator.FieldLevel) bool {
	req := fl.Parent().Interface().(vm.CreateVMRequest)
	return req.ReadBytesSec != 0 || req.WriteBytesSec != 0 || req.ReadIopsSec != 0 || req.WriteIopsSec != 0
}

// NewBinder initializes custom server binder.
func NewBinder() *CustomBinder {
	return &CustomBinder{b: &echo.DefaultBinder{}}
}

// CustomBinder struct.
type CustomBinder struct {
	b echo.Binder
}

// Bind tries to bind request into interface, and if it does then validate it.
func (cb *CustomBinder) Bind(i interface{}, c echo.Context) error {
	if c.Request().ContentLength == 0 {
		return errors.BadRequest(errEmptyBody)
	}
	if err := cb.b.Bind(i, c); err != nil && err != echo.ErrUnsupportedMediaType {
		return err
	}
	return c.Validate(i)
}

// CustomHTTPErrorHandler handles errors encountered during HTTP request
// processing.
func CustomHTTPErrorHandler(trans ut.Translator) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		l := c.Logger()
		res := errors.BuildErrorResponse(err, trans)
		if res.StatusCode() == http.StatusInternalServerError {
			debug.PrintStack()
			l.Errorf("encountered internal server error: %v", err)
		}
		if err = c.JSON(res.StatusCode(), res); err != nil {
			l.Errorf("failed writing error response: %v", err)
		}
	}
}
