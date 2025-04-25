package vm

import (
	"strings"

	"github.com/ayoubfaouzi/kvm-manager/internal/errors"
	"github.com/ayoubfaouzi/kvm-manager/pkg/log"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type middleware struct {
	service Service
	logger  log.Logger
}

// NewMiddleware creates a new file Middleware.
func NewMiddleware(service Service, logger log.Logger) middleware {
	return middleware{service, logger}
}

// VerifyID is the middleware function.
func (m middleware) VerifyID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		id := strings.ToLower(c.Param("id"))
		_, err := uuid.Parse(id)
		if err != nil {
			m.logger.Error("failed to match VM UUID %s", id)
			return errors.BadRequest("invalid VM id")
		}

		return next(c)
	}
}
