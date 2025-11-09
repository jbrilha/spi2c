package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"spi2c/data"
	"spi2c/env"
	"spi2c/util"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// Checks for JWT validity, if OK returns the fn handlerFunc; otherwise returns the altfn
// TODO this can probably be improved but I'm not sure how to do error handling with HTMX yet
func WithJWT(fn echo.HandlerFunc, altfn echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		jwtCookie, err := util.ReadCookie(c, "JWT")
		if err != nil {
			log.Println("err reading cookie in jwt middleware", err)
			if c.Request().Header.Get("HX-Request") != "" {
				// c.Response().Header().Add("HX-Retarget", "#notifications")
				// c.Response().Header().Add("HX-Reswap", "beforeend")
				return altfn(c)
			}
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		} else {
			_, err := ValidateJWT(jwtCookie.Value)
			if err != nil {
				log.Println(err)
				if c.Request().Header.Get("HX-Request") != "" {
					// c.Response().Header().Add("HX-Retarget", "#notifications")
					// c.Response().Header().Add("HX-Reswap", "beforeend")
					return altfn(c)
				}
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
			}

			// claims := token.Claims.(jwt.MapClaims)

			// log.Println("CLAIMS:", claims)
			// log.Println("TOKEN:", token.Raw)

			// c.Set("JWT", token.Raw)

		}
		return fn(c)
	}
}

func CreateJWT(u data.User, remember bool) (string, error) {
	var expiresAt *jwt.NumericDate
	if remember {
		expiresAt = jwt.NewNumericDate(time.Now().Add(720 * time.Hour)) // 30 days
	} else {
		expiresAt = jwt.NewNumericDate(time.Now().Add(30 * time.Minute))
	}

	claims := jwt.RegisteredClaims{
		ExpiresAt: expiresAt,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "spi2c",
		Subject:   u.Username,
		ID:        strconv.Itoa(u.ID),
		// Audience:  []string{"somebody_else"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := env.JWTSecret()

	return token.SignedString([]byte(secret))
}

func WriteJWTCookie(c echo.Context, jwt string) error {
	token, err := ValidateJWT(jwt)
	if err != nil {
		log.Println("Invalid JWT token in cookie")
		return err
	}

	expiry, err := token.Claims.GetExpirationTime()
	if err != nil {
		log.Println("Invalid JWT expiration time")
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "JWT"
	cookie.Path = "/"
	cookie.Value = jwt
	cookie.Expires = expiry.Time
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(cookie)
	return nil
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	secret := env.JWTSecret()

	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

func IsAuthenticated(c context.Context) bool {
	jwtValue := c.Value("JWT")
	if jwtValue == nil {
		// log.Println("JWT token is missing")
		return false
	}

	jwtString, ok := jwtValue.(string)
	if !ok {
		// log.Println("JWT token is not a string") // this should never happen but alas
		return false
	}
	_, err := ValidateJWT(jwtString)
	if err != nil {
		return false
	}

	return true
}

func CheckPassword(encryptedPassword, password string) bool {
	return encryptedPassword == password
}

// func CheckForJWT(fn echo.HandlerFunc, altfn echo.HandlerFunc) echo.MiddlewareFunc {
// 	return func(hf echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			cook, err := util.ReadCookie(c, "JWT")
// 			if err != nil {
// 				log.Println("err jwt middleware", err)
// 				if c.Request().Header.Get("HX-Request") != "" {
// 					return altfn(c)
// 				}
// 				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
// 			} else {
// 				_, err = ValidateJWT(cook.Value)
// 				if err != nil {
// 					log.Println(err)
// 					if c.Request().Header.Get("HX-Request") != "" {
// 						return altfn(c)
// 					}
// 					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
// 				}
//
// 			}
// 			return hf(c)
// 		}
// 	}
// }
