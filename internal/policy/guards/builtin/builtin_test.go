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

package builtin

import (
	"testing"

	"github.com/santosr2/uptool/internal/policy/guards"
)

func TestCIGreenGuard_Name(t *testing.T) {
	g := &CIGreenGuard{}

	if got := g.Name(); got != "ci-green" {
		t.Errorf("Name() = %q, want %q", got, "ci-green")
	}
}

func TestCIGreenGuard_Description(t *testing.T) {
	g := &CIGreenGuard{}

	desc := g.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestCodeownersApproveGuard_Name(t *testing.T) {
	g := &CodeownersApproveGuard{}

	if got := g.Name(); got != "codeowners-approve" {
		t.Errorf("Name() = %q, want %q", got, "codeowners-approve")
	}
}

func TestCodeownersApproveGuard_Description(t *testing.T) {
	g := &CodeownersApproveGuard{}

	desc := g.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestSecurityScanGuard_Name(t *testing.T) {
	g := &SecurityScanGuard{}

	if got := g.Name(); got != "security-scan" {
		t.Errorf("Name() = %q, want %q", got, "security-scan")
	}
}

func TestSecurityScanGuard_Description(t *testing.T) {
	g := &SecurityScanGuard{}

	desc := g.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestBuiltinGuards_RegisteredInGlobalRegistry(t *testing.T) {
	// Verify all builtin guards are registered in the global registry
	expectedGuards := []string{"ci-green", "codeowners-approve", "security-scan"}

	for _, name := range expectedGuards {
		guard, exists := guards.Get(name)
		if !exists {
			t.Errorf("Guard %q not found in global registry", name)
			continue
		}

		if guard.Name() != name {
			t.Errorf("Guard %q has Name() = %q", name, guard.Name())
		}

		if guard.Description() == "" {
			t.Errorf("Guard %q has empty description", name)
		}
	}
}

func TestBuiltinGuards_InList(t *testing.T) {
	// Verify all builtin guards appear in List()
	guardList := guards.List()
	guardMap := make(map[string]bool)
	for _, name := range guardList {
		guardMap[name] = true
	}

	expectedGuards := []string{"ci-green", "codeowners-approve", "security-scan"}
	for _, name := range expectedGuards {
		if !guardMap[name] {
			t.Errorf("Guard %q not found in List()", name)
		}
	}
}
