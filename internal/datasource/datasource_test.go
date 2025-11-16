package datasource

import (
	"context"
	"strings"
	"sync"
	"testing"
)

// mockDatasource implements Datasource for testing
type mockDatasource struct {
	name              string
	latestVersion     string
	versions          []string
	packageInfo       *PackageInfo
	getLatestErr      error
	getVersionsErr    error
	getPackageInfoErr error
}

func (m *mockDatasource) Name() string {
	return m.name
}

func (m *mockDatasource) GetLatestVersion(ctx context.Context, pkg string) (string, error) {
	if m.getLatestErr != nil {
		return "", m.getLatestErr
	}
	return m.latestVersion, nil
}

func (m *mockDatasource) GetVersions(ctx context.Context, pkg string) ([]string, error) {
	if m.getVersionsErr != nil {
		return nil, m.getVersionsErr
	}
	return m.versions, nil
}

func (m *mockDatasource) GetPackageInfo(ctx context.Context, pkg string) (*PackageInfo, error) {
	if m.getPackageInfoErr != nil {
		return nil, m.getPackageInfoErr
	}
	return m.packageInfo, nil
}

// resetRegistry clears the global datasource registry for testing
func resetRegistry() {
	mu.Lock()
	defer mu.Unlock()
	datasources = make(map[string]Datasource)
}

func TestRegister(t *testing.T) {
	resetRegistry()

	t.Run("registers datasource successfully", func(t *testing.T) {
		ds := &mockDatasource{name: "test"}
		Register(ds)

		retrieved, err := Get("test")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if retrieved != ds {
			t.Error("Get() returned different datasource")
		}
	})

	t.Run("panics on duplicate registration", func(t *testing.T) {
		resetRegistry()

		ds1 := &mockDatasource{name: "duplicate"}
		Register(ds1)

		defer func() {
			if r := recover(); r == nil {
				t.Error("Register() did not panic on duplicate")
			}
		}()

		ds2 := &mockDatasource{name: "duplicate"}
		Register(ds2)
	})
}

func TestGet(t *testing.T) {
	resetRegistry()

	t.Run("returns registered datasource", func(t *testing.T) {
		ds := &mockDatasource{name: "npm"}
		Register(ds)

		retrieved, err := Get("npm")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if retrieved != ds {
			t.Error("Get() returned wrong datasource")
		}
		if retrieved.Name() != "npm" {
			t.Errorf("Get() name = %q, want %q", retrieved.Name(), "npm")
		}
	})

	t.Run("returns error for non-existent datasource", func(t *testing.T) {
		resetRegistry()

		_, err := Get("nonexistent")
		if err == nil {
			t.Fatal("Get() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Get() error = %q, want error containing 'not found'", err.Error())
		}
	})
}

