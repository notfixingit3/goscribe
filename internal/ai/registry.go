package ai

import (
	"fmt"
	"sort"
	"sync"

	"github.com/house/goscribe/pkg/providers"
)

// ProviderFactory creates a Provider from the given configuration.
type ProviderFactory func(config providers.ProviderConfig) (Provider, error)

// ProviderRegistry manages dynamically registered AI provider factories.
// It is safe for concurrent use.
type ProviderRegistry struct {
	mu        sync.RWMutex
	factories map[string]ProviderFactory
}

// defaultRegistry is the global provider registry used by NewProvider.
var defaultRegistry = NewProviderRegistry()

// NewProviderRegistry creates a new empty ProviderRegistry.
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		factories: make(map[string]ProviderFactory),
	}
}

// Register adds a provider factory under the given name.
// It panics if a factory is already registered for that name, which indicates
// a duplicate registration at init time — a programming error.
func (r *ProviderRegistry) Register(name string, factory ProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		panic(fmt.Sprintf("ai: provider %q already registered", name))
	}
	r.factories[name] = factory
}

// Get returns the factory for the given provider name, or nil if not found.
func (r *ProviderRegistry) Get(name string) (ProviderFactory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	f, ok := r.factories[name]
	return f, ok
}

// List returns the names of all registered providers, sorted alphabetically.
func (r *ProviderRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisterProvider adds a provider factory to the default global registry.
func RegisterProvider(name string, factory ProviderFactory) {
	defaultRegistry.Register(name, factory)
}

// GetProviderFactory returns the factory for the given name from the default registry.
func GetProviderFactory(name string) (ProviderFactory, bool) {
	return defaultRegistry.Get(name)
}

// RegisteredProviders returns the names of all providers in the default registry.
func RegisteredProviders() []string {
	return defaultRegistry.List()
}
