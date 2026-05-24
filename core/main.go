package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func parseTodoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
    
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	res, err := ParseTodo(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}

func main() {
	r := chi.NewRouter()

	r.Post("/parse-todo", parseTodoHandler)

	http.ListenAndServe(":80", r)
}
