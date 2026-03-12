package types

// PreviewConfig represents the root structure of .previewctl/config.yml.
type PreviewConfig struct {
	Version  int                      `yaml:"version"`
	Preview  PreviewSettings          `yaml:"preview"`
	Services map[string]ServiceConfig `yaml:"services"`
}

type PreviewSettings struct {
	TTL string `yaml:"ttl"`
}

type ServiceConfig struct {
	Build     *BuildConfig      `yaml:"build,omitempty"`
	Image     string            `yaml:"image,omitempty"`
	Port      int               `yaml:"port,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	Volumes   []string          `yaml:"volumes,omitempty"`
	Seed      *SeedConfig       `yaml:"seed,omitempty"`
	DependsOn []string          `yaml:"depends_on,omitempty"`
}

// SeedConfig holds pre-start (filesystem) and post-start (runtime) seed entries.
type SeedConfig struct {
	Prestart  []SeedEntry `yaml:"prestart,omitempty"`
	Poststart []SeedEntry `yaml:"poststart,omitempty"`
}

// SeedEntry describes a file or directory to copy into a container,
// with an optional command to run after copying (poststart only).
type SeedEntry struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	Cmd         string `yaml:"cmd,omitempty"`
}

const (
	BuildTypeDockerfile = "dockerfile"
	BuildTypeNixpacks   = "nixpacks"
	BuildTypeRailpack   = "railpack"
)

type BuildConfig struct {
	Type       string `yaml:"type"`
	Context    string `yaml:"context"`
	Dockerfile string `yaml:"dockerfile,omitempty"`
}
