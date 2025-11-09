package config

import (
	"spi2c/auth"
	"spi2c/handlers"

	"github.com/labstack/echo/v4"
)

func SetRoutes(e *echo.Echo) {
	e.Static("/public", "public")
	e.File("/favicon.ico", "public/favicon.ico")
	e.File("/robots.txt", "robots.txt")

	e.RouteNotFound("/*", handlers.Route404)

	e.GET("/", handlers.Home)

	e.GET("/login", handlers.LoginForm)
	e.POST("/login", handlers.Login)

	e.GET("/register", handlers.RegisterForm)
	e.POST("/register", handlers.Register)

	e.GET("/bookshelf", handlers.BookshelfBase)
	e.POST("/bookshelf/add-book", auth.WithJWT(handlers.AddBook, handlers.NeedLogin))
	e.DELETE("/bookshelf/remove-book", auth.WithJWT(handlers.RemoveBook, handlers.NeedLogin))
	e.GET("/bookshelf/book", handlers.HandleBook)

	e.GET("/posts", handlers.BlogBase)

    e.GET("/posts/:id", handlers.BlogPost)

    e.GET("/posts/create", handlers.CreateBlogPostForm)
    e.PUT("/posts/create", handlers.AddTag)
    e.POST("/posts/create", auth.WithJWT(handlers.CreateBlogPostSubmission, handlers.NeedLogin))

    e.GET("/posts/search", handlers.PostSearch)

    e.GET("/profile/:username", handlers.ProfileBase)

    e.GET("/alert", handlers.Alert)

	e.GET("/notif", handlers.Notif)
}
