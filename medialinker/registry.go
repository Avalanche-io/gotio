// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package medialinker

import (
	"sync"
)

// Registry holds registered media linkers.
type Registry struct {
	mu            sync.RWMutex
	linkers       map[string]MediaLinker
	defaultLinker string
}

// defaultRegistry is the global registry instance.
var defaultRegistry = &Registry{
	linkers: make(map[string]MediaLinker),
}

// Register adds a linker to the global registry.
func Register(linker MediaLinker) {
	defaultRegistry.Register(linker)
}

// Get retrieves a linker by name from the global registry.
func Get(name string) (MediaLinker, error) {
	return defaultRegistry.Get(name)
}

// SetDefault sets the default linker name in the global registry.
func SetDefault(name string) error {
	return defaultRegistry.SetDefault(name)
}

// Default returns the default linker from the global registry.
func Default() (MediaLinker, error) {
	return defaultRegistry.Default()
}

// Available returns all registered linker names from the global registry.
func Available() []string {
	return defaultRegistry.Available()
}

// Register adds a linker to this registry.
func (r *Registry) Register(linker MediaLinker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.linkers[linker.Name()] = linker
}

// Unregister removes a linker from this registry.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.linkers, name)
	if r.defaultLinker == name {
		r.defaultLinker = ""
	}
}

// Get retrieves a linker by name.
func (r *Registry) Get(name string) (MediaLinker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	linker, ok := r.linkers[name]
	if !ok {
		return nil, &LinkerError{
			LinkerName: name,
			Message:    "linker not found",
		}
	}
	return linker, nil
}

// SetDefault sets the default linker name.
func (r *Registry) SetDefault(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.linkers[name]; !ok {
		return &LinkerError{
			LinkerName: name,
			Message:    "cannot set default: linker not found",
		}
	}
	r.defaultLinker = name
	return nil
}

// Default returns the default linker.
func (r *Registry) Default() (MediaLinker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.defaultLinker == "" {
		return nil, &LinkerError{
			Message: "no default linker set",
		}
	}

	linker, ok := r.linkers[r.defaultLinker]
	if !ok {
		return nil, &LinkerError{
			LinkerName: r.defaultLinker,
			Message:    "default linker not found",
		}
	}
	return linker, nil
}

// DefaultName returns the name of the default linker.
func (r *Registry) DefaultName() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultLinker
}

// Available returns all registered linker names.
func (r *Registry) Available() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.linkers))
	for name := range r.linkers {
		names = append(names, name)
	}
	return names
}

// Clear removes all linkers from this registry.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.linkers = make(map[string]MediaLinker)
	r.defaultLinker = ""
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		linkers: make(map[string]MediaLinker),
	}
}
