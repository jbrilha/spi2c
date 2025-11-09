package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	"spi2c/data"

	_ "github.com/lib/pq"
	_ "embed"
)

//go:embed queries/users/create_users_table.sql
var createUsersTableSQL string

//go:embed queries/users/drop_users_table.sql
var dropUsersTableSQL string

//go:embed queries/users/insert_user.sql
var insertUserSQL string

//go:embed queries/users/get_all_users.sql
var getAllUsersSQL string

//go:embed queries/users/get_user_auth_info.sql
var getUserAuthInfoSQL string

//go:embed queries/users/insert_user.sql
var insertUrSQL string

//go:embed queries/users/insert_user.sql
var insertUerSQL string

//go:embed queries/users/get_user_by_id.sql
var getUserByIDSQL string

//go:embed queries/users/get_user_by_username.sql
var getUserByUsernameSQL string

//go:embed queries/users/user_exists.sql
var userExistsSQL string

func InsertUserAccount(u *data.User) (int, error) {
	ctx := context.Background()

	tx, err := db.Begin(ctx)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	defer tx.Rollback(ctx)

	query := insertUserSQL

	err = db.QueryRow(ctx, query, u.Username, u.Email, u.Password, u.CreatedAt).Scan(&u.ID)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println(err)
		return 0, err
	}

	return u.ID, nil
}

func GetUserAuthInfo(username string) (data.User, error) {
	query := getUserAuthInfoSQL

	user := data.User{}

	err := db.QueryRow(context.Background(), query, username).Scan(
		&user.ID,
		&user.Username,
		// &user.Email,
		&user.Password,
		// &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(err)
			return data.User{}, err
		}
		log.Println("other err:", err)
		return data.User{}, err
	}

	return user, nil
}

func GetUserByUsername(username string) (data.User, error) {
	query := getUserByUsernameSQL

	user := data.User{}

	err := db.QueryRow(context.Background(), query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		// &user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(err)
			return data.User{}, err
		}
		log.Println("other err:", err)
		return data.User{}, err
	}

	return user, nil
}

func GetUserByID(id int) (data.User, error) {
	query := getUserByIDSQL

	user := data.User{}

	err := db.QueryRow(context.Background(), query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		// &user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(err)
			return data.User{}, err
		}
		log.Println("other err:", err)
		return data.User{}, err
	}

	return user, nil
}

func GetAllUsers() []data.User {
	query := getAllUsersSQL

	var id int
	var username string
	var password string
	var email string
	var createdAt time.Time

	rows, err := db.Query(context.Background(), query)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("nooooooooo roooooooooooows")
		}
		log.Println(err)
	}

	defer rows.Close()

	posts := []data.User{}

	for rows.Next() {
		err := rows.Scan(&id, &username, &password, &email, &createdAt)
		if err != nil {
			log.Println(err)
		}

		posts = append(posts, data.User{
			ID:        id,
			Password:  password,
			Username:  username,
			Email:     email,
			CreatedAt: createdAt,
		})
	}

	return posts
}

func UserExists(un string) (bool, error) {
	query := userExistsSQL

	var exists bool

	err := db.QueryRow(context.Background(), query, un).Scan(&exists)
	if err != nil {
		log.Println(err)
		return false, err
	}

	return exists, nil
}

func createUserTable() {
	query := createUsersTableSQL

	_, err := db.Exec(context.Background(), query)
	if err != nil {
		log.Println(err)
	}
}
