package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"

	"knotwork/internal/todo/domain"
)

var core *Core

func getTodosHandler(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

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

	endDate, err := core.repository.ParseDate(endDateStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("parse end date: %s", err.Error()), http.StatusBadRequest)
		return
	}

	taskDurations, err := core.repository.GetTaskDurationsBetween(startDateStr, endDateStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("get task data: %s", err.Error()), http.StatusBadRequest)
		return
	}

	var taskInfo domain.TaskInfo

	taskType := r.URL.Query().Get("type")
	switch taskType {
	case "finished":
		taskInfo, err = domain.GetFinishedTaskInfoBetween(taskDurations, endDate, minDays)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	case "abandoned":
		taskInfo, err = domain.GetAbandonedTaskInfoBetween(taskDurations, endDate, minDays)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		taskInfo, err = domain.GetTaskInfoBetween(taskDurations, endDate, minDays)
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
	repo := os.Getenv("REPO_TYPE")
	repoDsn := os.Getenv("REPO_DSN")

	r := chi.NewRouter()

	coreObj, err := NewCore(repo, repoDsn)
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
