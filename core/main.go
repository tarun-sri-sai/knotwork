package main

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var core *Core

func parseTodoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	res, err := core.repository.GetTaskMapBefore(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}

func main() {
	repo := flag.String("repo", "", "repo type (e.g., git)")
	repoDsn := flag.String("dsn", "", "repo connection string")
	flag.Parse()

	r := chi.NewRouter()

	coreObj, err := NewCore(*repo, *repoDsn)
	if err != nil {
		panic(err)
	}

	core = coreObj

	r.Post("/parse-todo", parseTodoHandler)

	http.ListenAndServe(":80", r)
}
