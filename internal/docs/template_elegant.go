package docs

// elegantTemplate produces sophisticated, polished, professional documentation.
type elegantTemplate struct{}

func (elegantTemplate) Name() string { return "elegant" }

func (elegantTemplate) Description() string {
	return "Sophisticated, polished, professional tone"
}

func (elegantTemplate) StylePrompt() string {
	return "Write documentation with a sophisticated and polished tone. Use refined language, smooth transitions, and professional phrasing. Organize content with clear hierarchy, elegant section headings, and well-crafted introductions and conclusions. Maintain a warm but authoritative voice. Use precise vocabulary and avoid jargon where possible. Include carefully chosen examples that illustrate best practices."
}

func (elegantTemplate) FormatOutput(doc string) string { return doc }

func init() {
	RegisterTemplate(elegantTemplate{})
}
