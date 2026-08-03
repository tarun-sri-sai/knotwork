package main

import (
	"fmt"

	"github.com/tarun-sri-sai/knotwork/internal/core/adapters/git"
	"github.com/tarun-sri-sai/knotwork/internal/core/ports"
)

type Core struct {
	repository ports.Repository
}

func NewCore(repoType string, repoDsn string) (*Core, error) {
	repositoryAdapters := map[string]func(string) (ports.Repository, error){
		"git": git.NewGitRepository,
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
