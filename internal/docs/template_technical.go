package docs

// technicalTemplate produces precise, dense, code-heavy technical documentation.
type technicalTemplate struct{}

func (technicalTemplate) Name() string { return "technical" }
func (technicalTemplate) Description() string {
	return "Precise, dense, code-heavy technical documentation"
}
func (technicalTemplate) StylePrompt() string {
	return "Write documentation with technical precision and density. Focus on implementation details, API specifications, configuration parameters, and edge cases. Include extensive code examples, error handling patterns, and performance characteristics. Use concise, factual language. Prioritize completeness over readability. Include version compatibility notes and deprecation warnings where relevant."
}
func (technicalTemplate) FormatOutput(doc string) string { return doc }

func init() {
	RegisterTemplate(technicalTemplate{})
}
