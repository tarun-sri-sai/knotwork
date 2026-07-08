package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"

	"knotwork/internal/todo/domain"
)

var core *Core

func getTodosHandler(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	minDaysStr := r.URL.Query().Get("minDays")
	minDays := 0
	var err error
	if minDaysStr != "" {
		minDays, err = strconv.Atoi(minDaysStr)
		if err != nil {
			http.Error(w, "invalid minDays parameter", http.StatusBadRequest)
			return
		}
	}

	taskType := r.URL.Query().Get("type")

	var taskInfo domain.TaskInfo
	switch taskType {
	case "finished":
		taskInfo, err = core.repository.GetFinishedTaskInfoBetween(startDate, endDate, minDays)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	case "abandoned":
		taskInfo, err = core.repository.GetAbandonedTaskInfoBetween(startDate, endDate, minDays)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		taskInfo, err = core.repository.GetTaskInfoBetween(startDate, endDate, minDays)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(taskInfo)
}

func main() {
	repo := flag.String("repo", "", "repo type (e.g., git)")
	repoDsn := flag.String("dsn", "", "repo connection string")
	flag.Parse()

	r := chi.NewRouter()

	coreObj, err := NewCore(*repo, *repoDsn)
	if err != nil {
		log.Fatalf("failed to initialize core: %v\n", err)
	}

	core = coreObj

	r.Get("/todos", getTodosHandler)

	port := 80
	if s, ok := os.LookupEnv("PORT"); ok && s != "" {
		if p, err := strconv.Atoi(s); err == nil && p > 0 && p <= 0xffff {
			port = p
		}
	}

	log.Printf("serving on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
