package docs

// DefaultProfileName is the profile used when no explicit profile is selected.
const DefaultProfileName = "software-documenter"

// Profile customizes prompt construction and final markdown formatting for a documentation use case.
type Profile interface {
	Name() string
	Description() string
	BuildPrompt(file string, content []byte) string
	BuildUpdatePrompt(file string, content []byte) string
	FormatOutput(doc string) string
}

// ProfileConfig is the serializable/user-facing profile selection shape.
type ProfileConfig struct {
	Name string `mapstructure:"profile" yaml:"profile" json:"profile"`
}
