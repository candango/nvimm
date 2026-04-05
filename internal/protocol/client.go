// Package protocol provides implementations for communicating with the GitHub
// API to manage Neovim releases using the gopeasant transport.
package protocol

import (
	"errors"
	"net/http"

	peasant "github.com/candango/gopeasant"
)

// GithubDirectoryProvider is an in-memory implementation of DirectoryProvider.
// It points to github.
type GithubDirectoryProvider struct {
	url string
}

// Directory returns a static map containing the directory endpoints.
// It constructs the "releases" endpoint using the provider's URL.
func (p *GithubDirectoryProvider) Directory() (map[string]any, error) {
	return map[string]any{
		"releases": p.GetUrl() + "/releases",
	}, nil
}

// GetUrl returns the URL configured for the GitHub provider pointing to the
// Neovim repository.
func (p *GithubDirectoryProvider) GetUrl() string {
	return "https://api.github.com/repos/neovim/neovim"
}

// SetTransport is a no-op for GithubDirectoryProvider, as it does not use a
// transport. It always returns nil.
func (p *GithubDirectoryProvider) SetTransport(_ peasant.Transport) error {
	return nil
}

// GithubTransport extends the base peasant.HttpTransport to provide
// GitHub-specific operations.
type GithubTransport struct {
	*peasant.HttpTransport
}

// NewGithubTransport initializes a new GithubTransport using a
// GithubDirectoryProvider and a default HTTP transport.
func NewGithubTransport() (*GithubTransport, error) {
	p := &GithubDirectoryProvider{}
	ht, err := peasant.NewHttpTransport(p)
	if err != nil {
		return nil, err
	}
	return &GithubTransport{
		ht,
	}, nil
}

// GetReleases performs an HTTP GET request to the releases endpoint and
// returns the raw http.Response. It returns an error if the request fails or
// if the status code indicates a non-success result (greater than 299).
func (gt *GithubTransport) GetReleases() (*http.Response, error) {
	d, err := gt.Directory()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, d["releases"].(string), nil)
	if err != nil {
		return nil, err
	}

	res, err := gt.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 299 {
		return nil, errors.New(res.Status)
	}
	return res, nil
}

// GithubPeasant is the domain-level client for GitHub operations. It wraps
// peasant.Peasant and delegates HTTP operations to the underlying
// GithubTransport.
type GithubPeasant struct {
	*peasant.Peasant
}

// NewGithubPeasant initializes a production GithubPeasant backed by a real
// GithubTransport.
func NewGithubPeasant() (*GithubPeasant, error) {
	gt, err := NewGithubTransport()
	if err != nil {
		return nil, err
	}
	return &GithubPeasant{peasant.NewPeasant(gt)}, nil
}

// NewTestPeasant initializes a GithubPeasant from an injected peasant.Peasant,
// allowing tests to supply a mock transport.
func NewTestPeasant(p *peasant.Peasant) *GithubPeasant {
	return &GithubPeasant{p}
}

// GetReleases delegates to the underlying GithubTransport to fetch Neovim
// release metadata from the GitHub API.
func (p *GithubPeasant) GetReleases() (*http.Response, error) {
	return p.Transport.(*GithubTransport).GetReleases()
}
