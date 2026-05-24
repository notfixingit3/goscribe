package docs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const stateFile = ".goscribe-state"

// State tracks the last processed git commit for incremental updates.
type State struct {
	LastCommit string `json:"last_commit"`
}

// SaveCommitState persists the current commit hash to the state file.
func SaveCommitState(sourcePath, commit string) error {
	statePath := filepath.Join(sourcePath, stateFile)
	state := State{LastCommit: commit}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	return os.WriteFile(statePath, data, 0600)
}

// LoadCommitState reads the last saved commit hash from the state file.
func LoadCommitState(sourcePath string) (string, error) {
	statePath := filepath.Join(sourcePath, stateFile)
	data, err := os.ReadFile(statePath) // #nosec G304 -- statePath is constructed from sourcePath + constant filename
	if err != nil {
		return "", fmt.Errorf("read state: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err == nil && state.LastCommit != "" {
		return state.LastCommit, nil
	}

	return string(data), nil
}
