// Copyright (c) 2024 santosr2
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package policy handles cadence-based update filtering.
package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/santosr2/uptool/internal/secureio"
)

// CadenceState tracks when manifests were last checked for updates.
type CadenceState struct {
	LastChecked map[string]time.Time `json:"last_checked"` // manifestPath -> timestamp
}

// ShouldCheckForUpdates determines if a manifest should be checked based on cadence policy.
func (cs *CadenceState) ShouldCheckForUpdates(manifestPath, cadence string) bool {
	if cadence == "" {
		return true // No cadence restriction
	}

	lastCheck, exists := cs.LastChecked[manifestPath]
	if !exists {
		return true // Never checked before
	}

	now := time.Now()
	switch cadence {
	case "daily":
		return now.Sub(lastCheck) >= 24*time.Hour
	case "weekly":
		return now.Sub(lastCheck) >= 7*24*time.Hour
	case "monthly":
		return now.Sub(lastCheck) >= 30*24*time.Hour
	default:
		return true // Unknown cadence, allow check
	}
}

// MarkChecked records that a manifest was checked at the current time.
func (cs *CadenceState) MarkChecked(manifestPath string) {
	if cs.LastChecked == nil {
		cs.LastChecked = make(map[string]time.Time)
	}
	cs.LastChecked[manifestPath] = time.Now()
}

// LoadCadenceState loads the cadence state from disk.
func LoadCadenceState(stateFile string) (*CadenceState, error) {
	data, err := secureio.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return empty state
			return &CadenceState{
				LastChecked: make(map[string]time.Time),
			}, nil
		}
		return nil, fmt.Errorf("read state file: %w", err)
	}

	var state CadenceState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse state file: %w", err)
	}

	if state.LastChecked == nil {
		state.LastChecked = make(map[string]time.Time)
	}

	return &state, nil
}

// SaveCadenceState writes the cadence state to disk.
func SaveCadenceState(stateFile string, state *CadenceState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(stateFile)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0o600); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}

	return nil
}

// GetDefaultStateFile returns the default location for the cadence state file.
func GetDefaultStateFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".uptool.state.json"
	}
	return filepath.Join(homeDir, ".config", "uptool", "state.json")
}
