package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repository struct {
	repo *git.Repository
}

func OpenRepo(path string) (*Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}
	return &Repository{repo: repo}, nil
}

func (r *Repository) GetCurrentCommit() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("get head: %w", err)
	}
	return head.Hash().String(), nil
}

func (r *Repository) GetChangedFiles(fromCommit, toCommit string) ([]string, error) {
	fromHash, err := r.repo.ResolveRevision(plumbing.Revision(fromCommit))
	if err != nil {
		return nil, fmt.Errorf("resolve from commit: %w", err)
	}

	toHash, err := r.repo.ResolveRevision(plumbing.Revision(toCommit))
	if err != nil {
		return nil, fmt.Errorf("resolve to commit: %w", err)
	}

	fromObj, err := r.repo.CommitObject(*fromHash)
	if err != nil {
		return nil, fmt.Errorf("get from commit: %w", err)
	}

	toObj, err := r.repo.CommitObject(*toHash)
	if err != nil {
		return nil, fmt.Errorf("get to commit: %w", err)
	}

	fromTree, err := fromObj.Tree()
	if err != nil {
		return nil, fmt.Errorf("get from tree: %w", err)
	}

	toTree, err := toObj.Tree()
	if err != nil {
		return nil, fmt.Errorf("get to tree: %w", err)
	}

	changes, err := object.DiffTree(fromTree, toTree)
	if err != nil {
		return nil, fmt.Errorf("diff trees: %w", err)
	}

	var files []string
	for _, change := range changes {
		if change.From.Name != "" {
			files = append(files, change.From.Name)
		}
		if change.To.Name != "" && change.To.Name != change.From.Name {
			files = append(files, change.To.Name)
		}
	}

	return files, nil
}
