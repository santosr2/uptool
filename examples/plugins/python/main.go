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
