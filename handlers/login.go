package handlers

import (
	"log"

	"spi2c/auth"
	"spi2c/data"
	"spi2c/db"

	"spi2c/views/components"

	"github.com/labstack/echo/v4"
)

func LoginForm(c echo.Context) error {
	ff := components.FormFill{
		Values: components.AccountFormValues{},
		Errors: make(map[string]string),
	}

	return Render(c, components.LoginForm(ff))
}

func Login(c echo.Context) error {
	u, ff := validateLoginForm(c)
	if len(ff.Errors) > 0 {
        log.Println(ff.Errors)
		return Render(c, components.LoginForm(ff))
	}

	jwt, err := auth.CreateJWT(u, ff.RememberMe)
	if err != nil {
		log.Println("Failed to create JWT", err)
	}

	err = auth.WriteJWTCookie(c, jwt)
	if err != nil {
		log.Println("Cookie failed to write")
	}

	c.Response().Header().Add("Hx-Reswap", "outerHTML")
	c.Response().Header().Add("Hx-Retarget", "#sign-in-button")
	return Render(c, components.AccountButton())
}

func validateLoginForm(c echo.Context) (data.User, components.FormFill) {
	un := c.FormValue("username")
	pw := c.FormValue("password")
	rm := c.FormValue("remember-me")

	ff := components.FormFill{
		Values: components.AccountFormValues{
			Username: un,
			Password: pw,
		},
		RememberMe: rm != "",
		Errors:     make(map[string]string),
	}

	u, err := db.GetUserAuthInfo(ff.Values.Username)
	if err != nil || !auth.CheckPassword(u.Password, pw) {
		ff.Errors["INVALID_LOGIN"] = "Incorrect username or password"
	}

	return u, ff
}
