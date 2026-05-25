package docs

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Template defines the interface for documentation output templates.
// Templates modify the style and formatting of generated documentation
// by providing style prompts and post-processing formatting.
type Template interface {
	// Name returns the unique identifier for this template.
	Name() string
	// Description returns a human-readable description of the template.
	Description() string
	// StylePrompt returns a style instruction string composed with the profile prompt.
	StylePrompt() string
	// FormatOutput applies post-processing formatting to the documentation output.
	FormatOutput(doc string) string
}

// TemplateRegistry manages dynamically registered documentation templates.
// It is safe for concurrent use.
type TemplateRegistry struct {
	mu        sync.RWMutex
	templates map[string]Template
}

// defaultTemplateRegistry is the global template registry used by package helpers.
var defaultTemplateRegistry = NewTemplateRegistry()

// NewTemplateRegistry creates a new empty TemplateRegistry.
func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: make(map[string]Template),
	}
}

// Register adds a template under its trimmed name.
// It panics if the template is nil, has an empty name, or duplicates an existing name.
func (r *TemplateRegistry) Register(template Template) {
	if template == nil {
		panic("docs: cannot register nil template")
	}

	name := strings.TrimSpace(template.Name())
	if name == "" {
		panic("docs: template name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.templates[name]; exists {
		panic(fmt.Sprintf("docs: template %q already registered", name))
	}
	r.templates[name] = template
}

// Get returns the template for the given name, or nil if not found.
// An empty name returns nil, false (no template = no style modification).
func (r *TemplateRegistry) Get(name string) (Template, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	template, ok := r.templates[name]
	return template, ok
}

// List returns the names of all registered templates, sorted alphabetically.
func (r *TemplateRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.templates))
	for name := range r.templates {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisterTemplate adds a template to the default global registry.
func RegisterTemplate(template Template) {
	defaultTemplateRegistry.Register(template)
}

// GetTemplate returns the template for the given name from the default registry.
func GetTemplate(name string) (Template, bool) {
	return defaultTemplateRegistry.Get(name)
}

// RegisteredTemplates returns the names of all templates in the default registry.
func RegisteredTemplates() []string {
	return defaultTemplateRegistry.List()
}

// ResolveTemplate returns the named template, or nil when name is empty.
// An empty or whitespace-only name returns (nil, nil) with no error.
func ResolveTemplate(name string) (Template, error) {
	template, ok := GetTemplate(name)
	if ok {
		return template, nil
	}

	requested := strings.TrimSpace(name)
	if requested == "" {
		return nil, nil
	}

	return nil, fmt.Errorf("unknown template %q (available: %s)", requested, strings.Join(RegisteredTemplates(), ", "))
}
