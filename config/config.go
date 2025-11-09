package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	ansi "spi2c/util"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func ApplyEchoConfig(e *echo.Echo) {
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogProtocol:      true,
		LogLatency:       true,
		LogMethod:        true,
		LogStatus:        true,
		LogHost:          true,
		LogURI:           true,
		LogError:         true,
		LogContentLength: true,
		HandleError:      true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			dT := v.StartTime.Format(time.DateTime)
			method := fmt.Sprintf("%s%v%s", ansi.BoldPurple, v.Method, ansi.None)
			url := fmt.Sprintf("%s%v%v%s", ansi.Cyan, v.Host, v.URI, ansi.None)

			latColor := ansi.Green
			if v.Latency > 10*time.Second {
				latColor = ansi.Red
			} else if v.Latency >= 5*time.Second {
				latColor = ansi.Orange
			} else if v.Latency >= 1*time.Second {
				latColor = ansi.Yellow
			}
			lat := fmt.Sprintf("%s%v%s", latColor, v.Latency, ansi.None)

			reqConLen := v.ContentLength
			if reqConLen == "" {
				reqConLen = "0"
			}
			reqConLen = fmt.Sprintf("%s%vB%s", ansi.BoldBlue, reqConLen, ansi.None)

			resConLen := strconv.FormatInt(c.Response().Size, 10)
			resConLen = fmt.Sprintf("%s%vB%s", ansi.Blue, resConLen, ansi.None)

			if v.Error == nil {
				log := "%v: %v %v [%v] — ꜜ%v ꜛ%v in %v\n"

				status := fmt.Sprintf("%s%v%s", ansi.BoldGreen, v.Status, ansi.None)

				fmt.Printf(log, dT, method, url, status, reqConLen, resConLen, lat)
			} else {
				log := "%v: %v %v [%v] — ꜜ%v ꜛ%v in %v\nError message: %v\n"

				status := fmt.Sprintf("%s%v%s", ansi.BoldRed, v.Status, ansi.None)
				error := strings.Split(v.Error.Error(), "message=")[1]

				fmt.Printf(log, dT, method, url, status, reqConLen, resConLen, lat, error)
			}

			return nil
		},
	}))
}
