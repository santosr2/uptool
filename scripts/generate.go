//go:build tools

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

// Command generate generates code for integrations and guards.
//
// Usage:
//
//	go run scripts/generate.go [integrations|guards|all]
//
// If no argument is provided, generates both integrations and guards.
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	target := "all"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	var exitCode int

	if target == "integrations" || target == "all" {
		if err := runGenerator("gen_integrations.go", "integrations"); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating integrations: %v\n", err)
			exitCode = 1
		}
	}

	if target == "guards" || target == "all" {
		if err := runGenerator("gen_guards.go", "guards"); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating guards: %v\n", err)
			exitCode = 1
		}
	}

	if target != "integrations" && target != "guards" && target != "all" {
		fmt.Fprintf(os.Stderr, "Unknown target: %s (expected: integrations, guards, or all)\n", target)
		exitCode = 1
	}

	os.Exit(exitCode)
}

func runGenerator(script, name string) error {
	fmt.Printf("Generating %s...\n", name)
	cmd := exec.Command("go", "run", "scripts/"+script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
