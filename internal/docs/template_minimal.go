package docs

// minimalTemplate produces bare-bones, fact-only documentation.
type minimalTemplate struct{}

func (minimalTemplate) Name() string {
	return "minimal"
}

func (minimalTemplate) Description() string {
	return "Bare bones, just the facts, no embellishment"
}

func (minimalTemplate) StylePrompt() string {
	return "Write documentation with extreme brevity. Include only essential " +
		"information. Use terse sentences, bullet points, and minimal " +
		"explanation. Eliminate all filler, pleasantries, and decorative " +
		"language. Structure as headings with brief facts underneath. No " +
		"examples unless absolutely critical. Prioritize scannability and terseness."
}

func (minimalTemplate) FormatOutput(doc string) string {
	return doc
}

func init() {
	RegisterTemplate(minimalTemplate{})
}
