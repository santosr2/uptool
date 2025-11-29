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

package policy

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCadenceState_ShouldCheckForUpdates(t *testing.T) {
	now := time.Now()
	tests := []struct {
		lastChecked time.Time
		name        string
		cadence     string
		description string
		want        bool
	}{
		{
			name:        "no cadence always allows",
			lastChecked: now.Add(-1 * time.Hour),
			cadence:     "",
			want:        true,
			description: "empty cadence should always allow checks",
		},
		{
			name:        "daily allows after 24h",
			lastChecked: now.Add(-25 * time.Hour),
			cadence:     "daily",
			want:        true,
			description: "daily cadence should allow after 24 hours",
		},
		{
			name:        "daily blocks before 24h",
			lastChecked: now.Add(-23 * time.Hour),
			cadence:     "daily",
			want:        false,
			description: "daily cadence should block before 24 hours",
		},
		{
			name:        "weekly allows after 7 days",
			lastChecked: now.Add(-8 * 24 * time.Hour),
			cadence:     "weekly",
			want:        true,
			description: "weekly cadence should allow after 7 days",
		},
		{
			name:        "weekly blocks before 7 days",
			lastChecked: now.Add(-6 * 24 * time.Hour),
			cadence:     "weekly",
			want:        false,
			description: "weekly cadence should block before 7 days",
		},
		{
			name:        "monthly allows after 30 days",
			lastChecked: now.Add(-31 * 24 * time.Hour),
			cadence:     "monthly",
			want:        true,
			description: "monthly cadence should allow after 30 days",
		},
		{
			name:        "monthly blocks before 30 days",
			lastChecked: now.Add(-29 * 24 * time.Hour),
			cadence:     "monthly",
			want:        false,
			description: "monthly cadence should block before 30 days",
		},
		{
			name:        "unknown cadence allows",
			lastChecked: now.Add(-1 * time.Hour),
			cadence:     "invalid",
			want:        true,
			description: "unknown cadence should allow checks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &CadenceState{
				LastChecked: map[string]time.Time{
					"test-manifest": tt.lastChecked,
				},
			}

			got := cs.ShouldCheckForUpdates("test-manifest", tt.cadence)
			if got != tt.want {
				t.Errorf("ShouldCheckForUpdates() = %v, want %v (%s)", got, tt.want, tt.description)
			}
		})
	}
}

func TestCadenceState_NeverChecked(t *testing.T) {
	cs := &CadenceState{
		LastChecked: make(map[string]time.Time),
	}

	// Should allow check for never-checked manifest
	if !cs.ShouldCheckForUpdates("never-checked", "daily") {
		t.Error("ShouldCheckForUpdates() for never-checked manifest = false, want true")
	}
}

func TestCadenceState_MarkChecked(t *testing.T) {
	cs := &CadenceState{}

	before := time.Now()
	cs.MarkChecked("test-manifest")
	after := time.Now()

	if cs.LastChecked == nil {
		t.Fatal("MarkChecked() did not initialize LastChecked map")
	}

	lastCheck, exists := cs.LastChecked["test-manifest"]
	if !exists {
		t.Fatal("MarkChecked() did not record check time")
	}

	if lastCheck.Before(before) || lastCheck.After(after) {
		t.Errorf("MarkChecked() time = %v, want between %v and %v", lastCheck, before, after)
	}
}

func TestLoadAndSaveCadenceState(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "test-state.json")

	// Create a state with some data
	originalState := &CadenceState{
		LastChecked: map[string]time.Time{
			"package.json": time.Now().Add(-24 * time.Hour),
			"Chart.yaml":   time.Now().Add(-48 * time.Hour),
		},
	}

	// Save the state
	err := SaveCadenceState(stateFile, originalState)
	if err != nil {
		t.Fatalf("SaveCadenceState() error = %v", err)
	}

	// Verify file exists
	if _, statErr := os.Stat(stateFile); os.IsNotExist(statErr) {
		t.Fatal("SaveCadenceState() did not create state file")
	}

	// Load the state back
	loadedState, err := LoadCadenceState(stateFile)
	if err != nil {
		t.Fatalf("LoadCadenceState() error = %v", err)
	}

	// Verify data matches
	if len(loadedState.LastChecked) != len(originalState.LastChecked) {
		t.Errorf("LoadCadenceState() loaded %d entries, want %d", len(loadedState.LastChecked), len(originalState.LastChecked))
	}

	for path, originalTime := range originalState.LastChecked {
		loadedTime, exists := loadedState.LastChecked[path]
		if !exists {
			t.Errorf("LoadCadenceState() missing entry for %s", path)
			continue
		}

		// Times should match (allowing for JSON marshaling precision)
		diff := originalTime.Sub(loadedTime)
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Second {
			t.Errorf("LoadCadenceState() time for %s differs by %v", path, diff)
		}
	}
}

func TestLoadCadenceState_NonExistent(t *testing.T) {
	// Loading non-existent file should return empty state, not error
	state, err := LoadCadenceState("/nonexistent/path/state.json")
	if err != nil {
		t.Fatalf("LoadCadenceState() for non-existent file error = %v, want nil", err)
	}

	if state == nil {
		t.Fatal("LoadCadenceState() returned nil state")
	}

	if state.LastChecked == nil {
		t.Error("LoadCadenceState() returned state with nil LastChecked map")
	}

	if len(state.LastChecked) != 0 {
		t.Errorf("LoadCadenceState() returned %d entries, want 0", len(state.LastChecked))
	}
}

func TestGetDefaultStateFile(t *testing.T) {
	path := GetDefaultStateFile()

	if path == "" {
		t.Error("GetDefaultStateFile() returned empty string")
	}

	// Should contain .config/uptool or be .uptool.state.json
	if !filepath.IsAbs(path) && path != ".uptool.state.json" {
		t.Errorf("GetDefaultStateFile() = %q, expected absolute path or .uptool.state.json", path)
	}
}
