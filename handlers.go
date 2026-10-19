package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (app *App) getTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := loading(app.db)
	if err != nil {
		log.Printf("ERROR: Unable to load tasks from DB: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error", "Unable to load data.")
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (app *App) getTaskById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Bad Request", "The 'id' parameter must be a valid integer number.")
		return
	}

	targetTask, err := finding(app.db, id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "Task not found", "")
		return
	}
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "Internal server error", "An unexpected error occurred on the server. Please try again later.")
		return
	}

	writeJSON(w, http.StatusOK, targetTask)

}

type TaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func validate(input *TaskInput) []string {
	var problems []string
	if strings.TrimSpace(input.Title) == "" {
		problems = append(problems, "Title is required")
	}
	if strings.TrimSpace(input.Description) == "" {
		problems = append(problems, "Content is required")
	}
	return problems
}

func (app *App) postNewTask(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input TaskInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body", "")
		return
	}

	if problems := validate(&input); len(problems) > 0 {
		writeError(w, http.StatusBadRequest, "Validation failed", strings.Join(problems, "; "))
		return
	}

	id, err := adding(app.db, &input)
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "Internal server error", "An unexpected error occurred on the server. Please try again later.")
		return
	}

	newTask, err := finding(app.db, int(id))
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "Internal server error", "")
		return
	}

	writeJSON(w, http.StatusCreated, newTask)
}

func (app *App) updateTaskById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Bad Request", "The 'id' parameter must be a valid integer number.")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var input TaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body", "")
		return
	}

	if problems := validate(&input); len(problems) > 0 {
		writeError(w, http.StatusBadRequest, "Validation failed", strings.Join(problems, "; "))
		return
	}

	err = editing(app.db, &input, id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "Task not found", "")
		return
	}
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "Internal error", "")
		return
	}

	updatedTask, err := finding(app.db, id)
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "Internal error", "")
		return
	}

	writeJSON(w, http.StatusOK, updatedTask)

}

func (app *App) deleteTaskById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Bad Request", "The 'id' parameter must be a valid integer number.")
		return
	}

	err = deleting(app.db, id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "Task not found", "")
		return
	}
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "Internal error", "")
		return
	}
}
