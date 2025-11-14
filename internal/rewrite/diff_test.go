package rewrite

import (
	"strings"
	"testing"
)

func TestGenerateUnifiedDiff(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		oldContent  string
		newContent  string
		wantErr     bool
		wantContain []string
	}{
		{
			name:       "simple change",
			filename:   "test.txt",
			oldContent: "line 1\nline 2\nline 3\n",
			newContent: "line 1\nline 2 modified\nline 3\n",
			wantErr:    false,
			wantContain: []string{
				"test.txt",
				"-line 2",
				"+line 2 modified",
			},
		},
		{
			name:       "addition",
			filename:   "package.json",
			oldContent: "{\n  \"name\": \"test\"\n}",
			newContent: "{\n  \"name\": \"test\",\n  \"version\": \"1.0.0\"\n}",
			wantErr:    false,
			wantContain: []string{
				"package.json",
				"+",
			},
		},
		{
			name:       "deletion",
			filename:   "config.yaml",
			oldContent: "key1: value1\nkey2: value2\nkey3: value3",
			newContent: "key1: value1\nkey3: value3",
			wantErr:    false,
			wantContain: []string{
				"config.yaml",
				"-key2: value2",
			},
		},
		{
			name:       "no change",
			filename:   "unchanged.txt",
			oldContent: "same content\n",
			newContent: "same content\n",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateUnifiedDiff(tt.filename, tt.oldContent, tt.newContent)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateUnifiedDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("GenerateUnifiedDiff() output should contain %q, got:\n%s", want, got)
				}
			}
		})
	}
}

func TestGeneratePatch(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		oldContent  string
		newContent  string
		wantErr     bool
		wantContain []string
	}{
		{
			name:       "git-style patch",
			filename:   "README.md",
			oldContent: "# Title\nOld description\n",
			newContent: "# Title\nNew description\n",
			wantErr:    false,
			wantContain: []string{
				"a/README.md",
				"b/README.md",
				"-Old description",
				"+New description",
			},
		},
		{
			name:       "multiple changes",
			filename:   "config.yaml",
			oldContent: "version: 1.0.0\nname: old\nport: 8080",
			newContent: "version: 2.0.0\nname: new\nport: 8080",
			wantErr:    false,
			wantContain: []string{
				"a/config.yaml",
				"b/config.yaml",
				"-version: 1.0.0",
				"+version: 2.0.0",
				"-name: old",
				"+name: new",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GeneratePatch(tt.filename, tt.oldContent, tt.newContent)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("GeneratePatch() output should contain %q, got:\n%s", want, got)
				}
			}

			// Verify it has timestamp
			if !strings.Contains(got, "T") {
				t.Error("GeneratePatch() should include timestamps in RFC3339 format")
			}
		})
	}
}

func TestCountChanges(t *testing.T) {
	tests := []struct {
		name          string
		diff          string
		wantAdditions int
		wantDeletions int
	}{
		{
			name: "simple diff",
			diff: `--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
 line 1
-line 2
+line 2 modified
 line 3`,
			wantAdditions: 1,
			wantDeletions: 1,
		},
		{
			name: "multiple additions",
			diff: `--- a/file.txt
+++ b/file.txt
@@ -1,2 +1,5 @@
 line 1
 line 2
+line 3
+line 4
+line 5`,
			wantAdditions: 3,
			wantDeletions: 0,
		},
		{
			name: "multiple deletions",
			diff: `--- a/file.txt
+++ b/file.txt
@@ -1,5 +1,2 @@
 line 1
-line 2
-line 3
-line 4
 line 5`,
			wantAdditions: 0,
			wantDeletions: 3,
		},
		{
			name:          "no changes",
			diff:          `--- a/file.txt\n+++ b/file.txt\n`,
			wantAdditions: 0,
			wantDeletions: 0,
		},
		{
			name: "mixed changes",
			diff: `--- a/file.txt
+++ b/file.txt
@@ -1,6 +1,6 @@
 line 1
-old line 2
+new line 2
 line 3
-old line 4
-old line 5
+new line 4
+new line 5
+new line 6
 line 7`,
			wantAdditions: 4,
			wantDeletions: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdditions, gotDeletions := CountChanges(tt.diff)
			if gotAdditions != tt.wantAdditions {
				t.Errorf("CountChanges() additions = %v, want %v", gotAdditions, tt.wantAdditions)
			}
			if gotDeletions != tt.wantDeletions {
				t.Errorf("CountChanges() deletions = %v, want %v", gotDeletions, tt.wantDeletions)
			}
		})
	}
}
