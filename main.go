package main

import (
	"spi2c/config"
	"spi2c/db"
	"spi2c/env"
	"spi2c/util/policy"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	env.New()
    policy.New()
	db.New(env.DBConn())
	defer db.Close()

	config.ApplyEchoConfig(e)

	config.SetRoutes(e)

	e.Logger.Fatal(e.Start(env.Port()))
}
