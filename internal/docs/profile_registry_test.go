package docs

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

type testProfile struct {
	name string
}

func (p testProfile) Name() string { return p.name }

func (p testProfile) Description() string { return "test profile " + p.name }

func (p testProfile) BuildPrompt(file string, content []byte) string {
	return "generate:" + file + ":" + string(content)
}

func (p testProfile) BuildUpdatePrompt(file string, content []byte) string {
	return "update:" + file + ":" + string(content)
}

func (p testProfile) FormatOutput(doc string) string { return doc }

func TestNewProfileRegistry_Empty(t *testing.T) {
	reg := NewProfileRegistry()
	if names := reg.List(); len(names) != 0 {
		t.Fatalf("expected empty registry, got %v", names)
	}
}

func TestProfileRegistry_RegisterAndGet(t *testing.T) {
	reg := NewProfileRegistry()
	profile := testProfile{name: "test-profile"}

	reg.Register(profile)

	got, ok := reg.Get("test-profile")
	if !ok {
		t.Fatal("expected to find registered profile")
	}
	if got.Name() != "test-profile" {
		t.Fatalf("got profile %q, want %q", got.Name(), "test-profile")
	}

	_, ok = reg.Get("missing")
	if ok {
		t.Fatal("expected not to find unregistered profile")
	}
}

func TestProfileRegistry_RegisterTrimsName(t *testing.T) {
	reg := NewProfileRegistry()
	reg.Register(testProfile{name: "  trimmed-profile  "})

	got, ok := reg.Get("trimmed-profile")
	if !ok {
		t.Fatal("expected trimmed profile name to be registered")
	}
	if got.Name() != "  trimmed-profile  " {
		t.Fatalf("stored profile name = %q", got.Name())
	}
}

func TestProfileRegistry_GetEmptyUsesDefaultName(t *testing.T) {
	reg := NewProfileRegistry()
	reg.Register(testProfile{name: DefaultProfileName})

	got, ok := reg.Get("")
	if !ok {
		t.Fatal("expected empty lookup to resolve default profile")
	}
	if got.Name() != DefaultProfileName {
		t.Fatalf("got profile %q, want %q", got.Name(), DefaultProfileName)
	}
}

func TestProfileRegistry_DuplicateRegisterPanics(t *testing.T) {
	reg := NewProfileRegistry()
	reg.Register(testProfile{name: "dup"})

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	reg.Register(testProfile{name: "dup"})
}

func TestProfileRegistry_DuplicateRegisterPanicsAfterTrim(t *testing.T) {
	reg := NewProfileRegistry()
	reg.Register(testProfile{name: "dup"})

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration after trimming")
		}
	}()
	reg.Register(testProfile{name: " dup "})
}

func TestProfileRegistry_EmptyNamePanics(t *testing.T) {
	reg := NewProfileRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on empty profile name")
		}
	}()
	reg.Register(testProfile{name: "  "})
}

func TestProfileRegistry_NilProfilePanics(t *testing.T) {
	reg := NewProfileRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil profile")
		}
	}()
	reg.Register(nil)
}

func TestProfileRegistry_ListSorted(t *testing.T) {
	reg := NewProfileRegistry()
	reg.Register(testProfile{name: "zebra"})
	reg.Register(testProfile{name: "alpha"})
	reg.Register(testProfile{name: "middle"})

	names := reg.List()
	expected := []string{"alpha", "middle", "zebra"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(names))
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("names[%d] = %q, want %q", i, names[i], name)
		}
	}
}

func TestProfileRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewProfileRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("profile-%d", i)
		wg.Add(3)

		go func() {
			defer wg.Done()
			reg.Register(testProfile{name: name})
		}()

		go func() {
			defer wg.Done()
			reg.List()
		}()

		go func() {
			defer wg.Done()
			reg.Get(name)
		}()
	}
	wg.Wait()
}

