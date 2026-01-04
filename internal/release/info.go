package release

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Releases []Info

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
func (rs *Releases) Process(data []byte) error {
	err := json.Unmarshal(data, &rs)
	if err != nil {
		return fmt.Errorf("failed to unmarshal releases: %w", err)
	}
	releases := *rs

	var stable Info
	for i, info := range releases {
		if info.TagName == "stable" {
			stable = info
			releases = append(releases[:i], releases[i+1:]...)
		}
	}
	for i, info := range releases {
		if info.Name == stable.Name {
			info.Stable = true
			releases[i] = info
		}
	}
	rs = &releases
	return nil
}

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

func (i *Info) CleanTagName() string {
	if i.TagName == "nightly" {
		return i.TagName
	}
	return strings.ReplaceAll(i.TagName, "v", "")
}

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
