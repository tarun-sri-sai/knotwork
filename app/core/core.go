package main

import (
	"fmt"

	"knotwork/internal/todo/adapters"
	"knotwork/internal/todo/ports"
)

type Core struct {
	repository ports.Repository
}

func NewCore(repoType string, repoDsn string) (*Core, error) {
	repositoryAdapters := map[string]func(string) (ports.Repository, error){
		"git": adapters.NewGitRepository,
	}

	repoAdapterFunc, ok := repositoryAdapters[repoType]
	if !ok {
		return nil, fmt.Errorf("unsupported repository type: %s", repoType)
	}

	repository, err := repoAdapterFunc(repoDsn)
	if err != nil {
		return nil, fmt.Errorf("init repository: %w", err)
	}

	return &Core{repository: repository}, nil
}
