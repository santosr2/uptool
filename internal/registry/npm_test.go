package registry

import (
	"context"
	"testing"
	"time"
)

func TestNPMClient_GetLatestVersion(t *testing.T) {
	client := NewNPMClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test with a well-known package
	version, err := client.GetLatestVersion(ctx, "lodash")
	if err != nil {
		t.Fatalf("GetLatestVersion failed: %v", err)
	}

	if version == "" {
		t.Fatal("Expected non-empty version")
	}

	t.Logf("Latest lodash version: %s", version)
}

func TestNPMClient_GetPackageInfo(t *testing.T) {
	client := NewNPMClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info, err := client.GetPackageInfo(ctx, "express")
	if err != nil {
		t.Fatalf("GetPackageInfo failed: %v", err)
	}

	if info.Name != "express" {
		t.Errorf("Expected name 'express', got %q", info.Name)
	}

	if len(info.Versions) == 0 {
		t.Error("Expected non-zero versions")
	}

	t.Logf("Found %d versions of express", len(info.Versions))
}
