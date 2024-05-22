package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var (
	sqlDB *sql.DB
)

type User struct {
	ID          int    `json:"-"`
	Name        string `json:"name"`
	UserName    string `json:"username"`
	Bio         string `json:"bio"`
	IsAvailable bool   `json:"is_available"`
}

type Response struct {
	Status  int
	Message string
	Data    interface{}
}

func main() {
	sqlDB = OpenDBConnection()
	router := http.NewServeMux()

	router.HandleFunc("GET /list", Handler)
	router.HandleFunc("GET /list-fix", HandlerFix)

	http.ListenAndServe(":4000", router)
}

func OpenDBConnection() *sql.DB {
	host := "mysql"
	port := "3306"
	user := "root"
	password := "root"
	database := "test"
	sqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local", user, password, host, port, database))
	if err != nil {
		panic(err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	return sqlDB
}

func Handler(w http.ResponseWriter, r *http.Request) {
	users, err := GetListUsers(r.Context())
	if err != nil {
		resp := &Response{
			Status:  http.StatusInternalServerError,
			Message: "ERROR",
		}
		bytes, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(bytes)
		return
	}

	resp := &Response{
		Status:  http.StatusOK,
		Message: "SUCCESS",
		Data:    users,
	}
	bytes, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func HandlerFix(w http.ResponseWriter, r *http.Request) {
	users, err := GetListUsersFix(r.Context())
	if err != nil {
		resp := &Response{
			Status:  http.StatusInternalServerError,
			Message: "ERROR",
		}
		bytes, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(bytes)
		return
	}

	for _, user := range users {
		isAvailable, err := GetIsUsersAvailable(r.Context(), user.ID)
		if err != nil {
			resp := &Response{
				Status:  http.StatusInternalServerError,
				Message: "ERROR",
			}
			bytes, _ := json.Marshal(resp)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bytes)
			return
		}
		user.IsAvailable = isAvailable
	}

	resp := &Response{
		Status:  http.StatusOK,
		Message: "SUCCESS",
		Data:    users,
	}
	bytes, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func GetListUsers(ctx context.Context) ([]*User, error) {
	users := []*User{}
	query := "SELECT id, name, username, bio FROM users"
	rows, err := sqlDB.QueryContext(ctx, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.ID, &user.Name, &user.UserName, &user.Bio); err != nil {
			return nil, err
		}

		isAvailable, err := GetIsUsersAvailable(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		user.IsAvailable = isAvailable

		users = append(users, user)
	}

	return users, nil
}

func GetListUsersFix(ctx context.Context) ([]*User, error) {
	users := []*User{}
	query := "SELECT id, name, username, bio FROM users"
	rows, err := sqlDB.QueryContext(ctx, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.ID, &user.Name, &user.UserName, &user.Bio); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func GetIsUsersAvailable(ctx context.Context, id int) (bool, error) {
	idUser := 0
	query := "SELECT user_id FROM availables WHERE user_id = ?"
	row := sqlDB.QueryRowContext(ctx, query, id)
	err := row.Scan(&idUser)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
