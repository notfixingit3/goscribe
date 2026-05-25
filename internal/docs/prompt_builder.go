package docs

// PromptBuilder constructs prompts and formats output using a Profile
// and an optional Template. A nil profile resolves to the default profile.
// A nil template is an identity pass-through.
type PromptBuilder struct {
	profile  Profile
	template Template
}

// NewPromptBuilder creates a PromptBuilder for the given profile.
// If profile is nil, it resolves to the default profile.
func NewPromptBuilder(profile Profile) *PromptBuilder {
	if profile == nil {
		profile = DefaultProfile()
	}
	return &PromptBuilder{profile: profile}
}

// WithTemplate sets an optional Template on the PromptBuilder.
// Pass nil to remove a previously set template.
func (b *PromptBuilder) WithTemplate(t Template) *PromptBuilder {
	b.template = t
	return b
}

// BuildGeneratePrompt returns the generation prompt for a file.
// If a template is set, its style prompt is appended.
func (b *PromptBuilder) BuildGeneratePrompt(file string, content []byte) string {
	prompt := b.profile.BuildPrompt(file, content)
	if b.template != nil {
		prompt += "\n\n" + b.template.StylePrompt()
	}
	return prompt
}

// BuildUpdatePrompt returns the update prompt for a changed file.
// If a template is set, its style prompt is appended.
func (b *PromptBuilder) BuildUpdatePrompt(file string, content []byte) string {
	prompt := b.profile.BuildUpdatePrompt(file, content)
	if b.template != nil {
		prompt += "\n\n" + b.template.StylePrompt()
	}
	return prompt
}

// FormatOutput applies the profile's output formatting, then the
// template's FormatOutput if a template is set.
func (b *PromptBuilder) FormatOutput(doc string) string {
	out := b.profile.FormatOutput(doc)
	if b.template != nil {
		out = b.template.FormatOutput(out)
	}
	return out
}
