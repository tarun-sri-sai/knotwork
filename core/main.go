package main

import (
	"encoding/json"
	"flag"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var core *Core

func getTodosHandler(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	res, err := core.repository.GetTaskDurationsBetween(startDate, endDate)
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

	r.Get("/todos", getTodosHandler)

	http.ListenAndServe(":80", r)
}
