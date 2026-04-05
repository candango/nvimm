package protocol

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	peasant "github.com/candango/gopeasant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testDirectoryProvider is a DirectoryProvider that points to a test server.
type testDirectoryProvider struct {
	baseURL string
}

func (p *testDirectoryProvider) Directory() (map[string]any, error) {
	return map[string]any{
		"releases": p.baseURL + "/releases",
	}, nil
}

func (p *testDirectoryProvider) GetUrl() string {
	return p.baseURL
}

func (p *testDirectoryProvider) SetTransport(_ peasant.Transport) error {
	return nil
}

// newTestGithubPeasant creates a GithubPeasant backed by a test HTTP server.
func newTestGithubPeasant(t *testing.T, baseURL string) *GithubPeasant {
	t.Helper()
	ht, err := peasant.NewHttpTransport(&testDirectoryProvider{baseURL: baseURL})
	require.NoError(t, err)
	gt := &GithubTransport{ht}
	return NewTestPeasant(peasant.NewPeasant(gt))
}

func TestGithubPeasant_GetReleases(t *testing.T) {
	releases := []map[string]any{
		{"tag_name": "stable", "name": "NVIM v0.11.5"},
		{"tag_name": "v0.11.5", "name": "NVIM v0.11.5"},
	}
	body, err := json.Marshal(releases)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/releases" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(body)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	gp := newTestGithubPeasant(t, srv.URL)

	res, err := gp.GetReleases()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var got []map[string]any
	err = json.NewDecoder(res.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "stable", got[0]["tag_name"])
}

func TestGithubPeasant_GetReleases_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	gp := newTestGithubPeasant(t, srv.URL)

	_, err := gp.GetReleases()
	assert.Error(t, err)
}

func TestNewGithubPeasant(t *testing.T) {
	gp, err := NewGithubPeasant()
	require.NoError(t, err)
	assert.NotNil(t, gp)
	assert.NotNil(t, gp.Peasant)
}