func TestList(t *testing.T) {
	resetRegistry()

	t.Run("returns empty list for no datasources", func(t *testing.T) {
		names := List()
		if len(names) != 0 {
			t.Errorf("List() count = %d, want 0", len(names))
		}
	})

	t.Run("returns all registered datasource names", func(t *testing.T) {
		resetRegistry()

		Register(&mockDatasource{name: "npm"})
		Register(&mockDatasource{name: "helm"})
		Register(&mockDatasource{name: "terraform"})

		names := List()
		if len(names) != 3 {
			t.Errorf("List() count = %d, want 3", len(names))
		}

		// Check all names are present
		nameMap := make(map[string]bool)
		for _, name := range names {
			nameMap[name] = true
		}

		if !nameMap["npm"] || !nameMap["helm"] || !nameMap["terraform"] {
			t.Errorf("List() names = %v, missing expected names", names)
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	resetRegistry()

	// Register some datasources
	Register(&mockDatasource{name: "npm"})
	Register(&mockDatasource{name: "helm"})

	// Test concurrent reads
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = Get("npm")
			_ = List()
		}()
	}

	wg.Wait()
}

func TestNPMDatasource(t *testing.T) {
	t.Run("returns correct name", func(t *testing.T) {
		ds := NewNPMDatasource()
		if ds.Name() != "npm" {
			t.Errorf("Name() = %q, want %q", ds.Name(), "npm")
		}
	})

	t.Run("GetLatestVersion calls client", func(t *testing.T) {
		ds := NewNPMDatasource()
		// Note: This will make a real HTTP call unless the registry client is mocked
		// For now, we're testing that it doesn't panic and has the right signature
		ctx := context.Background()
		_, _ = ds.GetLatestVersion(ctx, "test-package")
	})

	t.Run("GetVersions calls client", func(t *testing.T) {
		ds := NewNPMDatasource()
		ctx := context.Background()
		_, _ = ds.GetVersions(ctx, "test-package")
	})

	t.Run("GetPackageInfo calls client", func(t *testing.T) {
		ds := NewNPMDatasource()
		ctx := context.Background()
		_, _ = ds.GetPackageInfo(ctx, "test-package")
	})
}

func TestHelmDatasource(t *testing.T) {
	t.Run("returns correct name", func(t *testing.T) {
		ds := NewHelmDatasource()
		if ds.Name() != "helm" {
			t.Errorf("Name() = %q, want %q", ds.Name(), "helm")
		}
	})

	t.Run("GetLatestVersion handles invalid format", func(t *testing.T) {
		ds := NewHelmDatasource()
		ctx := context.Background()

		// Invalid format (missing |)
		version, err := ds.GetLatestVersion(ctx, "invalid-format")
		if err != nil {
			t.Errorf("GetLatestVersion() error = %v, want nil for invalid format", err)
		}
		if version != "" {
			t.Errorf("GetLatestVersion() = %q, want empty string for invalid format", version)
		}
	})

	t.Run("GetVersions handles invalid format", func(t *testing.T) {
		ds := NewHelmDatasource()
		ctx := context.Background()

		// Invalid format (missing |)
		versions, err := ds.GetVersions(ctx, "invalid-format")
		if err != nil {
			t.Errorf("GetVersions() error = %v, want nil for invalid format", err)
		}
		if versions != nil {
			t.Errorf("GetVersions() = %v, want nil for invalid format", versions)
		}
	})

	t.Run("GetPackageInfo handles invalid format", func(t *testing.T) {
		ds := NewHelmDatasource()
		ctx := context.Background()

		// Invalid format (missing |)
		info, err := ds.GetPackageInfo(ctx, "invalid-format")
		if err != nil {
			t.Errorf("GetPackageInfo() error = %v, want nil for invalid format", err)
		}
		if info != nil {
			t.Errorf("GetPackageInfo() = %v, want nil for invalid format", info)
		}
	})

	t.Run("GetLatestVersion calls client with valid format", func(t *testing.T) {
		ds := NewHelmDatasource()
		ctx := context.Background()
		_, _ = ds.GetLatestVersion(ctx, "https://charts.bitnami.com/bitnami|nginx")
	})
}

func TestTerraformDatasource(t *testing.T) {
	t.Run("returns correct name", func(t *testing.T) {
		ds := NewTerraformDatasource()
		if ds.Name() != "terraform" {
			t.Errorf("Name() = %q, want %q", ds.Name(), "terraform")
		}
	})

	t.Run("GetLatestVersion calls client", func(t *testing.T) {
		ds := NewTerraformDatasource()
		ctx := context.Background()
		_, _ = ds.GetLatestVersion(ctx, "hashicorp/consul/aws")
	})

	t.Run("GetVersions calls client", func(t *testing.T) {
		ds := NewTerraformDatasource()
		ctx := context.Background()
		_, _ = ds.GetVersions(ctx, "hashicorp/consul/aws")
	})

	t.Run("GetPackageInfo calls client", func(t *testing.T) {
		ds := NewTerraformDatasource()
		ctx := context.Background()
		_, _ = ds.GetPackageInfo(ctx, "hashicorp/consul/aws")
	})
}

func TestGitHubDatasource(t *testing.T) {
	t.Run("returns correct name", func(t *testing.T) {
		ds := NewGitHubDatasource()
		if ds.Name() != "github-releases" {
			t.Errorf("Name() = %q, want %q", ds.Name(), "github-releases")
		}
	})

	t.Run("GetLatestVersion handles invalid format", func(t *testing.T) {
		ds := NewGitHubDatasource()
		ctx := context.Background()

		// Invalid format (missing /)
		version, err := ds.GetLatestVersion(ctx, "invalid-format")
		if err != nil {
			t.Errorf("GetLatestVersion() error = %v, want nil for invalid format", err)
		}
		if version != "" {
			t.Errorf("GetLatestVersion() = %q, want empty string for invalid format", version)
		}
	})

	t.Run("GetLatestVersion handles too many parts", func(t *testing.T) {
		ds := NewGitHubDatasource()
		ctx := context.Background()

		// Invalid format (too many /)
		version, err := ds.GetLatestVersion(ctx, "owner/repo/extra")
		if err != nil {
			t.Errorf("GetLatestVersion() error = %v, want nil for invalid format", err)
		}
		if version != "" {
			t.Errorf("GetLatestVersion() = %q, want empty string for invalid format", version)
		}
	})

	t.Run("GetVersions handles invalid format", func(t *testing.T) {
		ds := NewGitHubDatasource()
		ctx := context.Background()

		// Invalid format (missing /)
		versions, err := ds.GetVersions(ctx, "invalid-format")
		if err != nil {
			t.Errorf("GetVersions() error = %v, want nil for invalid format", err)
		}
		if versions != nil {
			t.Errorf("GetVersions() = %v, want nil for invalid format", versions)
		}
	})

	t.Run("GetPackageInfo handles invalid format", func(t *testing.T) {
		ds := NewGitHubDatasource()
		ctx := context.Background()

		// Invalid format (missing /)
		info, err := ds.GetPackageInfo(ctx, "invalid-format")
		if err != nil {
			t.Errorf("GetPackageInfo() error = %v, want nil for invalid format", err)
		}
		if info != nil {
			t.Errorf("GetPackageInfo() = %v, want nil for invalid format", info)
		}
	})

	t.Run("GetLatestVersion calls client with valid format", func(t *testing.T) {
		ds := NewGitHubDatasource()
		ctx := context.Background()
		_, _ = ds.GetLatestVersion(ctx, "hashicorp/terraform")
	})

	t.Run("GetPackageInfo builds correct repository URL", func(t *testing.T) {
		ds := NewGitHubDatasource()
		ctx := context.Background()

		// Test with a known repository (may fail without network, but tests the logic)
		info, err := ds.GetPackageInfo(ctx, "owner/repo")
		// Even if it fails to fetch, check error handling is correct
		if err == nil && info != nil {
			expectedURL := "https://github.com/owner/repo"
			if info.Repository != expectedURL {
				t.Errorf("GetPackageInfo() Repository = %q, want %q", info.Repository, expectedURL)
			}
			if info.Homepage != expectedURL {
				t.Errorf("GetPackageInfo() Homepage = %q, want %q", info.Homepage, expectedURL)
			}
		}
	})
}

func TestDatasourceInterface(t *testing.T) {
	// Verify all datasources implement the interface
	var _ Datasource = &NPMDatasource{}
	var _ Datasource = &HelmDatasource{}
	var _ Datasource = &TerraformDatasource{}
	var _ Datasource = &GitHubDatasource{}
	var _ Datasource = &mockDatasource{}
}

func TestPackageInfoStruct(t *testing.T) {
	t.Run("creates PackageInfo with all fields", func(t *testing.T) {
		info := &PackageInfo{
			Name:        "test-package",
			Description: "A test package",
			Homepage:    "https://example.com",
			Repository:  "https://github.com/test/package",
			Versions: []VersionInfo{
				{
					Version:      "1.0.0",
					PublishedAt:  "2024-01-01",
					IsPrerelease: false,
					Deprecated:   false,
				},
				{
					Version:      "2.0.0-beta.1",
					PublishedAt:  "2024-06-01",
					IsPrerelease: true,
					Deprecated:   false,
				},
			},
		}

		if info.Name != "test-package" {
			t.Errorf("Name = %q, want %q", info.Name, "test-package")
		}
		if len(info.Versions) != 2 {
			t.Errorf("Versions count = %d, want 2", len(info.Versions))
		}
		if !info.Versions[1].IsPrerelease {
			t.Error("Second version should be marked as prerelease")
		}
	})
}

func TestVersionInfoStruct(t *testing.T) {
	t.Run("creates VersionInfo with all fields", func(t *testing.T) {
		v := VersionInfo{
			Version:      "1.2.3",
			PublishedAt:  "2024-01-15T10:30:00Z",
			IsPrerelease: false,
			Deprecated:   true,
		}

		if v.Version != "1.2.3" {
			t.Errorf("Version = %q, want %q", v.Version, "1.2.3")
		}
		if !v.Deprecated {
			t.Error("Should be marked as deprecated")
		}
		if v.IsPrerelease {
			t.Error("Should not be marked as prerelease")
		}
	})
}
