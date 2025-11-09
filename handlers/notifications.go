package handlers

import (
	"spi2c/views/components"
	"math/rand"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

var notis = []templ.Component{
	components.Alert("0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 ", false),
	components.Alert("max length b4 wrapping is 34", false),
	components.Alert("disappears in 5 seconds", true),
	components.Alert("really long alert message that will probably look really bad? maybe", false),
	components.Alert("disappears in 5 seconds but has longer text", true),
}

func Notif(c echo.Context) error {

	return Render(c, notis[rand.Intn(len(notis))])
}

// Sends an alert with the provided message to the #notifications div;
//
// 'timed' bool determines if it auto-disappears after 5 seconds or not
func alert(c echo.Context, msg string, timed bool) error {
	c.Response().Header().Add("HX-Retarget", "#notifications")
	c.Response().Header().Add("HX-Reswap", "beforeend")

	return Render(c, components.Alert(msg, timed))
}

func Alert(c echo.Context) error {
    a := c.QueryParam("a")

    return alert(c, a, true)
}
