package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	"spi2c/data"

	sq "github.com/Masterminds/squirrel"

	_ "embed"
)

type PostSearchParams struct {
	Creator    string
	Tags       []string
	FuzzyTerms []string
	ExactTerms []string
	ID         int
	Timestamp  time.Time
	Limit      int
	Refresh    bool
}

//go:embed queries/posts/create_posts_table.sql
var createPostsTableSQL string

//go:embed queries/posts/drop_posts_table.sql
var dropPostsTableSQL string

//go:embed queries/posts/get_post_by_id.sql
var getPostByIDSQL string

//go:embed queries/posts/get_posts_by_creator.sql
var getPostsByCreatorSQL string

//go:embed queries/posts/get_posts_by_tag.sql
var getPostsByTagSQL string

//go:embed queries/posts/insert_post.sql
var insertPostSQL string

//go:embed queries/posts/increment_post_views.sql
var incrementPostViewsSQL string

func InsertBlogPost(p *data.Post) (int, error) {
	tx, err := db.Begin(context.Background())
	if err != nil {
		log.Println(err)
		return 0, err
	}

	defer tx.Rollback(context.Background())

	query := insertPostSQL

	err = tx.QueryRow(
		context.Background(),
		query,
		p.Creator,
		p.Title,
		p.Tags,
		p.Content,
		time.Now(),
	).Scan(
		&p.ID,
	)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	if err = tx.Commit(context.Background()); err != nil {
		log.Println(err)
		return 0, err
	}

	return p.ID, nil
}

func IncrPostViews(id int) error {
	tx, err := db.Begin(context.Background())
	if err != nil {
		log.Println(err)
		return err
	}

	defer tx.Rollback(context.Background())

	query := incrementPostViewsSQL

	_, err = tx.Exec(context.Background(), query, id)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("nooooooooo %d", id)
			return err
		}
		log.Println("wtf", id)
		return err
	}

	if err = tx.Commit(context.Background()); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetBlogPostByID(id int) (data.Post, error) {
	query := getPostByIDSQL

	post := data.Post{}

	err := db.QueryRow(context.Background(), query, id).Scan(
		&post.ID,
		&post.Creator,
		&post.Title,
		&post.Tags,
		&post.Content,
		&post.Views,
		&post.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("nooooooooo %d", id)
			return data.Post{}, err
		}
		log.Println("wtf", id)
		return data.Post{}, err
	}

	return post, nil
}

func SearchPosts(sp PostSearchParams) ([]data.Post, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.
		Select("*").
		From("posts").
		OrderBy("created_at DESC").
		Limit(uint64(sp.Limit))

	if sp.Refresh {
		query = query.Where(sq.Gt{"created_at": sp.Timestamp})
	} else {
		query = query.Where(sq.Lt{"created_at": sp.Timestamp})
	}

	if sp.Creator != "" {
		query = query.Where(sq.Eq{"creator": sp.Creator})
	}

	var ands []sq.Sqlizer

	if len(sp.FuzzyTerms) > 0 {
		var fuzzy_clauses []sq.Sqlizer
		for _, term := range sp.FuzzyTerms {
			pattern := "%" + term + "%"
			fuzzy_clauses = append(fuzzy_clauses,
				sq.Or{
					sq.ILike{"content": pattern},
					sq.ILike{"title": pattern},
				},
			)
		}
		ands = append(ands, sq.Or(fuzzy_clauses))
	}

	if len(sp.ExactTerms) > 0 {
		var exact_clauses []sq.Sqlizer
		for _, term := range sp.ExactTerms {
			regex := `\y` + term + `\y`
			exact_clauses = append(exact_clauses,
				sq.Or{
					sq.Expr("content ~* ?", regex),
					sq.Expr("title ~* ?", regex),
				},
			)
		}
		ands = append(ands, sq.And(exact_clauses))
	}

	if len(sp.Tags) > 0 {
		ands = append(ands,
			sq.Expr("tags && ?", sp.Tags),
		)
	}

	if len(ands) > 0 {
		query = query.Where(sq.And(ands))
	}

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	return getPosts(sqlStr, args...)
}

