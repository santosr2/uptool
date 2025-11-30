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

package secureio

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid absolute path",
			path:    "/tmp/test.txt",
			wantErr: false,
		},
		{
			name:    "valid absolute path with subdirs",
			path:    "/home/user/projects/file.go",
			wantErr: false,
		},
		{
			name:    "relative path rejected",
			path:    "relative/path.txt",
			wantErr: true,
		},
		{
			name:    "directory traversal rejected",
			path:    "/tmp/../etc/passwd",
			wantErr: true,
		},
		{
			name:    "hidden directory traversal",
			path:    "/tmp/foo/../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "dot-dot in middle rejected",
			path:    "/tmp/test/../secret.txt",
			wantErr: true,
		},
		{
			name:    "current directory reference",
			path:    "./file.txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	t.Run("reads valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := []byte("test content")

		if err := os.WriteFile(testFile, content, 0o644); err != nil {
			t.Fatal(err)
		}

		got, err := ReadFile(testFile)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}

		if !bytes.Equal(got, content) {
			t.Errorf("ReadFile() = %q, want %q", got, content)
		}
	})

	t.Run("rejects relative path", func(t *testing.T) {
		_, err := ReadFile("relative/path.txt")
		if err == nil {
			t.Error("ReadFile() expected error for relative path")
		}
	})

	t.Run("rejects directory traversal", func(t *testing.T) {
		_, err := ReadFile("/tmp/../etc/passwd")
		if err == nil {
			t.Error("ReadFile() expected error for directory traversal")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := ReadFile("/tmp/nonexistent-file-12345.txt")
		if err == nil {
			t.Error("ReadFile() expected error for non-existent file")
		}
	})
}

func TestWriteFile(t *testing.T) {
	t.Run("writes valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := []byte("test content")

		err := WriteFile(testFile, content, 0o644)
		if err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("os.ReadFile() error = %v", err)
		}

		if !bytes.Equal(got, content) {
			t.Errorf("WriteFile() wrote %q, want %q", got, content)
		}
	})

	t.Run("rejects relative path", func(t *testing.T) {
		err := WriteFile("relative/path.txt", []byte("test"), 0o644)
		if err == nil {
			t.Error("WriteFile() expected error for relative path")
		}
	})

	t.Run("rejects directory traversal", func(t *testing.T) {
		err := WriteFile("/tmp/../etc/test.txt", []byte("test"), 0o644)
		if err == nil {
			t.Error("WriteFile() expected error for directory traversal")
		}
	})
}

func TestCreate(t *testing.T) {
	t.Run("creates valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")

		f, err := Create(testFile)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		defer f.Close()

		// Verify file was created
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Error("Create() did not create file")
		}
	})

	t.Run("rejects relative path", func(t *testing.T) {
		_, err := Create("relative/path.txt")
		if err == nil {
			t.Error("Create() expected error for relative path")
		}
	})

	t.Run("rejects directory traversal", func(t *testing.T) {
		_, err := Create("/tmp/../etc/test.txt")
		if err == nil {
			t.Error("Create() expected error for directory traversal")
		}
	})
}
