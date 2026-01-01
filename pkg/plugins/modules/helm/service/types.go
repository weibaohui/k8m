package helm

import (
	"time"
)

type IndexFile struct {
	// This is used ONLY for validation against chartmuseum's index files and is discarded after validation.
	ServerInfo map[string]any           `json:"serverInfo,omitempty"`
	APIVersion string                   `json:"apiVersion"`
	Generated  time.Time                `json:"generated"`
	Entries    map[string]ChartVersions `json:"entries"`
	PublicKeys []string                 `json:"publicKeys,omitempty"`

	// Annotations are additional mappings uninterpreted by Helm. They are made available for
	// other applications to add information to the index file.
	Annotations map[string]string `json:"annotations,omitempty"`
}
type ChartVersions []*ChartVersion
type ChartVersion struct {
	*Metadata
	URLs    []string  `json:"urls"`
	Created time.Time `json:"created,omitempty"`
	Removed bool      `json:"removed,omitempty"`
	Digest  string    `json:"digest,omitempty"`

	// ChecksumDeprecated is deprecated in Helm 3, and therefore ignored. Helm 3 replaced
	// this with Digest. However, with a strict YAML parser enabled, a field must be
	// present on the struct for backwards compatibility.
	ChecksumDeprecated string `json:"checksum,omitempty"`

	// EngineDeprecated is deprecated in Helm 3, and therefore ignored. However, with a strict
	// YAML parser enabled, this field must be present.
	EngineDeprecated string `json:"engine,omitempty"`

	// TillerVersionDeprecated is deprecated in Helm 3, and therefore ignored. However, with a strict
	// YAML parser enabled, this field must be present.
	TillerVersionDeprecated string `json:"tillerVersion,omitempty"`

	// URLDeprecated is deprecated in Helm 3, superseded by URLs. It is ignored. However,
	// with a strict YAML parser enabled, this must be present on the struct.
	URLDeprecated string `json:"url,omitempty"`
}

// Metadata for a Chart file. This models the structure of a Chart.yaml file.
type Metadata struct {
	// The name of the chart. Required.
	Name string `json:"name,omitempty"`
	// The URL to a relevant project page, git repo, or contact person
	Home string `json:"home,omitempty"`
	// Source is the URL to the source code of this chart
	Sources []string `json:"sources,omitempty"`
	// A SemVer 2 conformant version string of the chart. Required.
	Version string `json:"version,omitempty"`
	// A one-sentence description of the chart
	Description string `json:"description,omitempty"`
	// A list of string keywords
	Keywords []string `json:"keywords,omitempty"`
	// A list of name and URL/email address combinations for the maintainer(s)
	Maintainers []*Maintainer `json:"maintainers,omitempty"`
	// The URL to an icon file.
	Icon string `json:"icon,omitempty"`
	// The API Version of this chart. Required.
	APIVersion string `json:"apiVersion,omitempty"`
	// The condition to check to enable chart
	Condition string `json:"condition,omitempty"`
	// The tags to check to enable chart
	Tags string `json:"tags,omitempty"`
	// The version of the application enclosed inside of this chart.
	AppVersion string `json:"appVersion,omitempty"`
	// Whether or not this chart is deprecated
	Deprecated bool `json:"deprecated,omitempty"`
	// Annotations are additional mappings uninterpreted by Helm,
	// made available for inspection by other applications.
	Annotations map[string]string `json:"annotations,omitempty"`
	// KubeVersion is a SemVer constraint specifying the version of Kubernetes required.
	KubeVersion string `json:"kubeVersion,omitempty"`
	// Dependencies are a list of dependencies for a chart.
	Dependencies []*Dependency `json:"dependencies,omitempty"`
	// Specifies the chart type: application or library
	Type string `json:"type,omitempty"`
}

// Dependency describes a chart upon which another chart depends.
//
// Dependencies can be used to express developer intent, or to capture the state
// of a chart.
type Dependency struct {
	// Name is the name of the dependency.
	//
	// This must mach the name in the dependency's Chart.yaml.
	Name string `json:"name" yaml:"name"`
	// Version is the version (range) of this chart.
	//
	// A lock file will always produce a single version, while a dependency
	// may contain a semantic version range.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// The URL to the repository.
	//
	// Appending `index.yaml` to this string should result in a URL that can be
	// used to fetch the repository index.
	Repository string `json:"repository" yaml:"repository"`
	// A yaml path that resolves to a boolean, used for enabling/disabling charts (e.g. subchart1.enabled )
	Condition string `json:"condition,omitempty" yaml:"condition,omitempty"`
	// Tags can be used to group charts for enabling/disabling together
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	// Enabled bool determines if chart should be loaded
	Enabled bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	// ImportValues holds the mapping of source values to parent key to be imported. Each item can be a
	// string or pair of child/parent sublist items.
	ImportValues []any `json:"import-values,omitempty" yaml:"import-values,omitempty"`
	// Alias usable alias to be used for the chart
	Alias string `json:"alias,omitempty" yaml:"alias,omitempty"`
}

type Maintainer struct {
	// Name is a user name or organization name
	Name string `json:"name,omitempty"`
	// Email is an optional email address to contact the named maintainer
	Email string `json:"email,omitempty"`
	// URL is an optional URL to an address for the named maintainer
	URL string `json:"url,omitempty"`
}
type RepoVO struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
