package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"spi2c/auth"
	"spi2c/data"
	"spi2c/db"
	"spi2c/util"
	"spi2c/util/policy"
	"spi2c/views/blog"
	"spi2c/views/routes"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/html"
)

func BlogBase(c echo.Context) error {
	return Render(c, blog.Index())
}

func PostSearch(c echo.Context) error {
	id, err := strconv.Atoi(c.QueryParam("id"))
	// if err != nil {
	// 	log.Println("id err:", err)
	// }

    timestamp, err := time.Parse("01-02-2006 15:04:05:00", c.QueryParam("ts"))
    // if err != nil {
    //     log.Println("ts err:", err)
    // }

	refresh, err := strconv.ParseBool(c.QueryParam("r"))
	// if err != nil {
	// 	log.Println("refresh err:", err)
    // }

	// fetch, err := strconv.ParseBool(c.QueryParam("f"))
	// if err != nil {
	// 	log.Println("refresh err:", err)
    // }

	scroll, err := strconv.ParseBool(c.QueryParam("sc"))
	// if err != nil {
	// 	log.Println("refresh err:", err)
    // }

	limit, err := strconv.Atoi(c.QueryParam("l"))
	if err != nil {
		// log.Println("limit err:", err)
        if refresh { limit =  100 } else { limit = 20 }
	}

	query := strings.TrimSpace(c.QueryParam("q"))

	sp := parseParams(query)
	sp.ID = id
	sp.Timestamp = timestamp
    sp.Limit = limit
    sp.Refresh = refresh

	p, err := db.SearchPosts(sp)
	// p, err := db.SearchPosts(sp)
	if err != nil {
		log.Println(err)
		return Render(c, routes.Route404())
	}

	if c.Request().Header.Get("HX-Request") == "" {
		// if it's not an htmx request it means it was a direct link access,
		// therefore I need to send @layouts.Base along with the results or else
		// it's just the results in plain html (no tailwind etc)
        return Render(c, blog.IndexWComponent(blog.PageElements(), blog.Posts(p, refresh)))
	}

    if len(p) == 0 && !refresh {
        if scroll {
            log.Println("scrolling but no more results")
	        c.Response().Header().Add("HX-Retarget", "#posts-list")
	        c.Response().Header().Add("HX-Reswap", "beforeend")

            return Render(c, blog.NoMorePosts())
        } else if query != "" {
            log.Println("queried something but no results")
            c.Response().Header().Add("HX-Reselect", "#no-posts")
            c.Response().Header().Add("HX-Reswap", "innerHTML")

            return Render(c, blog.NoPosts())
        }
    }

	return Render(c, blog.Posts(p, refresh))
}

func parseParams(query string) db.PostSearchParams {
	sp := db.PostSearchParams{}

    if query == "" {
        return sp
    }

	re := regexp.MustCompile(`"(.*?)"|from:(\S+)|#(\w+)`)

	matches := re.FindAllStringSubmatch(query, -1)
	for _, match := range matches {
		if match[2] != "" { // captured creator
			sp.Creator = match[2]
		} else if match[1] != "" { // captured string between quotes for exact matching
			sp.ExactTerms = append(sp.ExactTerms, match[1])
		} else if match[3] != "" { // captured tags
			sp.Tags = append(sp.Tags, match[3])
		}
	}

	for _, match := range matches {
		query = strings.ReplaceAll(query, match[0], "")
	}

	// query = strings.TrimSpace(query)
	if query != "" {
		sp.FuzzyTerms = strings.Fields(query)
	}

	return sp
}

func CreatorCard(c echo.Context) error {
	username := c.Param("creator")

	u, err := db.GetUserByUsername(username)
	if err != nil {
		log.Println(err)
		return Render(c, routes.Route404())
	}
	log.Println(u)

	return Render(c, blog.CreatorCard(u))
}