func TestProfileConfigTags(t *testing.T) {
	cfg := ProfileConfig{Name: "api-reference"}
	if cfg.Name != "api-reference" {
		t.Fatalf("ProfileConfig.Name = %q", cfg.Name)
	}
}

func TestDefaultProfileName(t *testing.T) {
	if DefaultProfileName != "software-documenter" {
		t.Fatalf("DefaultProfileName = %q", DefaultProfileName)
	}
}

func TestDefaultRegistryHelpers(t *testing.T) {
	withIsolatedDefaultProfileRegistry(t)

	RegisterProfile(testProfile{name: "helper-profile"})

	got, ok := GetProfile("helper-profile")
	if !ok {
		t.Fatal("expected helper-profile in default registry")
	}
	if got.Name() != "helper-profile" {
		t.Fatalf("got profile %q, want helper-profile", got.Name())
	}

	names := RegisteredProfiles()
	if len(names) != 1 || names[0] != "helper-profile" {
		t.Fatalf("RegisteredProfiles() = %v, want [helper-profile]", names)
	}
}

func TestDefaultProfile(t *testing.T) {
	withIsolatedDefaultProfileRegistry(t)
	if got := DefaultProfile(); got != nil {
		t.Fatalf("DefaultProfile() before registration = %v, want nil", got)
	}

	RegisterProfile(testProfile{name: DefaultProfileName})

	got := DefaultProfile()
	if got == nil {
		t.Fatal("expected DefaultProfile() after registration")
	}
	if got.Name() != DefaultProfileName {
		t.Fatalf("DefaultProfile().Name() = %q, want %q", got.Name(), DefaultProfileName)
	}
}

func TestResolveProfile(t *testing.T) {
	withIsolatedDefaultProfileRegistry(t)
	RegisterProfile(testProfile{name: DefaultProfileName})
	RegisterProfile(testProfile{name: "api-reference"})

	tests := []struct {
		name string
		want string
	}{
		{name: "", want: DefaultProfileName},
		{name: "  ", want: DefaultProfileName},
		{name: DefaultProfileName, want: DefaultProfileName},
		{name: "api-reference", want: "api-reference"},
		{name: " api-reference ", want: "api-reference"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveProfile(tt.name)
			if err != nil {
				t.Fatalf("ResolveProfile(%q) returned error: %v", tt.name, err)
			}
			if got.Name() != tt.want {
				t.Fatalf("ResolveProfile(%q).Name() = %q, want %q", tt.name, got.Name(), tt.want)
			}
		})
	}
}

func TestResolveProfile_Missing(t *testing.T) {
	withIsolatedDefaultProfileRegistry(t)
	RegisterProfile(testProfile{name: "alpha"})

	got, err := ResolveProfile("missing")
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
	if got != nil {
		t.Fatalf("expected nil profile on missing lookup, got %v", got)
	}
	if !strings.Contains(err.Error(), `unknown profile "missing"`) {
		t.Fatalf("error = %q, want unknown profile message", err.Error())
	}
	if !strings.Contains(err.Error(), "available: alpha") {
		t.Fatalf("error = %q, want available profile list", err.Error())
	}
}

func TestResolveProfile_MissingDefault(t *testing.T) {
	withIsolatedDefaultProfileRegistry(t)

	got, err := ResolveProfile("")
	if err == nil {
		t.Fatal("expected error when default profile is not registered")
	}
	if got != nil {
		t.Fatalf("expected nil profile on missing default, got %v", got)
	}
	if !strings.Contains(err.Error(), `unknown profile "software-documenter"`) {
		t.Fatalf("error = %q, want default profile name", err.Error())
	}
}

func withIsolatedDefaultProfileRegistry(t *testing.T) {
	t.Helper()
	original := defaultProfileRegistry
	defaultProfileRegistry = NewProfileRegistry()
	t.Cleanup(func() {
		defaultProfileRegistry = original
	})
}
