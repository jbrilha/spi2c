package handlers

import (
	"log"
	"time"

	"spi2c/auth"
	"spi2c/data"
	"spi2c/db"
	"spi2c/views/components"

	"github.com/labstack/echo/v4"
)

func RegisterForm(c echo.Context) error {
	ff := components.FormFill{
		Values: components.AccountFormValues{},
		Errors: make(map[string]string),
	}

	return Render(c, components.RegisterForm(ff))
}

func Register(c echo.Context) error {
	ff := validateRegisterForm(c)
	if len(ff.Errors) > 0 {
		return Render(c, components.RegisterForm(ff))
	}

	u := data.User{
		Username:  ff.Values.Username,
		Email:     "nomail",
		Password:  ff.Values.Password,
		CreatedAt: time.Now(),
	}

	_, err := db.InsertUserAccount(&u)
	if err != nil {
		log.Println("err in insetion:", err)
	}
	log.Println(u.ID)

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

func validateRegisterForm(c echo.Context) components.FormFill {
	un := c.FormValue("username")
	pw := c.FormValue("password")
	pwc := c.FormValue("confirmation")
	rm := c.FormValue("remember-me")

	ff := components.FormFill{
		Values: components.AccountFormValues{
			Username:     un,
			Password:     pw,
			Confirmation: pwc,
		},
        RememberMe: rm != "",
		Errors: make(map[string]string),
	}
	if len(pw) < 5 {
		ff.Errors["PW_LEN"] = "Password length must be at least 5"
	}
	if pw != pwc {
		ff.Errors["PW_CONF"] = "Confirmation does not match password"
	}

	exists, _ := db.UserExists(un)
	if exists {
		ff.Errors["USER_EXISTS"] = "Username already taken"
	}

	return ff
}