func BlogPost(c echo.Context) error {
	param := c.Param("id")

	if idStr := strings.TrimSuffix(param, ".json"); idStr != param {
		return BlogPostJSON(c, idStr)
	}

	id, err := strconv.Atoi(param)
	if err != nil {
		log.Println("Invalid param")
	}

	p, err := db.GetBlogPostByID(id)
	if err != nil {
		log.Println(err)
		return Render(c, routes.Route404())
	}

	go func(id int) {
		err = db.IncrPostViews(id)

		if err != nil {
			log.Println("err in incrPostViews goroutine", err)
		}
	}(id)

	p.Views += 1 // just to reflect current visit on page
	if c.Request().Header.Get("HX-Request") == "" {
        return Render(c, blog.IndexWComponent(blog.Post(p)))
	}
	return Render(c, blog.Post(p))
}

func BlogPostJSON(c echo.Context, idStr string) error {
	id, err := strconv.Atoi(idStr)

	if err != nil {
		log.Println("Invalid param:", idStr)
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"error": "Invalid post ID — should be a digit"},
		)
	}

	p, err := db.GetBlogPostByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]string{"error": "Blog post not found"})
		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, p)
}

func CreateBlogPostForm(c echo.Context) error {
	if c.Request().Header.Get("HX-Request") == "" {
		return Render(c, blog.IndexWComponent(blog.CreatePost()))
	}
	return Render(c, blog.CreatePost())
}

func AddTag(c echo.Context) error {
	tag := c.QueryParam("tag")

	if err := validateHashtag(tag); err != nil {
		return alert(c, err.Error(), true)
	}

	return Render(c, blog.Tag(tag))
}

func CreateBlogPostSubmission(c echo.Context) error {
	title := c.FormValue("title")
	content := strings.TrimSpace(c.FormValue("content"))
	tags := c.FormValue("tags")

	jwtCookie, err := util.ReadCookie(c, "JWT")
	if err != nil {
		return c.JSON(http.StatusBadRequest, data.Post{})
	}
	token, err := auth.ValidateJWT(jwtCookie.Value)
	if err != nil {
		return c.JSON(http.StatusForbidden, data.Post{})
	}
	creator, err := token.Claims.GetSubject()
	if err != nil {
		// TODO probably shouldn't return a teapot here
		return c.JSON(http.StatusTeapot, data.Post{})
	}

	content = policy.Sanitize(content)
	if err := validateHTMLTags(content); err != nil {
		return alert(c, err.Error(), true)
	}

	tagSlice := strings.Fields(tags)
	p := data.Post{
		Creator: creator,
		Title:   title,
		Tags:    tagSlice,
		Content: strings.Join(strings.Split(content, "\n"), "<br/>"),
	}

	_, err = db.InsertBlogPost(&p)
	if err != nil {
		log.Println("err in insertion:", err)
		return alert(c, err.Error(), true)
	}

	p.Views += 1
	c.Response().Header().Add("HX-Push-Url", fmt.Sprintf("/posts/%v", p.ID))
	return Render(c, blog.Post(p))
}

func validateHTMLTags(input string) error {
	reader := strings.NewReader(input)
	tokenizer := html.NewTokenizer(reader)

	var stack []string

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			if tokenizer.Err().Error() == "EOF" {
				if len(stack) > 0 {
					return fmt.Errorf("Unclosed tags: %v", stack)
				} else {
					return nil
				}
			}
			return fmt.Errorf("Error parsing HTML: %v", tokenizer.Err())
		case html.SelfClosingTagToken:
			continue
		case html.StartTagToken:
			tagName, _ := tokenizer.TagName()
            // if string(tagName) == "br" {
            //     continue
            // }
			stack = append(stack, "<"+string(tagName)+">")
		case html.EndTagToken:
			tagName, _ := tokenizer.TagName()
			if len(stack) == 0 {
				return fmt.Errorf("Unexpected closing tag: [</%v>]", string(tagName))
			}
			if stack[len(stack)-1] != ("<" + string(tagName) + ">") {
				return fmt.Errorf("Incorrect tag nesting: %v ... [</%v>]", stack, string(tagName))
			}
			stack = stack[:len(stack)-1]
		}
	}
}

func validateHashtag(tag string) error {
	alphaNum := `^[a-zA-Z0-9_]+$`
	re := regexp.MustCompile(alphaNum)

	if !re.MatchString(tag) {
		return errors.New("Only alphanumeric characters and underscores in tags!")
	}

	return nil
}
