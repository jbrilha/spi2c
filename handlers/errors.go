package handlers

import (
	"spi2c/views/routes"

	"github.com/labstack/echo/v4"
)

func Route404(c echo.Context) error {
	return Render(c, routes.Route404())
}

func NeedLogin(c echo.Context) error {
    return alert(c, "Not logged in!", true)
}
