package docs

// friendlyTemplate produces welcoming, approachable documentation.
type friendlyTemplate struct{}

func (friendlyTemplate) Name() string {
	return "friendly"
}

func (friendlyTemplate) Description() string {
	return "Conversational, welcoming, beginner-friendly"
}

func (friendlyTemplate) StylePrompt() string {
	return "Write documentation in a warm, conversational tone that welcomes " +
		"readers of all skill levels. Use encouraging language, avoid " +
		"intimidating jargon, and gently explain concepts. Include practical " +
		"examples with step-by-step guidance. Use 'you' and 'we' pronouns " +
		"to create connection. Anticipate common mistakes and address them " +
		"proactively. Make the reader feel supported and capable."
}

func (friendlyTemplate) FormatOutput(doc string) string {
	return doc
}

func init() {
	RegisterTemplate(friendlyTemplate{})
}
