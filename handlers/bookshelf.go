package handlers

import (
	"log"
	"spi2c/data"
	"spi2c/db"
	"spi2c/views/bookshelf"

	"github.com/labstack/echo/v4"
)

func AddBook(c echo.Context) error {
	// title := c.FormValue("title")
	// author := c.FormValue("author")

	b := data.Book{
		ISBN13:  "9780241251409",
        Title: "My Life Had Stood A Loaded Gun",
        Authors: []string{"Emily Dickinson"},
        Publishers: []string{"Penguin"},
        Pages: 52,
        PublishDate: "2016",
        Languages: []string{"English"},
	}
	_, err := db.InsertBook(&b)
	if err != nil {
		log.Println("err in insertion:", err)
	}
	log.Println(b.ID)

	return Render(c, bookshelf.AddBook(b))
}

func RemoveBook(c echo.Context) error {
	title := c.FormValue("title")
	author := c.FormValue("author")

	book := data.Book{Title: title, Authors: []string{author}}

	return Render(c, bookshelf.RemoveBook(book))
}

func HandleBook(c echo.Context) error {
	book := data.Book{
		Title:  "AA",
		Authors: []string{"BB", "CC"},
	}
	return Render(c, bookshelf.Show(book))
}

func BookshelfBase(c echo.Context) error {
	return Render(c, bookshelf.Index(db.GetBooks()))
}
