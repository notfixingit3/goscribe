package docs

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ProfileRegistry manages dynamically registered documentation profiles.
// It is safe for concurrent use.
type ProfileRegistry struct {
	mu       sync.RWMutex
	profiles map[string]Profile
}

// defaultProfileRegistry is the global profile registry used by package helpers.
var defaultProfileRegistry = NewProfileRegistry()

// NewProfileRegistry creates a new empty ProfileRegistry.
func NewProfileRegistry() *ProfileRegistry {
	return &ProfileRegistry{
		profiles: make(map[string]Profile),
	}
}

// Register adds a profile under its trimmed name.
// It panics if the profile is nil, has an empty name, or duplicates an existing name.
func (r *ProfileRegistry) Register(profile Profile) {
	if profile == nil {
		panic("docs: cannot register nil profile")
	}

	name := strings.TrimSpace(profile.Name())
	if name == "" {
		panic("docs: profile name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.profiles[name]; exists {
		panic(fmt.Sprintf("docs: profile %q already registered", name))
	}
	r.profiles[name] = profile
}

// Get returns the profile for the given name, or nil if not found.
// An empty name resolves to DefaultProfileName.
func (r *ProfileRegistry) Get(name string) (Profile, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = DefaultProfileName
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	profile, ok := r.profiles[name]
	return profile, ok
}

// List returns the names of all registered profiles, sorted alphabetically.
func (r *ProfileRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.profiles))
	for name := range r.profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisterProfile adds a profile to the default global registry.
func RegisterProfile(profile Profile) {
	defaultProfileRegistry.Register(profile)
}

// GetProfile returns the profile for the given name from the default registry.
func GetProfile(name string) (Profile, bool) {
	return defaultProfileRegistry.Get(name)
}

// RegisteredProfiles returns the names of all profiles in the default registry.
func RegisteredProfiles() []string {
	return defaultProfileRegistry.List()
}

// DefaultProfile returns the registered default profile, or nil if it has not been registered yet.
func DefaultProfile() Profile {
	profile, _ := GetProfile(DefaultProfileName)
	return profile
}

// ResolveProfile returns the named profile, or the default profile when name is empty.
func ResolveProfile(name string) (Profile, error) {
	profile, ok := GetProfile(name)
	if ok {
		return profile, nil
	}

	requested := strings.TrimSpace(name)
	if requested == "" {
		requested = DefaultProfileName
	}

	return nil, fmt.Errorf("unknown profile %q (available: %s)", requested, strings.Join(RegisteredProfiles(), ", "))
}
