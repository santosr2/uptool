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

// Package main implements an uptool plugin for Python requirements.txt dependencies.
// This plugin demonstrates the external plugin architecture for uptool.
package main

import "github.com/santosr2/uptool/internal/engine"

// RegisterWith is called by uptool to register this plugin's integrations.
// This function MUST be exported and have this exact signature for the plugin to work.
//
// uptool will call this function and pass its Register function, which the plugin
// should call to register each integration it provides.
func RegisterWith(register func(name string, constructor func() engine.Integration)) {
	// Register the Python integration
	register("python", New)
}

// main is not used when building as a plugin, but helps during development and testing.
// You can run tests using `go test` even though the main function isn't called in plugin mode.
func main() {
	// This is intentionally empty.
	// When built as a plugin (-buildmode=plugin), the main function is not executed.
	// The RegisterWith function is called by uptool instead.
}
