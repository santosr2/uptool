package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

const githubAPIURL = "https://api.github.com"

// GitHubClient queries GitHub API for release information.
type GitHubClient struct {
	client  *http.Client
	baseURL string
	token   string
}

// NewGitHubClient creates a new GitHub API client.
// Token is optional but recommended to avoid rate limiting.
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: githubAPIURL,
		token:   token,
	}
}

// Release represents a GitHub release.
type Release struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	CreatedAt   string `json:"created_at"`
	PublishedAt string `json:"published_at"`
}

// GetLatestRelease fetches the latest non-prerelease release for a repository.
func (c *GitHubClient) GetLatestRelease(ctx context.Context, owner, repo string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.baseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("repository not found: %s/%s", owner, repo)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	// Strip 'v' prefix if present
	version := strings.TrimPrefix(release.TagName, "v")
	return version, nil
}

// GetAllReleases fetches all releases for a repository.
func (c *GitHubClient) GetAllReleases(ctx context.Context, owner, repo string) ([]Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases", c.baseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository not found: %s/%s", owner, repo)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var releases []Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return releases, nil
}

// FindBestRelease finds the best release matching a constraint.
func (c *GitHubClient) FindBestRelease(ctx context.Context, owner, repo, constraint string, allowPrerelease bool) (string, error) {
	releases, err := c.GetAllReleases(ctx, owner, repo)
	if err != nil {
		return "", err
	}

	// Parse constraint
	constraintObj, err := semver.NewConstraint(constraint)
	if err != nil {
		// If constraint parsing fails, return latest
		return c.GetLatestRelease(ctx, owner, repo)
	}

	// Collect matching versions
	var versions []*semver.Version
	for _, rel := range releases {
		if rel.Draft {
			continue
		}

		if rel.Prerelease && !allowPrerelease {
			continue
		}

		// Strip 'v' prefix
		versionStr := strings.TrimPrefix(rel.TagName, "v")

		parsed, err := semver.NewVersion(versionStr)
		if err != nil {
			continue
		}

		if constraintObj.Check(parsed) {
			versions = append(versions, parsed)
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no releases match constraint: %s", constraint)
	}

	// Find the highest version
	var best *semver.Version
	for _, v := range versions {
		if best == nil || v.GreaterThan(best) {
			best = v
		}
	}

	return best.Original(), nil
}

// ParseGitHubURL extracts owner and repo from a GitHub URL.
// Supports:
// - https://github.com/owner/repo
// - github.com/owner/repo
// - owner/repo
func ParseGitHubURL(url string) (owner, repo string, err error) {
	// Remove common prefixes
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "github.com/")

	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Match owner/repo pattern
	re := regexp.MustCompile(`^([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(url)

	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid GitHub URL: %s", url)
	}

	return matches[1], matches[2], nil
}
