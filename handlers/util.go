package handlers

import (
	"context"

	"spi2c/util"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func Render(c echo.Context, comp templ.Component) error {
	ctx := c.Request().Context()

	jwtCookie, err := util.ReadCookie(c, "JWT")
	if err == nil {
		ctx = context.WithValue(context.Background(), "JWT", jwtCookie.Value)
		ctx = c.Request().WithContext(ctx).Context()
	}

	return comp.Render(ctx, c.Response())
}
