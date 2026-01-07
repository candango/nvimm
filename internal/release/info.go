package release

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/candango/iook/pathx"
	"github.com/candango/nvimm/internal/config"
)

// Releases represents a list of GitHub release information.
type Releases []Info

// Get retrieves the Info for a specific release. It supports the special
// identifier "stable" to fetch the stable release. If the release does not
// exist, it returns an error.
func (rs *Releases) Get(release string) (*Info, error) {
	releases := *rs

	for _, info := range releases {
		if release == "stable" && info.Stable == true {
			return &info, nil
		}
		if info.CleanTagName() == release {
			return &info, nil
		}
	}
	return nil, fmt.Errorf("release %s does not exists", release)
}

// Installed returns a list of releases that are present in the specified path.
func (rs *Releases) Installed(path string) []Info {
	releases := *rs
	installed := []Info{}
	for _, info := range releases {
		if pathx.Exists(filepath.Join(path, info.CleanTagName())) {
			installed = append(installed, info)
		}
	}
	return installed
}

// Available returns a list of releases that are not present in the installed
// releases.
func (rs *Releases) Available(installed []Info) []Info {
	releases := *rs
	installedDict := map[string]bool{}
	for _, info := range installed {
		installedDict[info.CleanTagName()] = true
	}
	available := []Info{}
	for _, info := range releases {
		ok := installedDict[info.CleanTagName()]
		if !ok {
			available = append(available, info)
		}
	}
	return available
}

// Process unmarshals the provided JSON data into the Releases struct. It also
// identifies the stable release and marks the corresponding Info entries
// accordingly.
func (rs *Releases) Process(data []byte, appOpts *config.AppOptions) error {
	err := json.Unmarshal(data, &rs)
	if err != nil {
		return fmt.Errorf("failed to unmarshal releases: %w", err)
	}

	releases := (*rs)[:0]
	var stable Info
	for _, info := range *rs {
		if info.TagName == "stable" {
			stable = info
			continue
		}

		if info.VersionLess(appOpts.MinRelease) {
			continue
		}

		if info.VersionLess("0.11.3") {
			checksums := info.ChecksumsFromBody()
			for i, asset := range info.Assets {
				digest, ok := checksums[asset.Name]
				if ok {
					asset.Digest = fmt.Sprintf("sha256:%s", digest)
					info.Assets[i] = asset
				}
			}
		}
		releases = append(releases, info)
	}

	for i, info := range releases {
		if info.Name == stable.Name {
			info.Stable = true
			releases[i] = info
		}
	}
	*rs = releases
	return nil
}

// Asset represents a GitHub release asset.
type Asset struct {
	Id            float64   `json:"id"`
	NodeId        string    `json:"node_id"`
	Name          string    `json:"name"`
	Label         string    `json:"label"`
	State         string    `json:"state"`
	ContentType   string    `json:"content_type"`
	Size          float64   `json:"size"`
	Digest        string    `json:"digest"`
	DownloadCount float64   `json:"download_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Uploader      User      `json:"uploader"`
}

// Info represents a GitHub release information.
type Info struct {
	Id              float64   `json:"id"`
	Author          User      `json:"author"`
	Assets          []Asset   `json:"assets"`
	Body            string    `json:"body"`
	Name            string    `json:"name"`
	TarballUrl      string    `json:"tarball_url"`
	ZipballUrl      string    `json:"zipball_url"`
	Url             string    `json:"url"`
	NodeId          string    `json:"node_id"`
	Immutable       bool      `json:"immutable"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PublishedAt     time.Time `json:"published_at"`
	TagName         string    `json:"tag_name"`
	Draft           bool      `json:"draft"`
	AssetsUrl       string    `json:"assets_url"`
	UploadUrl       string    `json:"upload_url"`
	HtmlUrl         string    `json:"html_url"`
	TargetCommitish string    `json:"target_commitish"`
	Stable          bool
	// Reactions   []Reaction `json:"reactions"`
}

// CleanTagName returns the tag name without the "v" prefix. If the tag name is
// "nightly", it is returned as is.
func (i *Info) CleanTagName() string {
	if i.TagName == "nightly" {
		return i.TagName
	}
	return strings.ReplaceAll(i.TagName, "v", "")
}

// VersionLess compares two version strings in "major.minor.patch" format.
// Returns true if the target version (i.CleanTagName()) is less than the
// reference version (v).
//
// Example: versionLess("0.11.3") returns true when i.CleanTagName() is
// "0.11.2".
func (i *Info) VersionLess(v string) bool {
	if i.CleanTagName() == "nightly" || i.TagName == "stable" {
		return false
	}
	s1 := strings.Split(i.CleanTagName(), ".")
	s2 := strings.Split(v, ".")
	for i := 0; i < len(s1) && i < len(s2); i++ {
		n1, _ := strconv.Atoi(s1[i])
		n2, _ := strconv.Atoi(s2[i])
		if n1 < n2 {
			return true
		} else if n1 > n2 {
			return false
		}
	}
	return len(s1) < len(s2)
}

var checksumRe = regexp.MustCompile(`([a-f0-9]{64})\s+([^\s]+)`)

func (i *Info) ChecksumsFromBody() map[string]string {
	result := make(map[string]string)
	matches := checksumRe.FindAllStringSubmatch(i.Body, -1)
	for _, m := range matches {
		result[m[2]] = m[1]
	}
	return result
}

// Reaction represents a GitHub reaction to a release.
type Reaction struct {
	Url        string `json:"url"`
	Confused   string `json:"confused"`
	Eyes       string `json:"eyes"`
	Heart      string `json:"heart"`
	Hooray     string `json:"hooray"`
	Laugh      string `json:"laugh"`
	MinusOne   string `json:"-1"`
	PlusOne    string `json:"+1"`
	Rocket     string `json:"rocket"`
	TotalCount string `json:"total_count"`
}

// User represents a GitHub user.
type User struct {
	Id                float64 `json:"id"`
	NodeId            string  `json:"node_id"`
	AvatarUrl         string  `json:"avatar_url"`
	GravatarId        string  `json:"gravatar_id"`
	Url               string  `json:"url"`
	HtmlUrl           string  `json:"html_url"`
	FollowersUrl      string  `json:"followers_url"`
	FollowingUrl      string  `json:"following_url"`
	GistsUrl          string  `json:"gists_url"`
	SubscriptionsUrl  string  `json:"subscriptions_url"`
	OrganizationsUrl  string  `json:"organizations_url"`
	ReposUrl          string  `json:"repos_url"`
	EventsUrl         string  `json:"events_url"`
	RecievedEventsUrl string  `json:"recieved_events_url"`
	Type              string  `json:"type"`
	SiteAdmin         bool    `json:"site_admin"`
}
