package main

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

const schemaTasks = `
CREATE TABLE IF NOT EXISTS Tasks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    title       TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);`

const schemaUsers = `
CREATE TABLE IF NOT EXISTS Users (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    name     TEXT NOT NULL,
    email    TEXT NOT NULL,
    password TEXT NOT NULL
);`

func initDB(db *sql.DB) error {
	_, err := db.Exec(schemaTasks)
	_, err = db.Exec(schemaUsers)
	return err
}

type Task struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func loading(db *sql.DB) ([]Task, error) {

	tasks := []Task{}

	query := "SELECT id, title, description, created_at, updated_at FROM Tasks"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.Id, &t.Title, &t.Description, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return tasks, err
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return tasks, err
	}

	return tasks, nil
}

func finding(db *sql.DB, id int) (*Task, error) {
	var t Task

	err := db.QueryRow(
		"SELECT id, title, description, created_at, updated_at FROM Tasks WHERE id = ?",
		id,
	).Scan(&t.Id, &t.Title, &t.Description, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func adding(db *sql.DB, input *TaskInput) (int64, error) {

	now := time.Now().Format(time.RFC3339)

	res, err := db.Exec("INSERT INTO Tasks (title, description, created_at, updated_at) VALUES (?, ?, ?, ?)",
		input.Title, input.Description, now, now,
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

func editing(db *sql.DB, input *TaskInput, id int) error {
	now := time.Now().Format(time.RFC3339)

	res, err := db.Exec("UPDATE Tasks SET title = ?, description = ?, updated_at = ? WHERE id = ?",
		input.Title, input.Description, now, id,
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

func deleting(db *sql.DB, id int) error {
	res, err := db.Exec("DELETE FROM Tasks WHERE id = ?", id)
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
