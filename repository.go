package main

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"
)

const schemaTasks = `
CREATE TABLE IF NOT EXISTS Tasks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    title       TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
	user_id     INTEGER
);`

const schemaUsers = `
CREATE TABLE IF NOT EXISTS Users (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    name     TEXT NOT NULL UNIQUE,
    email    TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);`

func initDB(db *sql.DB) error {
	if _, err := db.Exec(schemaUsers); err != nil {
		return err
	}
	if _, err := db.Exec(schemaTasks); err != nil {
		return err
	}
	return nil
}

type Task struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	UserId      int    `json:"userId"`
}

type User struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"password"`
}

type TaskPage struct {
	Data  []Task `json:"data"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Total int    `json:"total"`
}

func loading(db *sql.DB, userId, page, limit, offset int) (TaskPage, error) {

	tasks := TaskPage{Data: []Task{}, Page: page, Limit: limit, Total: offset}

	query := "SELECT id, title, description, created_at, updated_at, user_id FROM Tasks WHERE user_id = ? ORDER BY id LIMIT ? OFFSET ?"

	rows, err := db.Query(query, userId, limit, offset)
	if err != nil {
		return tasks, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.Id, &t.Title, &t.Description, &t.CreatedAt, &t.UpdatedAt, &t.UserId); err != nil {
			return tasks, err
		}

		tasks.Data = append(tasks.Data, t)
	}

	if err := rows.Err(); err != nil {
		return tasks, err
	}

	return tasks, nil
}

func finding(db *sql.DB, taskId, userId int) (*Task, error) {
	var t Task

	err := db.QueryRow(
		"SELECT id, title, description, created_at, updated_at, user_id FROM Tasks WHERE id = ? AND user_id = ?",
		taskId, userId,
	).Scan(&t.Id, &t.Title, &t.Description, &t.CreatedAt, &t.UpdatedAt, &t.UserId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func adding(db *sql.DB, input *TaskInput, userId int) (int64, error) {

	now := time.Now().Format(time.RFC3339)

	res, err := db.Exec("INSERT INTO Tasks (title, description, created_at, updated_at, user_id) VALUES (?, ?, ?, ?, ?)",
		input.Title, input.Description, now, now, userId,
	)

	if err != nil {
		log.Println(err)
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, err
	}
	log.Println(id, "ID Task was created succesfuly")
	return id, nil

}

func editing(db *sql.DB, input *TaskInput, id, userId int) error {
	now := time.Now().Format(time.RFC3339)

	res, err := db.Exec("UPDATE Tasks SET title = ?, description = ?, updated_at = ? WHERE id = ? AND user_id = ?",
		input.Title, input.Description, now, id, userId,
	)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func deleting(db *sql.DB, id, userId int) error {
	res, err := db.Exec("DELETE FROM Tasks WHERE id = ? AND user_id = ?", id, userId)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func createUser(db *sql.DB, name, email, passHash string) (int64, error) {
	res, err := db.Exec("INSERT INTO Users (name, email, password) VALUES (?, ?, ?)",
		name, email, passHash,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, ErrUserExists
		}
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, err
	}

	log.Println(id, "User was registrated succesfuly")
	return id, nil
}

func findUserByUsername(db *sql.DB, email string) (*User, error) {
	var targetUser = User{}

	err := db.QueryRow(
		"SELECT id, name, email, password FROM Users WHERE email = ?",
		email,
	).Scan(&targetUser.Id, &targetUser.Name, &targetUser.Email, &targetUser.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &targetUser, nil
}

func taskOwner(db *sql.DB, taskID int) (int, error) {
	var owner int
	err := db.QueryRow("SELECT user_id FROM Tasks WHERE id = ?", taskID).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return owner, err
}
