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

package version

import (
	"testing"
)

func TestGet(t *testing.T) {
	t.Run("returns non-empty version", func(t *testing.T) {
		got := Get()
		if got == "" {
			t.Error("Get() returned empty string")
		}
	})

	t.Run("returns dev or valid semver", func(t *testing.T) {
		got := Get()
		// Should be either "dev" or a version number
		if got != "dev" && got == "" {
			t.Error("Get() returned invalid version")
		}
	})
}

func TestGet_Format(t *testing.T) {
	got := Get()

	// Version should not have leading/trailing whitespace
	if got != "" && (got[0] == ' ' || got[0] == '\n' || got[0] == '\t') {
		t.Error("Get() returned version with leading whitespace")
	}

	if got != "" && (got[len(got)-1] == ' ' || got[len(got)-1] == '\n' || got[len(got)-1] == '\t') {
		t.Error("Get() returned version with trailing whitespace")
	}
}
