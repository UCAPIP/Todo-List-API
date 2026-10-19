package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

type App struct {
	db *sql.DB
}

func main() {

	db, err := sql.Open("sqlite", "store.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err := initDB(db); err != nil {
		log.Fatal(err)
	}

	app := &App{db: db}

	http.HandleFunc("GET /tasks", app.getTasks)
	http.HandleFunc("GET /tasks/{id}", app.getTaskById)
	http.HandleFunc("POST /tasks", app.postNewTask)
	http.HandleFunc("DELETE /tasks/{id}", app.deleteTaskById)
	http.HandleFunc("PUT /tasks/{id}", app.updateTaskById)

	log.Println("The server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
