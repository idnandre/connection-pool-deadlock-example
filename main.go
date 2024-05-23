package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var (
	sqlDB *sql.DB
)

type User struct {
	ID       int    `json:"-"`
	Name     string `json:"name"`
	UserName string `json:"username"`
	Bio      string `json:"bio"`
}

type Follow struct {
	SourceID int
	TargetID int
}

type Response struct {
	Status  int
	Message string
	Data    interface{}
}

func main() {
	sqlDB = OpenDBConnection()
	router := http.NewServeMux()

	router.HandleFunc("GET /list-following", Handler)
	router.HandleFunc("GET /list-following-fix", HandlerFix)

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
	userID, _ := strconv.Atoi(r.Header.Get("x-user-id"))

	users, err := GetListFollows(r.Context(), userID)
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
	userID, _ := strconv.Atoi(r.Header.Get("x-user-id"))

	follows, err := GetListFollowsFix(r.Context(), userID)
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

	users := []*User{}

	for _, follow := range follows {
		user, err := GetUserDetail(r.Context(), follow.TargetID)
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

		users = append(users, user)
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

func GetListFollows(ctx context.Context, id int) ([]*User, error) {
	users := []*User{}
	query := "SELECT source_id, target_id FROM follows where source_id = ?"
	rows, err := sqlDB.QueryContext(ctx, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		follow := &Follow{}
		if err := rows.Scan(&follow.SourceID, &follow.TargetID); err != nil {
			return nil, err
		}

		user, err := GetUserDetail(ctx, follow.TargetID)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func GetListFollowsFix(ctx context.Context, id int) ([]*Follow, error) {
	follows := []*Follow{}
	query := "SELECT source_id, target_id FROM follows where source_id = ?"
	rows, err := sqlDB.QueryContext(ctx, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		follow := &Follow{}
		if err := rows.Scan(&follow.SourceID, &follow.TargetID); err != nil {
			return nil, err
		}

		follows = append(follows, follow)
	}

	return follows, nil
}

func GetUserDetail(ctx context.Context, id int) (*User, error) {
	user := &User{}
	query := "SELECT id, name, username, bio FROM users WHERE id = ?"
	row := sqlDB.QueryRowContext(ctx, query, id)
	err := row.Scan(&user.ID, &user.Name, &user.UserName, &user.Bio)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
