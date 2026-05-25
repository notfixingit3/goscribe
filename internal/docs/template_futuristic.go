package docs

// futuristicTemplate produces modern, forward-looking documentation.
type futuristicTemplate struct{}

func (futuristicTemplate) Name() string {
	return "futuristic"
}

func (futuristicTemplate) Description() string {
	return "Modern, forward-looking, innovative tone"
}

func (futuristicTemplate) StylePrompt() string {
	return "Write documentation with a forward-looking, innovative tone. " +
		"Emphasize modern practices, emerging patterns, and progressive " +
		"architecture. Use contemporary language and structure. Highlight " +
		"innovation, scalability, and future-proof design decisions. Include " +
		"forward-looking examples that anticipate next-generation use cases. " +
		"Make the technology feel cutting-edge and exciting."
}

func (futuristicTemplate) FormatOutput(doc string) string {
	return doc
}

func init() {
	RegisterTemplate(futuristicTemplate{})
}
