package vault

import (
	"errors"
	"sync"
)

var (
	ErrVaultNotFound  = errors.New("vault not found")
	ErrVaultExists    = errors.New("vault already registered")
	ErrNoDefaultVault = errors.New("no default vault configured")
)

// Registry holds named Vault instances.
type Registry struct {
	mu           sync.RWMutex
	vaults       map[string]*Vault
	defaultVault string
}

func NewRegistry() *Registry {
	return &Registry{
		vaults: make(map[string]*Vault),
	}
}

// Register adds a vault. If asDefault is true, it becomes the default.
// If no current default, it becomes the default.
func (r *Registry) Register(v *Vault, asDefault bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := v.Name()
	if _, exists := r.vaults[name]; exists {
		return ErrVaultExists
	}
	r.vaults[name] = v
	if asDefault || r.defaultVault == "" {
		r.defaultVault = name
	}
	return nil
}

// Get returns a vault by name.
func (r *Registry) Get(name string) (*Vault, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.vaults[name]
	if !ok {
		return nil, ErrVaultNotFound
	}
	return v, nil
}

// Default returns the default vault.
func (r *Registry) Default() (*Vault, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.vaults[r.defaultVault]
	if !ok {
		return nil, ErrNoDefaultVault
	}
	return v, nil
}

// DefaultName returns the default vault's name (for URN parsing).
func (r *Registry) DefaultName() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultVault
}