func SearchPostsByCreator(creator string) ([]data.Post, error) {
	query := getPostsByCreatorSQL

	return getPosts(query, creator)
}

func SearchPostsByTag(tag string) ([]data.Post, error) {
	query := getPostsByTagSQL

	return getPosts(query, tag)
}

func getPosts(query string, args ...any) ([]data.Post, error) {
	rows, err := db.Query(context.Background(), query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("nooooooooo roooooooooooows")
		}
		log.Println(err)
	}

	defer rows.Close()

	posts := []data.Post{}

	for rows.Next() {
		var post data.Post
		err := rows.Scan(
			&post.ID,
			&post.Creator,
			&post.Title,
			&post.Tags,
			&post.Content,
			&post.Views,
			&post.CreatedAt,
		)
		if err != nil {
			log.Println(err)
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func createPostsTable() {
	query := createPostsTableSQL

	_, err := db.Exec(context.Background(), query)
	if err != nil {
		log.Println(err)
	}
}

// func SearchPosts(sp PostSearchParams) ([]data.Post, error) {
// 	var query strings.Builder
// 	args := []any{}
// 	offset := 2
// 	fCount := len(sp.FuzzyTerms)
// 	eCount := len(sp.ExactTerms)
// 	tCount := len(sp.Tags)
//
// 	query.WriteString(`SELECT * FROM post WHERE `)
// 	if sp.Creator == "" {
// 		if sp.Refresh {
// 			query.WriteString(`(created_at > $1)`)
// 		} else {
// 			query.WriteString(`(created_at < $1)`)
// 		}
// 	} else {
// 		query.WriteString(`creator = $1 AND `)
// 		if sp.Refresh {
// 			query.WriteString(`(created_at > $2)`)
// 		} else {
// 			query.WriteString(`(created_at < $2)`)
// 		}
// 		offset = 3
//
// 		args = []any{sp.Creator}
// 	}
// 	args = append(args, sp.Timestamp)
//
// 	if tCount > 0 || fCount > 0 || eCount > 0 {
// 		query.WriteString(` AND (`)
// 	}
//
// 	q := []string{}
//
// 	if fCount > 0 {
// 		query.WriteString(`(`)
// 		for _, term := range sp.FuzzyTerms { // TODO already looping through terms in the handler, figure out optimization
// 			q = append(
// 				q,
// 				`(content ILIKE '%' || $`+fmt.Sprint(offset)+
// 					` || '%' OR title ILIKE '%' || $`+fmt.Sprint(offset)+` || '%')`)
// 			args = append(args, term)
//
// 			offset += 1
// 		}
// 		query.WriteString(strings.Join(q, ` OR `))
// 		query.WriteString(`)`)
// 	}
//
// 	if eCount > 0 {
// 		q = []string{}
// 		if fCount > 0 {
// 			query.WriteString(" AND (")
// 		}
//
// 		for _, term := range sp.ExactTerms {
// 			q = append(
// 				q,
// 				`(content ~* ('\y' || $`+fmt.Sprint(offset)+
// 					` || '\y') OR title ~* ('\y' || $`+fmt.Sprint(offset)+` || '\y'))`)
// 			args = append(args, term)
//
// 			offset += 1
// 		}
// 		query.WriteString(strings.Join(q, " OR "))
//
// 		if fCount > 0 {
// 			query.WriteString(")")
// 		}
// 	}
//
// 	if tCount > 0 {
// 		if fCount > 0 || eCount > 0 {
// 			query.WriteString(" AND (")
// 		}
//
// 		query.WriteString("tags && $" + fmt.Sprint(offset))
// 		args = append(args, sp.Tags)
//
// 		if fCount > 0 || eCount > 0 {
// 			query.WriteString(")")
// 		}
// 	}
//
// 	if tCount > 0 || fCount > 0 || eCount > 0 {
// 		query.WriteString(`)`)
// 	}
//
// 	args = append(args, sp.Limit)
// 	query.WriteString(` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(offset))
//
// 	return getPosts(query.String(), args...)
// }
