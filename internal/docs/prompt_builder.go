package docs

// PromptBuilder constructs prompts and formats output using a Profile.
// A nil profile resolves to the default profile.
type PromptBuilder struct {
	profile Profile
}

// NewPromptBuilder creates a PromptBuilder for the given profile.
// If profile is nil, it resolves to the default profile.
func NewPromptBuilder(profile Profile) *PromptBuilder {
	if profile == nil {
		profile = DefaultProfile()
	}
	return &PromptBuilder{profile: profile}
}

// BuildGeneratePrompt returns the generation prompt for a file.
func (b *PromptBuilder) BuildGeneratePrompt(file string, content []byte) string {
	return b.profile.BuildPrompt(file, content)
}

// BuildUpdatePrompt returns the update prompt for a changed file.
func (b *PromptBuilder) BuildUpdatePrompt(file string, content []byte) string {
	return b.profile.BuildUpdatePrompt(file, content)
}

// FormatOutput applies the profile's output formatting.
func (b *PromptBuilder) FormatOutput(doc string) string {
	return b.profile.FormatOutput(doc)
}
