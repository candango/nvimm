package release

import (
	"testing"

	"github.com/candango/nvimim/internal/config"
	"github.com/stretchr/testify/require"
)

func TestProcessResolvesStableAliasByConcreteVersion(t *testing.T) {
	data := []byte(`[
		{
			"tag_name": "stable",
			"name": "Nvim stable",
			"body": "This release points to NVIM v0.12.4."
		},
		{
			"tag_name": "v0.12.4",
			"name": "Nvim v0.12.4"
		},
		{
			"tag_name": "v0.12.3",
			"name": "Nvim v0.12.3"
		}
	]`)

	releases := Releases{}
	err := releases.Process(data, &config.AppOptions{MinRelease: "0.7.0"})
	require.NoError(t, err)

	stable, err := releases.Get("stable")
	require.NoError(t, err)
	require.Equal(t, "0.12.4", stable.CleanTagName())
}

func TestProcessResolvesStableAliasFromNameWhenBodyHasNoVersion(t *testing.T) {
	data := []byte(`[
		{
			"tag_name": "stable",
			"name": "Nvim v0.12.4",
			"body": "Stable release"
		},
		{
			"tag_name": "v0.12.4",
			"name": "Nvim 0.12.4"
		}
	]`)

	releases := Releases{}
	err := releases.Process(data, &config.AppOptions{MinRelease: "0.7.0"})
	require.NoError(t, err)

	stable, err := releases.Get("stable")
	require.NoError(t, err)
	require.Equal(t, "0.12.4", stable.CleanTagName())
}
