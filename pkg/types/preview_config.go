package types

// PreviewConfig represents the root structure of .previewctrl/config.yml.
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
	DependsOn []string          `yaml:"depends_on,omitempty"`
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
